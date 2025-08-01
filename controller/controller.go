package controller

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/bible_brain/timestamp/update"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/courier"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/encode"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/fetch"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/match/align"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/match/diff"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/mms"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/mms/adapter"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/mms/mms_align"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/mms/mms_asr"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/output"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/read"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/speech_to_text"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/timestamp"
	"os"
	"path/filepath"
	"time"
)

type OutputFiles struct {
	Directory string
	FilePaths []string
}

type Controller struct {
	ctx         context.Context
	yamlRequest []byte
	req         request.Request
	bucket      courier.Courier
	ident       db.Ident
	database    db.DBAdapter
	postFiles   *input.PostFiles
}

func NewController(ctx context.Context, yamlContent []byte) Controller {
	var c Controller
	c.ctx = ctx
	log.Info(ctx, "Request: ", string(yamlContent))
	c.yamlRequest = yamlContent
	c.bucket = courier.NewCourier(ctx, yamlContent)
	c.bucket.IsUnitTest = false // set to true when testing to make Courier work.
	return c
}

func (c *Controller) SetPostFiles(postFiles *input.PostFiles) {
	c.postFiles = postFiles
}

// Process is deprecated for production, but is a test only convenience method
func (c *Controller) Process() (string, *log.Status) {
	output, status := c.ProcessV2()
	if status != nil {
		return "", status
	}
	if len(output.FilePaths) > 0 {
		return output.FilePaths[0], status
	} else {
		return "NO OUTPUT", status
	}
}

// ProcessV2 is the production means to execute the controller
func (c *Controller) ProcessV2() (OutputFiles, *log.Status) {
	var start = time.Now()
	if c.postFiles != nil {
		defer c.postFiles.RemoveDir()
	}
	log.Debug(c.ctx)
	var status = c.processSteps()
	if status != nil {
		filename := c.outputStatus(*status)
		c.bucket.AddOutput(filename)
	}
	var output OutputFiles
	output.Directory = c.req.Output.Directory
	output.FilePaths = c.bucket.GetOutputPaths()
	log.Info(c.ctx, "Duration", time.Since(start))
	log.Debug(c.ctx)
	_ = c.bucket.PersistToBucket()                              // do not propagate error
	_ = c.bucket.Notification(c.req, status, time.Since(start)) // do not propagate error
	return output, status
}

