package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func main() {
	var (
		audioFile   = flag.String("audio", "", "Path to input MP3 file (required)")
		timestampCSV = flag.String("timestamps", "", "Path to CSV file with timestamps (required)")
		outputDir   = flag.String("output", "", "Output directory for verse segments (default: same as audio file)")
		bookId      = flag.String("book", "", "Book ID (e.g., MAT, MRK) - required if not in CSV")
		chapterNum  = flag.Int("chapter", 0, "Chapter number - required if not in CSV")
	)
	flag.Parse()

	if *audioFile == "" || *timestampCSV == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -audio <mp3_file> -timestamps <csv_file> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCSV Format:\n")
		fmt.Fprintf(os.Stderr, "  The CSV file should have a header row with columns:\n")
		fmt.Fprintf(os.Stderr, "  - verse_str (required): Verse identifier (e.g., '1', '2', '1-2')\n")
		fmt.Fprintf(os.Stderr, "  - script_begin_ts or begin_ts (required): Start timestamp in seconds\n")
		fmt.Fprintf(os.Stderr, "  - script_end_ts or end_ts (required): End timestamp in seconds\n")
		fmt.Fprintf(os.Stderr, "  - book_id (optional): Book ID, or use -book flag\n")
		fmt.Fprintf(os.Stderr, "  - chapter_num (optional): Chapter number, or use -chapter flag\n")
		os.Exit(1)
	}

	// Validate input files exist
	if _, err := os.Stat(*audioFile); os.IsNotExist(err) {
		log.Fatalf("Audio file not found: %s", *audioFile)
	}
	if _, err := os.Stat(*timestampCSV); os.IsNotExist(err) {
		log.Fatalf("Timestamp CSV file not found: %s", *timestampCSV)
	}

	// Set output directory
	if *outputDir == "" {
		*outputDir = filepath.Dir(*audioFile)
	}
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Read timestamps from CSV
	timestamps, err := readTimestampsFromCSV(*timestampCSV, *bookId, *chapterNum)
	if err != nil {
		log.Fatalf("Failed to read timestamps: %v", err)
	}

	if len(timestamps) == 0 {
		log.Fatalf("No valid timestamps found in CSV file")
	}

	fmt.Printf("Found %d verse timestamps\n", len(timestamps))
	fmt.Printf("Input audio: %s\n", *audioFile)
	fmt.Printf("Output directory: %s\n", *outputDir)

	// Create temp directory for processing
	tempDir, err := os.MkdirTemp("", "chop_audio_*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Chop audio by timestamps - process each segment individually
	fileExt := filepath.Ext(*audioFile)
	var outputFiles []string
	for i, ts := range timestamps {
		if ts.BeginTS == 0.0 && ts.EndTS == 0.0 {
			continue
		}

		// Generate output filename in Bible Brain format
		outputFilename, err := generateBibleBrainFilename(*audioFile, ts.BookId, ts.ChapterNum, ts.VerseStr, fileExt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to generate filename for verse %s: %v\n", ts.VerseStr, err)
			// Fallback to simple format
			beginTS := fmt.Sprintf("%.3f", ts.BeginTS)
			outputFilename = fmt.Sprintf("verse_%s_%d_%s_%s%s",
				ts.BookId, ts.ChapterNum, ts.VerseStr, beginTS, fileExt)
		}
		outputPath := filepath.Join(*outputDir, outputFilename)

		// Chop one segment at a time, preserving the input format
		tempFile := filepath.Join(tempDir, fmt.Sprintf("%d%s", time.Now().UnixNano(), fileExt))
		err = ffmpeg.Input(*audioFile).Output(tempFile, ffmpeg.KwArgs{
			"codec:a": "copy",
			"c":       "copy",
			"y":       "",
			"ss":      ts.BeginTS,
			"to":      ts.EndTS,
		}).Silent(true).OverWriteOutput().Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to chop verse %s: %v\n", ts.VerseStr, err)
			continue
		}

		// Move file from temp to output directory
		if err := os.Rename(tempFile, outputPath); err != nil {
			// If rename fails (cross-device), copy instead
			data, err := os.ReadFile(tempFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to read %s: %v\n", tempFile, err)
				continue
			}
			if err := os.WriteFile(outputPath, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to write %s: %v\n", outputPath, err)
				continue
			}
			os.Remove(tempFile) // Clean up temp file after copy
		}
		outputFiles = append(outputFiles, outputPath)

		// Progress indicator
		if (i+1)%10 == 0 {
			fmt.Printf("Processed %d/%d verses...\n", i+1, len(timestamps))
		}
	}

	fmt.Printf("\n✅ Successfully created %d verse segments:\n", len(outputFiles))
	for _, file := range outputFiles {
		fmt.Printf("  - %s\n", filepath.Base(file))
	}
}

