package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/revise_audio"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
)

func main() {
	ctx := context.Background()

	// Ensure GOPROJ is set
	if os.Getenv("GOPROJ") == "" {
		wd, _ := os.Getwd()
		if filepath.Base(filepath.Dir(wd)) == "cmd" {
			repoRoot := filepath.Join(wd, "..", "..")
			os.Setenv("GOPROJ", repoRoot)
		} else {
			os.Setenv("GOPROJ", wd)
		}
	}

	basePath := os.Getenv("FCBH_DATASET_FILES")
	if basePath == "" {
		fmt.Println("Error: FCBH_DATASET_FILES not set")
		os.Exit(1)
	}

	dbPath := "~/data/jrstear/engniv2011.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}
	// Expand ~
	if dbPath[:2] == "~/" {
		dbPath = filepath.Join(os.Getenv("HOME"), dbPath[2:])
	}

	fmt.Println("=== Prosody Matching: TTS Segment ===\n")

	// Open database
	conn := db.NewDBAdapter(ctx, dbPath)

	// Get verse 1 timestamps
	fmt.Println("1. Getting verse 1 timestamps...")
	verseQuery := `SELECT script_begin_ts, script_end_ts FROM scripts 
		WHERE book_id = 'JUD' AND chapter_num = 1 AND verse_str = '1' LIMIT 1`
	var verseBeginTS, verseEndTS float64
	err := conn.DB.QueryRow(verseQuery).Scan(&verseBeginTS, &verseEndTS)
	if err != nil {
		fmt.Printf("Error getting verse timestamps: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   Verse 1: %.3fs - %.3fs\n", verseBeginTS, verseEndTS)

	// Find where "called" ends (boundary point)
	fmt.Println("\n2. Finding boundary after 'have been called'...")
	extractor := revise_audio.NewSnippetExtractor(ctx, conn, basePath)
	defer extractor.Cleanup()

	words, status := extractor.QueryWordsByVerse("JUD", 1, "1")
	if status != nil {
		fmt.Printf("Error querying words: %v\n", status)
		os.Exit(1)
	}

	// Find "called" word
	var calledWord *db.Audio
	for i := range words {
		if words[i].Text == "called" {
			calledWord = &words[i]
			break
		}
	}
	if calledWord == nil {
		fmt.Println("Error: Could not find 'called' word")
		os.Exit(1)
	}

	// Find "who" word (start of reference segment)
	var whoWord *db.Audio
	for i := range words {
		if words[i].WordSeq > calledWord.WordSeq && words[i].Text == "who" {
			whoWord = &words[i]
			break
		}
	}
	if whoWord == nil {
		fmt.Println("Error: Could not find 'who' word")
		os.Exit(1)
	}

	referenceStartTS := whoWord.BeginTS
	fmt.Printf("   Reference segment starts at 'who': %.3fs\n", referenceStartTS)
	fmt.Printf("   Reference segment ends at verse end: %.3fs\n", verseEndTS)

	// Get original chapter audio
	query := `SELECT audio_file FROM scripts WHERE book_id = 'JUD' AND chapter_num = 1 AND audio_file != '' LIMIT 1`
	var audioFileName string
	err = conn.DB.QueryRow(query).Scan(&audioFileName)
	if err != nil {
		fmt.Printf("Error getting chapter audio: %v\n", err)
		os.Exit(1)
	}

	originalAudioFile := filepath.Join(basePath, "engniv2011/n1da", audioFileName)
	if _, err := os.Stat(originalAudioFile); err != nil {
		filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == audioFileName {
				originalAudioFile = path
				return filepath.SkipAll
			}
			return nil
		})
	}
	fmt.Printf("   Original audio: %s\n", originalAudioFile)

	// Extract reference audio segment (original "who are loved..." part)
	fmt.Println("\n3. Extracting reference audio segment...")
	tempDir := os.TempDir()
	referenceAudioPath, status := ffmpeg.ChopOneSegment(ctx, tempDir, originalAudioFile, referenceStartTS, verseEndTS)
	if status != nil {
		fmt.Printf("Error extracting reference segment: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Reference audio: %s\n", referenceAudioPath)

	// Input: raw TTS segment
	ttsPath := filepath.Join(os.Getenv("HOME"), "tmp", "arti_revised_audio", "JUD_01_v1_raw_tts_segment.wav")
	if _, err := os.Stat(ttsPath); err != nil {
		fmt.Printf("Error: TTS file not found: %s\n", ttsPath)
		os.Exit(1)
	}
	fmt.Printf("\n4. TTS source: %s\n", ttsPath)

	// Apply prosody matching
	fmt.Println("\n5. Applying prosody matching...")
	prosodyConfig := revise_audio.ProsodyConfig{
		UseDSP:          true,
		F0Method:        "auto",
		PitchShiftRange: 2.0,
		TimeStretchRange: 1.2,
		SampleRate:      16000,
	}

	prosodyAdapter := revise_audio.NewDSPProsodyAdapter(ctx, prosodyConfig)
	defer prosodyAdapter.Close()

	matchedPath, status := prosodyAdapter.MatchProsody(ttsPath, referenceAudioPath)
	if status != nil {
		fmt.Printf("Error matching prosody: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Prosody-matched audio: %s\n", matchedPath)

	// Save to output directory
	outputDir := filepath.Join(os.Getenv("HOME"), "tmp", "arti_revised_audio")
	os.MkdirAll(outputDir, 0755)
	outputWav := filepath.Join(outputDir, "JUD_01_v1_tts_prosody_matched.wav")
	outputMp3 := filepath.Join(outputDir, "JUD_01_v1_tts_prosody_matched.mp3")

	// Copy WAV
	data, err := os.ReadFile(matchedPath)
	if err != nil {
		fmt.Printf("Error reading prosody-matched file: %v\n", err)
		os.Exit(1)
	}
	err = os.WriteFile(outputWav, data, 0644)
	if err != nil {
		fmt.Printf("Error writing WAV: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   WAV saved: %s\n", outputWav)

	// Convert to MP3
	fmt.Println("\n6. Converting to MP3...")
	cmd := exec.Command("ffmpeg", "-i", outputWav, "-codec:a", "libmp3lame", "-b:a", "192k", outputMp3, "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error converting to MP3: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Complete! Output saved to: %s\n", outputMp3)
}