func (c *Controller) processSteps() *log.Status {
	var filename string
	var status *log.Status
	// Decode YAML Request File
	log.Info(c.ctx, "Parse .yaml file.")
	reqDecoder := decode_yaml.NewRequestDecoder(c.ctx)
	c.req, status = reqDecoder.Process(c.yamlRequest)
	if status != nil {
		return status
	}
	c.ctx = context.WithValue(c.ctx, `request`, string(c.yamlRequest))
	// Open Database
	if c.req.Database.AWSS3 != "" {
		// Create an empty database in order to get the correct path.
		c.database, status = db.NewerDBAdapter(c.ctx, true, c.req.Username, c.req.DatasetName)
		if status != nil {
			return status
		}
		c.database.Close()
		// Import an existing database from S3.
		status = input.DownloadFile(c.ctx, c.req.Database.AWSS3, c.database.DatabasePath)
		if status != nil {
			return status
		}
	}
	c.database, status = db.NewerDBAdapter(c.ctx, c.req.IsNew, c.req.Username, c.req.DatasetName)
	if status != nil {
		return status
	}
	defer c.database.Close()
	c.bucket.AddDatabase(c.database)
	// Fetch Ident Data from Ident
	c.ident, status = c.database.SelectIdent()
	if status != nil {
		return status
	}
	// Update Ident Data from DBP
	c.ident, status = c.fetchData()
	if status != nil {
		if c.req.TextData.AnyBibleBrain() || c.req.AudioData.AnyBibleBrain() {
			return status
		}
	}
	// Collect Text Input
	var textFiles []input.InputFile
	if !c.req.TextData.NoText {
		log.Info(c.ctx, "Load text files.")
		textFiles, status = c.collectTextInput()
		if status != nil {
			return status
		}
	}
	// Read Text Data
	if !c.req.TextData.NoText {
		log.Info(c.ctx, "Read and parse text files.")
		status = c.readText(textFiles)
		if status != nil {
			return status
		}
	}
	// Collect Audio Input
	var audioFiles []input.InputFile
	if !c.req.AudioData.NoAudio {
		log.Info(c.ctx, "Load audio files.")
		audioFiles, status = c.collectAudioInput()
		if status != nil {
			return status
		}
	}
	// Update Ident Table
	status = input.UpdateIdent(c.database, &c.ident, textFiles, audioFiles)
	if status != nil {
		return status
	}
	// Timestamps
	if len(audioFiles) > 0 {
		log.Info(c.ctx, "Read or create audio timestamp data.")
		status = c.timestamps(audioFiles)
		if status != nil {
			return status
		}
	}
	// Train MMS Adapter
	if !c.req.Training.NoTraining {
		log.Info(c.ctx, "Train", c.ident.LanguageISO)
		trainer := adapter.NewTrainAdapter(c.ctx, c.database, c.ident.LanguageISO, c.req.Training.MMSAdapter)
		status = trainer.Train(audioFiles)
		if status != nil {
			return status
		}
	}
	// Copy for STT
	//if !c.req.TextData.NoText &&
	if !c.req.SpeechToText.NoSpeechToText {
		c.req.Compare.BaseDataset = c.database.Project
		c.req.AudioProof.BaseDataset = c.database.Project // ? should there be one BaseDataset ?
		// This makes a copy of database, and closes it.  Names the new database *_audio, and returns new
		c.database, status = c.database.CopyDatabase(`_audio`)
		if status != nil {
			return status
		}
		c.bucket.AddDatabase(c.database)
		status = c.database.UpdateEraseScriptText()
		if status != nil {
			return status
		}
	}
	// Speech to Text
	if !c.req.SpeechToText.NoSpeechToText {
		log.Info(c.ctx, "Perform speech to text.")
		status = c.speechToText(audioFiles)
		if status != nil {
			return status
		}
	}
	// Encode Audio
	if !c.req.AudioEncoding.NoEncoding {
		log.Info(c.ctx, "Perform audio encoding.")
		status = c.encodeAudio(audioFiles)
		if status != nil {
			return status
		}
	}
	// Encode Text
	if !c.req.TextEncoding.NoEncoding {
		log.Info(c.ctx, "Perform text encoding.")
		status = c.encodeText()
		if status != nil {
			return status
		}
	}
	// Audio Proofing
	if c.req.AudioProof.HTMLReport {
		log.Info(c.ctx, "Perform audio proof Report.")
		filename, status = c.audioProofing(audioFiles)
		if status != nil {
			return status
		}
		c.bucket.AddOutput(filename)
	}
	// Compare
	if c.req.Compare.HTMLReport {
		log.Info(c.ctx, "Perform text comparison.")
		filename, status = c.matchText()
		if status != nil {
			return status
		}
		c.bucket.AddOutput(filename)
	}
	// Update DBP Timestamps
	if len(c.req.UpdateDBP.Timestamps) > 0 {
		log.Info(c.ctx, "Update DBP timestamps.")
		upd := update.NewUpdateTimestamps(c.ctx, c.req, c.database)
		status = upd.Process()
		if status != nil {
			return status
		}
	}
	// Prepare output
	log.Info(c.ctx, "Generate output.")
	if c.req.Output.Sqlite {
		c.bucket.AddOutput(c.database.DatabasePath)
	}
	if c.req.Output.CSV || c.req.Output.JSON {
		status = c.output()
		// added to bucket in c.output()
	}
	return status
}

