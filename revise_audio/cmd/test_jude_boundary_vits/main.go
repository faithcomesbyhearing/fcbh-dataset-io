package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/revise_audio"
)

func main() {
	ctx := context.Background()

	// Ensure GOPROJ is set
	if os.Getenv("GOPROJ") == "" {
		wd, _ := os.Getwd()
		if filepath.Base(wd) == "test_jude_boundary_vits" {
			repoRoot := filepath.Join(wd, "..", "..", "..")
			os.Setenv("GOPROJ", repoRoot)
		} else if filepath.Base(filepath.Dir(wd)) == "cmd" {
			repoRoot := filepath.Join(wd, "..", "..")
			os.Setenv("GOPROJ", repoRoot)
		} else {
			os.Setenv("GOPROJ", wd)
		}
	}

	// Check environment variables
	soVitsRoot := os.Getenv("SO_VITS_SVC_ROOT")
	if soVitsRoot == "" {
		fmt.Println("Error: SO_VITS_SVC_ROOT not set")
		os.Exit(1)
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

	fmt.Println("=== Jude 1:1 Boundary-Based Revision Test ===\n")
	fmt.Println("Goal: Replace 'loved by' → 'loved in' and 'kept by' → 'kept for'")
	fmt.Println("Using: Boundary detection → MMS-TTS → VITS voice conversion → Stitching\n")

	// Open database
	conn := db.NewDBAdapter(ctx, dbPath)

	// Step 1: Get word data for boundary detection
	fmt.Println("1. Loading word data for boundary detection...")
	extractor := revise_audio.NewSnippetExtractor(ctx, conn, basePath)
	defer extractor.Cleanup()

	words, status := extractor.QueryWordsByVerse("JUD", 1, "1")
	if status != nil {
		fmt.Printf("Error querying words: %v\n", status)
		os.Exit(1)
	}

	if len(words) == 0 {
		fmt.Println("Error: No words found for Jude 1:1")
		os.Exit(1)
	}

	fmt.Printf("   Loaded %d words\n", len(words))

	// Step 2: Detect boundaries
	fmt.Println("\n2. Detecting segment boundaries...")
	boundaryDetector := revise_audio.NewSegmentBoundaryDetector(ctx)
	defer boundaryDetector.Cleanup()

	// Convert words to boundary detector format
	wordData := make([]struct {
		BeginTS float64
		EndTS   float64
		FAScore float64
	}, len(words))
	for i, w := range words {
		wordData[i] = struct {
			BeginTS float64
			EndTS   float64
			FAScore float64
		}{
			BeginTS: w.BeginTS,
			EndTS:   w.EndTS,
			FAScore: w.FAScore,
		}
	}

	boundaries := boundaryDetector.DetectBoundariesFromDB(wordData, 0.8)
	fmt.Printf("   Found %d boundaries\n", len(boundaries))

	// Find "loved" and "kept" word positions
	var lovedWord, keptWord *db.Audio
	for i := range words {
		if words[i].Text == "loved" {
			lovedWord = &words[i]
		}
		if words[i].Text == "kept" {
			keptWord = &words[i]
		}
	}

	if lovedWord == nil || keptWord == nil {
		fmt.Println("Error: Could not find 'loved' or 'kept' words")
		os.Exit(1)
	}

	// Step 3: Find best boundaries for replacement segments
	// Segment 1: "loved by" → "loved in God the Father"
	// Target: Replace from "loved" to "Father"
	fmt.Println("\n3. Finding optimal boundaries for replacement segments...")

	// Find segment 1 boundaries: "loved" to "Father"
	// Look for "Father" word
	var fatherWord *db.Audio
	for i := range words {
		if words[i].Text == "Father" {
			fatherWord = &words[i]
			break
		}
	}
	if fatherWord == nil {
		fmt.Println("Error: Could not find 'Father' word")
		os.Exit(1)
	}

	segment1Start, segment1End, status := boundaryDetector.FindBestBoundaries(
		boundaries,
		lovedWord.BeginTS,
		fatherWord.EndTS,
	)
	if status != nil {
		fmt.Printf("Error finding boundaries for segment 1: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Segment 1 (loved → Father): %.3fs - %.3fs\n", segment1Start, segment1End)

	// Find segment 2 boundaries: "kept" to "Christ" (end of verse)
	// Look for last "Christ" word
	var lastChristWord *db.Audio
	for i := len(words) - 1; i >= 0; i-- {
		if words[i].Text == "Christ" {
			lastChristWord = &words[i]
			break
		}
	}
	if lastChristWord == nil {
		fmt.Println("Error: Could not find 'Christ' word")
		os.Exit(1)
	}

	segment2Start, segment2End, status := boundaryDetector.FindBestBoundaries(
		boundaries,
		keptWord.BeginTS,
		lastChristWord.EndTS,
	)
	if status != nil {
		fmt.Printf("Error finding boundaries for segment 2: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Segment 2 (kept → Christ): %.3fs - %.3fs\n", segment2Start, segment2End)

	// Step 4: Get original chapter audio
	fmt.Println("\n4. Loading original chapter audio...")
	query := `SELECT audio_file FROM scripts WHERE book_id = 'JUD' AND chapter_num = 1 AND audio_file != '' LIMIT 1`
	var audioFileName string
	err := conn.DB.QueryRow(query).Scan(&audioFileName)
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

	// Step 5: Generate TTS for revised segments
	fmt.Println("\n5. Generating TTS for revised segments...")
	ttsAdapter, status := revise_audio.NewMMSTTSAdapter(ctx, "eng")
	if status != nil {
		fmt.Printf("Error creating TTS adapter: %v\n", status)
		os.Exit(1)
	}
	defer ttsAdapter.Close()

	// Segment 1: "loved in God the Father"
	segment1Text := "loved in God the Father"
	tts1Path, status := ttsAdapter.GeneratePhrase(segment1Text)
	if status != nil {
		fmt.Printf("Error generating TTS for segment 1: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Segment 1 TTS: %s\n", tts1Path)

	// Segment 2: "kept for Jesus Christ"
	segment2Text := "kept for Jesus Christ"
	tts2Path, status := ttsAdapter.GeneratePhrase(segment2Text)
	if status != nil {
		fmt.Printf("Error generating TTS for segment 2: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Segment 2 TTS: %s\n", tts2Path)

	// Step 6: Voice convert using VITS
	fmt.Println("\n6. Voice converting with So-VITS-SVC...")

	// Find latest checkpoint
	modelDir := filepath.Join(soVitsRoot, "logs", "44k")
	checkpointArg := "12000" // Latest from earlier check
	if len(os.Args) > 2 {
		checkpointArg = os.Args[2]
	}
	modelPath := filepath.Join(modelDir, fmt.Sprintf("G_%s.pth", checkpointArg))
	if _, err := os.Stat(modelPath); err != nil {
		fmt.Printf("Error: Model checkpoint not found: %s\n", modelPath)
		os.Exit(1)
	}
	fmt.Printf("   Using checkpoint: %s\n", filepath.Base(modelPath))

	configPath := filepath.Join(soVitsRoot, "configs", "config.json")
	if _, err := os.Stat(configPath); err != nil {
		configPath = filepath.Join(modelDir, "config.json")
	}
	if _, err := os.Stat(configPath); err != nil {
		fmt.Printf("Error: Config file not found: %s\n", configPath)
		os.Exit(1)
	}

	vcConfig := revise_audio.VoiceConversionConfig{
		F0Method: "rmvpe",
		Device:   "auto",
	}

	vitsAdapter := revise_audio.NewVITSAdapter(ctx, vcConfig)
	defer vitsAdapter.Close()

	speakerName := "jude_narrator"

	// Convert segment 1
	fmt.Println("   Converting segment 1...")
	vc1Path, status := vitsAdapter.ConvertVoice(tts1Path, modelPath, configPath, speakerName)
	if status != nil {
		fmt.Printf("Error converting segment 1: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Segment 1 converted: %s\n", vc1Path)

	// Convert segment 2
	fmt.Println("   Converting segment 2...")
	vc2Path, status := vitsAdapter.ConvertVoice(tts2Path, modelPath, configPath, speakerName)
	if status != nil {
		fmt.Printf("Error converting segment 2: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Segment 2 converted: %s\n", vc2Path)

	// Step 7: Stitch segments into chapter
	fmt.Println("\n7. Stitching revised segments into chapter...")
	stitcher := revise_audio.NewAudioStitcher(ctx)
	defer stitcher.Cleanup()

	// First replacement: segment 1
	fmt.Println("   Replacing segment 1...")
	revised1, status := stitcher.ReplaceSegmentInChapter(
		originalAudioFile,
		vc1Path,
		segment1Start,
		segment1End,
		0.05, // 50ms cross-fade
	)
	if status != nil {
		fmt.Printf("Error replacing segment 1: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   After segment 1 replacement: %s\n", revised1)

	// Second replacement: segment 2 (in the already-revised audio)
	fmt.Println("   Replacing segment 2...")
	revised2, status := stitcher.ReplaceSegmentInChapter(
		revised1,
		vc2Path,
		segment2Start,
		segment2End,
		0.05, // 50ms cross-fade
	)
	if status != nil {
		fmt.Printf("Error replacing segment 2: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Final revised audio: %s\n", revised2)

	// Step 8: Save final result
	outputDir := filepath.Join(os.Getenv("HOME"), "tmp", "arti_revised_audio")
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, "JUD_01_v1_boundary_vits.wav")

	data, err := os.ReadFile(revised2)
	if err != nil {
		fmt.Printf("Error reading final audio: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		fmt.Printf("Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Complete! Output saved to: %s\n", outputPath)
	fmt.Println("\nSummary:")
	fmt.Printf("  - Segment 1: %.3fs - %.3fs → '%s'\n", segment1Start, segment1End, segment1Text)
	fmt.Printf("  - Segment 2: %.3fs - %.3fs → '%s'\n", segment2Start, segment2End, segment2Text)
	fmt.Printf("  - VITS checkpoint: %s\n", filepath.Base(modelPath))
}

