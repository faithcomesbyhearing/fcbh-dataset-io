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

	// Open database
	conn := db.NewDBAdapter(ctx, "/Users/jrstear/fcbh/arti/old/engniv2011/00005/database/engniv2011.db")

	// Create snippet extractor
	basePath := os.Getenv("FCBH_DATASET_FILES")
	if basePath == "" {
		fmt.Println("Error: FCBH_DATASET_FILES not set")
		os.Exit(1)
	}
	extractor := revise_audio.NewSnippetExtractor(ctx, conn, basePath)
	defer extractor.Cleanup()

	// Create audio stitcher
	stitcher := revise_audio.NewAudioStitcher(ctx)
	defer stitcher.Cleanup()

	fmt.Println("=== Testing Audio Stitching for Jude 1:1 ===\n")

	// Extract replacement snippets with margin for better context
	fmt.Println("1. Extracting replacement snippets with margin...")
	margin := 0.15 // 150ms margin before and after
	inFile, status := extractor.ExtractWord("JUD", 1, "2", 14, margin) // "in" from verse 2
	if status != nil {
		fmt.Printf("   Error extracting 'in': %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Extracted 'in' (with %.2fs margin): %s\n", margin, inFile)

	forFile, status := extractor.ExtractWord("JUD", 1, "3", 55, margin) // "for" from verse 3
	if status != nil {
		fmt.Printf("   Error extracting 'for': %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Extracted 'for' (with %.2fs margin): %s\n", margin, forFile)

	// Save snippets to permanent location for listening
	outputDir := filepath.Join(os.Getenv("HOME"), "tmp", "arti_revised_audio")
	os.MkdirAll(outputDir, 0755)
	
	inSnippet := filepath.Join(outputDir, "snippet_in_from_v2.wav")
	forSnippet := filepath.Join(outputDir, "snippet_for_from_v3.wav")
	
	data, err := os.ReadFile(inFile)
	if err == nil {
		err = os.WriteFile(inSnippet, data, 0644)
		if err == nil {
			fmt.Printf("   Saved 'in' snippet: %s\n", inSnippet)
		}
	}
	
	data, err = os.ReadFile(forFile)
	if err == nil {
		err = os.WriteFile(forSnippet, data, 0644)
		if err == nil {
			fmt.Printf("   Saved 'for' snippet: %s\n", forSnippet)
		}
	}

	// Get original chapter audio file - query database directly
	query := `SELECT audio_file FROM scripts WHERE book_id = 'JUD' AND chapter_num = 1 AND audio_file != '' LIMIT 1`
	var audioFileName string
	err = conn.DB.QueryRow(query).Scan(&audioFileName)
	if err != nil {
		fmt.Printf("   Error getting chapter audio: %v\n", err)
		os.Exit(1)
	}
	
	// Resolve full path (database may store just filename)
	originalAudioFile := filepath.Join(basePath, "engniv2011/n1da", audioFileName)
	if _, err := os.Stat(originalAudioFile); err != nil {
		// Try searching in subdirectories
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == audioFileName {
				originalAudioFile = path
				return filepath.SkipAll
			}
			return nil
		})
		if err != nil || originalAudioFile == filepath.Join(basePath, "engniv2011/n1da", audioFileName) {
			fmt.Printf("   Error: Could not find original audio file: %s\n", audioFileName)
			os.Exit(1)
		}
	}
	fmt.Printf("   Original audio: %s\n", originalAudioFile)

	fmt.Println()

	// Replace first "by" with "in" (10.12-10.24s)
	fmt.Println("2. Replacing 'by' (10.12-10.24s) with 'in'...")
	revised1, status := stitcher.ReplaceSegmentInChapter(
		originalAudioFile,
		inFile,
		10.12,
		10.24,
		0.0, // Zero cross-fade for testing
	)
	if status != nil {
		fmt.Printf("   Error: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Success: %s\n", revised1)

	fmt.Println()

	// Replace second "by" with "for" (11.68-11.92s) in the already-revised audio
	fmt.Println("3. Replacing 'by' (11.68-11.92s) with 'for' in revised audio...")
	revised2, status := stitcher.ReplaceSegmentInChapter(
		revised1,
		forFile,
		11.68,
		11.92,
		0.0, // Zero cross-fade for testing
	)
	if status != nil {
		fmt.Printf("   Error: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Success: %s\n", revised2)

	fmt.Println("\n=== Summary ===")
	fmt.Println("✅ Successfully created revised chapter audio with both replacements")
	fmt.Printf("   Final output: %s\n", revised2)
	
	// Copy to a permanent location for listening
	finalOutput := filepath.Join(outputDir, "JUD_01_revised.wav")
	
	// Copy file
	data, err = os.ReadFile(revised2)
	if err != nil {
		fmt.Printf("   Warning: Could not copy file: %v\n", err)
	} else {
		err = os.WriteFile(finalOutput, data, 0644)
		if err != nil {
			fmt.Printf("   Warning: Could not write to permanent location: %v\n", err)
		} else {
			fmt.Printf("\n📁 Saved to permanent location for listening:\n")
			fmt.Printf("   %s\n", finalOutput)
		}
	}
}