func (c *Controller) fetchData() (db.Ident, *log.Status) {
	var status *log.Status
	var info fetch.BibleInfoType
	client := fetch.NewAPIDBPClient(c.ctx, c.req.BibleId)
	info, status = client.BibleInfo()
	if status != nil {
		c.ident, status = client.UpdateIdent(c.ident, info, c.req)
		return c.ident, status
	}
	client.FindFilesets(&info, c.req.AudioData.BibleBrain, c.req.TextData.BibleBrain, c.req.Testament)
	download := fetch.NewAPIDownloadClient(c.ctx, c.req.BibleId, c.req.Testament)
	status = download.Download(info)
	if status != nil {
		return c.ident, status
	}
	c.ident, status = client.UpdateIdent(c.ident, info, c.req)
	return c.ident, status
}

func (c *Controller) collectTextInput() ([]input.InputFile, *log.Status) {
	var files []input.InputFile
	var status *log.Status
	var expectFiles = true
	bb := c.req.TextData.BibleBrain
	if bb.TextPlain || bb.TextPlainEdit || bb.TextUSXEdit {
		files, status = input.DBPDirectory(c.ctx, c.req.BibleId, c.ident.TextSource, c.ident.TextOTId,
			c.ident.TextNTId)
	} else if c.req.TextData.File != `` {
		files, status = input.FileInput(c.ctx, c.req.TextData.File)
	} else if c.req.TextData.AWSS3 != `` {
		files, status = input.AWSS3Input(c.ctx, c.req.TextData.AWSS3)
	} else if c.req.TextData.POST != `` && c.postFiles != nil {
		files, status = c.postFiles.PostInput("text")
	} else {
		expectFiles = false
	}
	if status != nil {
		return files, status
	}
	files, status = input.FillInputFile(c.ctx, c.req.Testament, files)
	if status != nil {
		return files, status
	}
	if expectFiles && len(files) == 0 {
		return files, log.ErrorNoErr(c.ctx, 400, `No text files found for`, c.ident.TextSource)
	}
	return files, nil
}

func (c *Controller) collectAudioInput() ([]input.InputFile, *log.Status) {
	var files []input.InputFile
	var status *log.Status
	var expectFiles = true
	bb := c.req.AudioData.BibleBrain
	if bb.MP3_64 || bb.MP3_16 || bb.OPUS {
		bibleId := c.req.BibleId
		files, status = input.DBPDirectory(c.ctx, bibleId, request.Audio, c.ident.AudioOTId,
			c.ident.AudioNTId)
	} else if c.req.AudioData.File != `` {
		files, status = input.FileInput(c.ctx, c.req.AudioData.File)
	} else if c.req.AudioData.AWSS3 != `` {
		files, status = input.AWSS3Input(c.ctx, c.req.AudioData.AWSS3)
	} else if c.req.AudioData.POST != `` && c.postFiles != nil {
		files, status = c.postFiles.PostInput("audio")
	} else {
		expectFiles = false
	}
	files, status = input.FillInputFile(c.ctx, c.req.Testament, files)
	if status != nil {
		return files, status
	}
	if expectFiles && len(files) == 0 {
		status = log.ErrorNoErr(c.ctx, 400, `No audio files found for`, c.ident.AudioNTId)
	}
	return files, status
}

func (c *Controller) readText(textFiles []input.InputFile) *log.Status {
	var status *log.Status
	if len(textFiles) == 0 {
		return status
	}
	if textFiles[0].MediaType == request.TextUSXEdit {
		reader := read.NewUSXParser(c.database)
		status = reader.ProcessFiles(textFiles)
		if status != nil {
			return status
		}
	} else if textFiles[0].MediaType == request.TextPlainEdit {
		reader := read.NewDBPTextEditReader(c.database, c.req)
		status = reader.Process()
		if status != nil {
			return status
		}
	} else if textFiles[0].MediaType == request.TextPlain {
		reader := read.NewDBPTextReader(c.database, c.req.Testament)
		status = reader.ProcessFiles(textFiles)
		if status != nil {
			return status
		}
	} else if textFiles[0].MediaType == request.TextScript {
		reader := read.NewScriptReader(c.database, c.req.Testament)
		status = reader.ProcessFiles(textFiles)
		if status != nil {
			return status
		}
	} else if textFiles[0].MediaType == request.TextCSV {
		reader := read.NewCSVReader(c.database)
		status = reader.ProcessFiles(textFiles)
		if status != nil {
			return status
		}
	} else {
		return status // This is not an error, it is nothing to do
	}
	if c.req.Detail.Words {
		words := read.NewWordParser(c.database)
		status = words.Parse()
	}
	return status
}