func readTimestampsFromCSV(csvPath string, defaultBookId string, defaultChapter int) ([]db.Audio, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header row and one data row")
	}

	// Parse header to find column indices
	header := records[0]
	headerMap := make(map[string]int)
	for i, col := range header {
		headerMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Find required columns - support both arti format (script_begin_ts, script_end_ts) 
	// and simple format (begin_ts, end_ts)
	verseIdx, hasVerse := headerMap["verse_str"]
	beginIdx, hasBegin := headerMap["script_begin_ts"]
	if !hasBegin {
		beginIdx, hasBegin = headerMap["begin_ts"]
	}
	endIdx, hasEnd := headerMap["script_end_ts"]
	if !hasEnd {
		endIdx, hasEnd = headerMap["end_ts"]
	}
	bookIdx, hasBook := headerMap["book_id"]
	chapterIdx, hasChapter := headerMap["chapter_num"]

	if !hasVerse || !hasBegin || !hasEnd {
		return nil, fmt.Errorf("CSV must have columns: verse_str, and either (script_begin_ts, script_end_ts) or (begin_ts, end_ts)")
	}

	// Check if we have book/chapter in CSV or flags
	bookId := defaultBookId
	chapterNum := defaultChapter
	if !hasBook && bookId == "" {
		return nil, fmt.Errorf("book_id not found in CSV and -book flag not provided")
	}
	if !hasChapter && chapterNum == 0 {
		return nil, fmt.Errorf("chapter_num not found in CSV and -chapter flag not provided")
	}

	// If book/chapter were provided via flags, use those for filtering
	// Otherwise, we'll use values from the CSV (don't filter)
	filterBookId := ""
	filterChapterNum := 0
	if defaultBookId != "" {
		filterBookId = defaultBookId
	}
	if defaultChapter != 0 {
		filterChapterNum = defaultChapter
	}

	var timestamps []db.Audio
	var skippedCount int
	for i, record := range records[1:] {
		if len(record) <= verseIdx || len(record) <= beginIdx || len(record) <= endIdx {
			fmt.Fprintf(os.Stderr, "Warning: Skipping row %d (insufficient columns)\n", i+2)
			continue
		}

		verseStr := strings.TrimSpace(record[verseIdx])
		beginStr := strings.TrimSpace(record[beginIdx])
		endStr := strings.TrimSpace(record[endIdx])

		if verseStr == "" || beginStr == "" || endStr == "" {
			fmt.Fprintf(os.Stderr, "Warning: Skipping row %d (empty required fields)\n", i+2)
			continue
		}

		beginTS, err := strconv.ParseFloat(beginStr, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping row %d (invalid timestamp: %s)\n", i+2, beginStr)
			continue
		}

		endTS, err := strconv.ParseFloat(endStr, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping row %d (invalid timestamp: %s)\n", i+2, endStr)
			continue
		}

		// Use book/chapter from CSV if available, otherwise use defaults
		tsBookId := bookId
		tsChapterNum := chapterNum
		if hasBook && len(record) > bookIdx && strings.TrimSpace(record[bookIdx]) != "" {
			tsBookId = strings.TrimSpace(record[bookIdx])
		}
		if hasChapter && len(record) > chapterIdx && strings.TrimSpace(record[chapterIdx]) != "" {
			if ch, err := strconv.Atoi(strings.TrimSpace(record[chapterIdx])); err == nil {
				tsChapterNum = ch
			}
		}

		// Filter by book and chapter if flags were provided
		if filterBookId != "" && tsBookId != filterBookId {
			skippedCount++
			continue // Skip rows that don't match the specified book
		}
		if filterChapterNum != 0 && tsChapterNum != filterChapterNum {
			skippedCount++
			continue // Skip rows that don't match the specified chapter
		}

		ts := db.Audio{
			BookId:     tsBookId,
			ChapterNum: tsChapterNum,
			VerseStr:   verseStr,
			BeginTS:    beginTS,
			EndTS:      endTS,
		}
		timestamps = append(timestamps, ts)
	}

	if skippedCount > 0 {
		var filterParts []string
		if filterBookId != "" {
			filterParts = append(filterParts, fmt.Sprintf("book=%s", filterBookId))
		}
		if filterChapterNum != 0 {
			filterParts = append(filterParts, fmt.Sprintf("chapter=%d", filterChapterNum))
		}
		fmt.Fprintf(os.Stderr, "Filtered out %d rows that didn't match %s\n", 
			skippedCount, strings.Join(filterParts, " "))
	}

	return timestamps, nil
}

