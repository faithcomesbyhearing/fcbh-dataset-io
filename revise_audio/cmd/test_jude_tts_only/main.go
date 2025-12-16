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
		if filepath.Base(wd) == "test_jude_tts_only" {
			repoRoot := filepath.Join(wd, "..", "..", "..")
			os.Setenv("GOPROJ", repoRoot)
		} else if filepath.Base(filepath.Dir(wd)) == "cmd" {
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

	fmt.Println("=== Jude 1:1 Partial Replacement Test ===\n")
	fmt.Println("Keeping first half from original, replacing second half with TTS\n")

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

	// Boundary is after "called" ends
	boundaryTS := calledWord.EndTS
	
	// Find next word ("who") to calculate gap
	var nextWord *db.Audio
	for i := range words {
		if words[i].WordSeq > calledWord.WordSeq && words[i].Text == "who" {
			nextWord = &words[i]
			break
		}
	}
	
	var gapAtBoundary float64
	if nextWord != nil {
		gapAtBoundary = nextWord.BeginTS - calledWord.EndTS
		fmt.Printf("   Boundary after 'called': %.3fs\n", boundaryTS)
		fmt.Printf("   Gap to next word ('who'): %.3fs\n", gapAtBoundary)
	} else {
		fmt.Printf("   Boundary after 'called': %.3fs\n", boundaryTS)
		fmt.Printf("   Warning: Could not find next word to calculate gap\n")
	}

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

	// Text to replace: "who are loved in God the Father and kept for Jesus Christ:"
	replacementText := "who are loved in God the Father and kept for Jesus Christ:"
	fmt.Printf("\n3. Replacement text: %s\n", replacementText)

	// Generate TTS for replacement text only
	fmt.Println("\n4. Generating TTS for replacement segment...")
	ttsAdapter, status := revise_audio.NewMMSTTSAdapter(ctx, "eng")
	if status != nil {
		fmt.Printf("Error creating TTS adapter: %v\n", status)
		os.Exit(1)
	}
	defer ttsAdapter.Close()

	ttsPath, status := ttsAdapter.GeneratePhrase(replacementText)
	if status != nil {
		fmt.Printf("Error generating TTS: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   TTS generated: %s\n", ttsPath)

	// Save a copy of the raw TTS segment for inspection
	outputDir := filepath.Join(os.Getenv("HOME"), "tmp", "arti_revised_audio")
	os.MkdirAll(outputDir, 0755)
	rawTtsWav := filepath.Join(outputDir, "JUD_01_v1_raw_tts_segment.wav")
	rawTtsMp3 := filepath.Join(outputDir, "JUD_01_v1_raw_tts_segment.mp3")
	
	// Copy WAV
	ttsData, err := os.ReadFile(ttsPath)
	if err != nil {
		fmt.Printf("Warning: Could not copy raw TTS file: %v\n", err)
	} else {
		err = os.WriteFile(rawTtsWav, ttsData, 0644)
		if err != nil {
			fmt.Printf("Warning: Could not save raw TTS file: %v\n", err)
		} else {
			fmt.Printf("   Raw TTS segment saved: %s\n", rawTtsWav)
			// Convert to MP3
			ttsCmd := exec.Command("ffmpeg", "-i", rawTtsWav, "-codec:a", "libmp3lame", "-b:a", "192k", rawTtsMp3, "-y")
			ttsCmd.Stdout = os.Stdout
			ttsCmd.Stderr = os.Stderr
			if err := ttsCmd.Run(); err != nil {
				fmt.Printf("Warning: Could not convert raw TTS to MP3: %v\n", err)
			} else {
				fmt.Printf("   Raw TTS segment (MP3): %s\n", rawTtsMp3)
			}
		}
	}

	// Stitch TTS into chapter with gap matching approach
	// Step-by-step: 1) Generate TTS (done), 2) Measure leading silence, 3) Match original gap
	fmt.Println("\n5. Stitching TTS into chapter with gap matching...")
	stitcher := revise_audio.NewAudioStitcher(ctx)
	defer stitcher.Cleanup()

	// Use the new gap matching approach:
	// - Measures leading silence in TTS
	// - Uses calculated gapAtBoundary (from word timestamps) for original gap
	// - Places TTS appropriately (with or without cross-fade based on TTS silence)
	revisedAudio, status := stitcher.ReplaceSegmentInChapterWithGapMatching(
		originalAudioFile,
		ttsPath,
		boundaryTS,     // Boundary where replacement starts (end of "called")
		verseEndTS,     // End of segment to replace
		0.05,           // Cross-fade duration (used only if TTS has leading silence)
		gapAtBoundary,  // Original gap at boundary (calculated from word timestamps)
	)
	if status != nil {
		fmt.Printf("Error stitching: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Revised audio: %s\n", revisedAudio)

	// Extract just verse 1 for output
	fmt.Println("\n6. Extracting verse 1 segment from revised audio...")
	tempDir := os.TempDir()
	// Include trailing gap
	trailingGap := 0.271
	verseEndWithGap := verseEndTS + trailingGap
	verseSegment, status := ffmpeg.ChopOneSegment(ctx, tempDir, revisedAudio, verseBeginTS, verseEndWithGap)
	if status != nil {
		fmt.Printf("Error extracting verse: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Extracted verse 1: %.3fs - %.3fs\n", verseBeginTS, verseEndWithGap)

	// Save as MP3 (outputDir already created earlier)
	outputWav := filepath.Join(outputDir, "JUD_01_v1_tts_only.wav")
	outputMp3 := filepath.Join(outputDir, "JUD_01_v1_tts_only.mp3")

	// Copy WAV
	data, err := os.ReadFile(verseSegment)
	if err != nil {
		fmt.Printf("Error reading verse segment: %v\n", err)
		os.Exit(1)
	}
	err = os.WriteFile(outputWav, data, 0644)
	if err != nil {
		fmt.Printf("Error writing WAV: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   WAV saved: %s\n", outputWav)

	// Convert to MP3
	fmt.Println("\n7. Converting to MP3...")
	cmd := exec.Command("ffmpeg", "-i", outputWav, "-codec:a", "libmp3lame", "-b:a", "192k", outputMp3, "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error converting to MP3: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Complete! Output saved to: %s\n", outputMp3)
	fmt.Println("\nSummary:")
	fmt.Printf("  - Original segment: %.3fs - %.3fs (up to 'called')\n", verseBeginTS, boundaryTS)
	fmt.Printf("  - TTS replacement: %.3fs - %.3fs ('who are loved...')\n", boundaryTS, verseEndTS)
	fmt.Printf("  - Gap preservation: Enabled (matching original gap at boundary)\n")
}