func (c *Controller) timestamps(audioFiles []input.InputFile) *log.Status {
	var status *log.Status
	if c.req.Timestamps.BibleBrain {
		//
		// Why isn't Bible Brain just processing input??
		//
		var filesetIds = []string{c.ident.AudioOTId, c.ident.AudioNTId}
		for _, filesetId := range filesetIds {
			if filesetId != `` {
				api := fetch.NewAPIDBPTimestamps(c.database, filesetId)
				// Load returns bool, which could be used to invoke aeneas
				_, status = api.LoadTimestamps(c.req.Testament)
				if status != nil {
					return status
				}
			}
		}
	} else if c.req.Timestamps.Aeneas {
		bibleId := c.req.BibleId
		aeneas := encode.NewAeneas(c.ctx, c.database, bibleId, c.ident.LanguageISO, c.req.Detail)
		status = aeneas.ProcessFiles(audioFiles)
	} else if c.req.Timestamps.TSBucket {
		var ts timestamp.TSBucket
		ts, status = timestamp.NewTSBucket(c.ctx, c.database)
		if status == nil {
			status = ts.ProcessFiles(audioFiles)
		}
	} else if c.req.Timestamps.MMSFAVerse {
		var ts mms.ForcedAlign
		ts = mms.NewForcedAlign(c.ctx, c.database, c.ident.LanguageISO, c.req.AltLanguage)
		status = ts.ProcessFiles(audioFiles)
	} else if c.req.Timestamps.MMSAlign {
		var ts mms_align.MMSAlign
		ts = mms_align.NewMMSAlign(c.ctx, c.database, c.ident.LanguageISO, c.req.AltLanguage)
		status = ts.ProcessFiles(audioFiles)
		if status != nil {
			return status
		}
	} else { // No timestamps, but has audio files
		status = timestamp.UpdateFilenames(c.ctx, c.database, audioFiles)
		if status != nil {
			return status
		}
	}
	return status
}

func (c *Controller) speechToText(audioFiles []input.InputFile) *log.Status {
	var status *log.Status
	bibleId := c.req.BibleId
	if c.req.SpeechToText.MMS {
		var asr mms_asr.MMSASR
		asr = mms_asr.NewMMSASR(c.ctx, c.database, c.ident.LanguageISO, c.req.AltLanguage, false)
		status = asr.ProcessFiles(audioFiles)
	} else if c.req.SpeechToText.MMSAdapter {
		var asr mms_asr.MMSASR
		asr = mms_asr.NewMMSASR(c.ctx, c.database, c.ident.LanguageISO, c.req.AltLanguage, true)
		status = asr.ProcessFiles(audioFiles)
	} else {
		var whisperModel = c.req.SpeechToText.Whisper.Model.String()
		if whisperModel != `` {
			var lang2 = c.req.AltLanguage
			var whisper = speech_to_text.NewWhisper(bibleId, c.database, whisperModel, lang2)
			status = whisper.ProcessFiles(audioFiles)
			if status != nil {
				return status
			}
			c.ident.TextSource = request.TextSTT
			if len(c.ident.AudioOTId) >= 10 {
				c.ident.TextOTId = c.ident.AudioOTId[:7] + `_TT`
			}
			if len(c.ident.AudioNTId) >= 10 {
				c.ident.TextNTId = c.ident.AudioNTId[:7] + `_TT`
			}
			_ = c.database.UpdateIdent(c.ident)
		}
	}
	return status
}