// generateBibleBrainFilename generates a filename according to Bible Brain V4 naming convention:
// {mediaid}_{A/B}{ordering}_{USFM book code}_{chapter start}[_{verse start}-{chapter stop}_{verse stop}].mp3
// Examples:
//   - ENGESVN2DA_B001_MAT_001.mp3 (full chapter)
//   - IRUNLCP1DA_B013_1TH_001_01-001_010.mp3 (partial chapter, verses 1-10)
func generateBibleBrainFilename(inputFile string, bookId string, chapterNum int, verseStr string, fileExt string) (string, error) {
	// Extract media ID from input filename
	// Pattern: look for something like ENGNIVN1DA, ENGESVN2DA, etc. (typically at the end before .mp3)
	baseName := filepath.Base(inputFile)
	mediaIdPattern := regexp.MustCompile(`([A-Z]{3}[A-Z0-9]{3,}N[12]DA)`)
	matches := mediaIdPattern.FindStringSubmatch(baseName)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not extract media ID from filename: %s", baseName)
	}
	mediaId := matches[1]

	// Get book sequence number
	bookSeq, ok := db.BookSeqMap[bookId]
	if !ok {
		return "", fmt.Errorf("unknown book ID: %s", bookId)
	}

	// Determine testament prefix (A for OT, B for NT)
	var prefix string
	var ordering int
	if bookSeq < 40 {
		// Old Testament
		prefix = "A"
		ordering = bookSeq
	} else if bookSeq < 68 {
		// New Testament
		prefix = "B"
		ordering = bookSeq - 40
	} else {
		// Apocrypha - use A prefix but different ordering?
		// For now, treat as OT-like
		prefix = "A"
		ordering = bookSeq - 67
	}

	// Format ordering as 3 digits
	abSeq := fmt.Sprintf("%s%03d", prefix, ordering)

	// Format chapter as 3 digits
	chapterStr := fmt.Sprintf("%03d", chapterNum)

	// Parse verse string to determine if it's a single verse or range
	// verseStr can be "1", "2", "1-2", "1-10", etc.
	verseParts := strings.Split(verseStr, "-")
	verseStart, err := strconv.Atoi(strings.TrimSpace(verseParts[0]))
	if err != nil {
		return "", fmt.Errorf("invalid verse string: %s", verseStr)
	}

	// Build filename
	var filename string
	if len(verseParts) == 1 {
		// Single verse - format: {mediaid}_{ABseq}_{book}_{chapter}_{verse}-{chapter}_{verse}.mp3
		verseStartStr := fmt.Sprintf("%02d", verseStart)
		filename = fmt.Sprintf("%s_%s_%s_%s_%s-%s_%s%s",
			mediaId, abSeq, bookId, chapterStr, verseStartStr, chapterStr, verseStartStr, fileExt)
	} else {
		// Verse range - format: {mediaid}_{ABseq}_{book}_{chapter}_{verseStart}-{chapter}_{verseEnd}.mp3
		verseEnd, err := strconv.Atoi(strings.TrimSpace(verseParts[1]))
		if err != nil {
			return "", fmt.Errorf("invalid verse end in range: %s", verseParts[1])
		}
		verseStartStr := fmt.Sprintf("%02d", verseStart)
		verseEndStr := fmt.Sprintf("%02d", verseEnd)
		filename = fmt.Sprintf("%s_%s_%s_%s_%s-%s_%s%s",
			mediaId, abSeq, bookId, chapterStr, verseStartStr, chapterStr, verseEndStr, fileExt)
	}

	return filename, nil
}

