package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const PostAudioWhisperJson = `is_new: yes
dataset_name: PostAudioWhisperJson_{bibleId}
bible_id: {bibleId}
username: GaryNTest
email: gary@shortsands.com
output:
  json: yes
audio_data:
  post: {namev4}
speech_to_text:
  whisper:
    model:
      tiny: yes
`

/*
1. Have yaml as always
2. perform replace
2a. filename format in POST: {mediaId}_{A/Bseq}_{book}_{chapter}_{verse}-{chapter_end}_{verse_end}
2b. IRUNLCP1DA_B013_1TH_001_01-001_010.mp3
3. save it to a file.
4. execute a curl command using cmd.exec, run()
5. save file
6. display name
*/
func TestPostAudioWhisperJsonAPI(t *testing.T) {
	type try struct {
		bibleId  string
		filePath string
		namev4   string
		expected int
	}
	var a try
	a.bibleId = `ENGWEB`
	a.filePath = `ENGWEB/ENGWEBN2DA-mp3-64/B23___01_1John_______ENGWEBN2DA.mp3`
	a.namev4 = `ENGWEBN2DA_B23_1JN_001.mp3`
	destFile := CopyAudio(a.namev4, a.filePath, t)
	a.expected = 15
	var request = strings.Replace(PostAudioWhisperJson, `{bibleId}`, a.bibleId, 2)
	request = strings.Replace(request, `{namev4}`, destFile, 1)
	stdout, stderr := CurlExec(request, destFile, t)
	fmt.Println(`STDOUT`, stdout)
	fmt.Println(`STDERR`, stderr)
	count := countRecords(stdout)
	if count != a.expected {
		t.Error(`expected,`, a.expected, `found`, count)
	}
}

func CopyAudio(dest string, source string, t *testing.T) string {
	srcPath := filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), source)
	srcFile, err := os.Open(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	destPath := filepath.Join(os.Getenv(`FCBH_DATASET_TMP`), dest)
	destFile, err := os.Create(destPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		t.Fatal(err)
	}
	_ = srcFile.Close()
	_ = destFile.Close()
	return destPath
}

func CurlExec(requestYaml string, filePath string, t *testing.T) (string, string) {
	yamlFile, err := os.CreateTemp(os.Getenv(`FCBH_DATASET_TMP`), `request`+"_*.yaml")
	if err != nil {
		t.Error(err)
	}
	_, _ = yamlFile.WriteString(requestYaml)
	_ = yamlFile.Close()
	audioPart := "audio=@" + filePath + ";type=audio/mpeg"
	yamlPart := "yaml=@" + yamlFile.Name() + ";type=application/x-yaml"
	var cmd = exec.Command(`curl`, `-X`, `POST`, UPLOADHOST, `-F`, audioPart, `-F`, yamlPart,
		`-H`, `Accept: application/json`)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	if err != nil {
		t.Error(err.Error())
	}
	_ = os.Remove(yamlFile.Name())
	return stdoutBuf.String(), stderrBuf.String()
}

func countRecords(output string) int {
	var count int
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, `book_id`) {
			count++
		}
	}
	return count
}

//curl -X POST http://localhost:8080 \
//-F "audio=@audio.mp3;type=audio/mpeg" \
//-F "yaml=@request.yaml;type=application/x-yaml" \
//-H "Accept: application/json"