func (c *Controller) encodeAudio(audioFiles []input.InputFile) *log.Status {
	var status *log.Status
	bibleId := c.req.BibleId
	if c.req.AudioEncoding.MFCC {
		mfcc := encode.NewMFCC(c.ctx, c.database, bibleId, c.req.Detail, 7)
		status = mfcc.ProcessFiles(audioFiles)
		if status != nil {
			return status
		}
	}
	return status
}

func (c *Controller) encodeText() *log.Status {
	var status *log.Status
	if c.req.TextEncoding.FastText {
		fast := encode.NewFastText(c.ctx, c.database)
		status = fast.Process()
	}
	return status
}

func (c *Controller) audioProofing(audioFiles []input.InputFile) (string, *log.Status) {
	// Using audioFiles here should be temporary, once the timestamps are updated with duration
	// there should be no need for the audio files to be present.
	var filename string
	var status *log.Status
	if len(audioFiles) == 0 {
		return filename, log.ErrorNoErr(c.ctx, 400, "There are no audio files to AudioProof")
	}
	audioDir := audioFiles[0].Directory
	var textConn db.DBAdapter
	textConn, status = db.NewerDBAdapter(c.ctx, false, c.req.Username, c.req.AudioProof.BaseDataset)
	if status != nil {
		return filename, status
	}
	calc := align.NewAlignSilence(c.ctx, textConn, c.database) // c.database is ASR result
	faLines, filenameMap, status := calc.Process(audioDir)
	if status != nil {
		return filename, status
	}
	writer := align.NewAlignWriter(c.ctx, textConn)
	filename, status = writer.WriteReport(c.req.DatasetName, faLines, filenameMap)
	return filename, status
}

func (c *Controller) matchText() (string, *log.Status) {
	var records []diff.Pair
	var fileMap string
	var languageISO string
	var status *log.Status
	compare := diff.NewCompare(c.ctx, c.req.Username, c.req.Compare.BaseDataset, c.database, c.ident.LanguageISO, c.req.Testament, c.req.Compare.CompareSettings)
	records, fileMap, languageISO, status = compare.Process()
	if status != nil {
		return "", status
	}
	if c.req.Compare.GordonFilter > 0 {
		records, status = diff.GordonFilter(c.ctx, records, c.req.Username, c.req.Compare.BaseDataset, c.req.Compare.GordonFilter)
		if status != nil {
			return "", status
		}
	}
	tempFilePath := filepath.Join(os.TempDir(), c.database.Project+"_compare.json")
	c.bucket.AddJson(records, tempFilePath)
	writer := diff.NewHTMLWriter(c.ctx, c.database.Project)
	filename, status := writer.WriteReport(c.req.Compare.BaseDataset, records, languageISO, fileMap)
	return filename, status
}

func (c *Controller) output() *log.Status {
	var filename string
	var status *log.Status
	var out = output.NewOutput(c.ctx, c.database, c.req.DatasetName, false, false)
	var records []any
	var meta []output.Meta
	if c.req.Detail.Lines {
		records, meta = out.PrepareScripts()
	} else {
		records, meta = out.PrepareWords()
	}
	if c.req.Output.CSV {
		filename, status = out.WriteCSV(records, meta)
		if status != nil {
			return status
		}
		c.bucket.AddOutput(filename)
	}
	if c.req.Output.JSON {
		filename, status = out.WriteJSON(records, meta)
		if status != nil {
			return status
		}
		c.bucket.AddOutput(filename)
	}
	records = nil
	return status
}

func (c *Controller) outputStatus(status log.Status) string {
	var filename string
	var status2 *log.Status
	var out = output.NewOutput(c.ctx, db.DBAdapter{}, c.req.DatasetName, false, false)
	if c.req.Output.CSV {
		filename, status2 = out.CSVStatus(status, true)
	} else if c.req.Output.JSON {
		filename, status2 = out.JSONStatus(status, true)
	} else {
		filename = status.String()
	}
	if status2 != nil {
		filename = status2.String()
	}
	return filename
}
