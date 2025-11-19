package main

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/lang_tree/search"
	_ "github.com/go-sql-driver/mysql"
)

type BibleInfo struct {
	BibleId      string
	LanguageISO  string
	FilesetId    string
	AudioFileset string
	TextFileset  string
}

const (
	audioAccessGroupID = 1013
	textAccessGroupID  = 1011
	dbpAssetID         = "dbp-prod"
)

type YAMLGenerator struct {
	config     Config
	db         *sql.DB
	template   string
	mmssupport map[string]bool
	ctx        context.Context
	langTree   search.LanguageTree
}

type queryOptions struct {
	specificBible bool
	duplicate     bool
}

func NewYAMLGenerator(config Config) (*YAMLGenerator, error) {
	// Connect to MySQL database
	db, err := sql.Open("mysql", getDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Create context and load language tree
	ctx := context.Background()
	langTree := search.NewLanguageTree(ctx)
	err = langTree.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load language tree: %v", err)
	}

	// Load MMS supported languages using the real language tree
	mmsSupport, err := loadMMSSupportFromTree(langTree)
	if err != nil {
		return nil, fmt.Errorf("failed to load MMS support: %v", err)
	}

	// Load template
	template, err := loadTemplate(config.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %v", err)
	}

	return &YAMLGenerator{
		config:     config,
		db:         db,
		template:   template,
		mmssupport: mmsSupport,
		ctx:        ctx,
		langTree:   langTree,
	}, nil
}

func (g *YAMLGenerator) Generate() error {
	defer g.db.Close()

	// Find matching languages
	bibles, err := g.findMatchingLanguages()
	if err != nil {
		return fmt.Errorf("failed to find matching languages: %v", err)
	}

	if len(bibles) == 0 {
		fmt.Printf("No matching languages found for testament=%s, text=%s\n", g.config.Testament, g.config.TextType)
		return nil
	}

	fmt.Printf("Found %d matching languages\n", len(bibles))

	// Generate YAML for each Bible
	for _, bible := range bibles {
		if err := g.generateYAML(bible); err != nil {
			fmt.Printf("Warning: Failed to generate YAML for %s: %v\n", bible.BibleId, err)
			continue
		}

		if g.config.Verbose {
			fmt.Printf("Generated: %s.yaml\n", bible.FilesetId)
		}
	}

	return nil
}

func (g *YAMLGenerator) findMatchingLanguages() ([]BibleInfo, error) {
	opts := queryOptions{
		specificBible: g.config.BibleId != "",
		duplicate:     g.config.Duplicate,
	}

	var args []interface{}

	textPattern := ""
	textExclusion := ""
	includeTextExclusion := false

	if !opts.duplicate {
		textPattern = g.getTextPattern()
		textExclusion = g.getTextExclusionPattern()
		includeTextExclusion = textExclusion != ""
	}

	query := g.buildQuery(opts, includeTextExclusion)

	if opts.specificBible {
		args = append(args, g.config.BibleId)
	}

	args = append(args, g.getAudioPattern())

	if !opts.duplicate {
		args = append(args, textPattern)
		if includeTextExclusion {
			args = append(args, textExclusion)
		}
	}

	args = append(args, g.getSAPattern())

	if g.config.Verbose {
		fmt.Println("Executing query:")
		fmt.Println(formatQueryWithArgs(query, args))
	}

	rows, err := g.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var bibles []BibleInfo
	for rows.Next() {
		var bible BibleInfo
		err := rows.Scan(&bible.BibleId, &bible.LanguageISO, &bible.FilesetId, &bible.AudioFileset, &bible.TextFileset)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		// Check MMS support using language tree (skip in duplicate mode)
		if !g.config.Duplicate && !g.isMMSSupported(bible.LanguageISO) {
			if g.config.Verbose {
				fmt.Printf("Skipping %s: language %s not supported by MMS\n", bible.BibleId, bible.LanguageISO)
			}
			continue
		}

		bibles = append(bibles, bible)
	}

	return bibles, nil
}

func (g *YAMLGenerator) buildQuery(opts queryOptions, includeTextExclusion bool) string {
	var builder strings.Builder

	builder.WriteString("SELECT \n")
	builder.WriteString("\tb.id as bible_id,\n")
	builder.WriteString("\tl.iso as language_iso,\n")
	builder.WriteString("\tfs.id as fileset_id,\n")
	builder.WriteString("\tfs.id as audio_fileset,\n")
	if opts.duplicate {
		builder.WriteString("\t'' as text_fileset\n")
	} else {
		builder.WriteString("\tMIN(text_fs.id) as text_fileset\n")
	}
	builder.WriteString("FROM bibles b\n")
	builder.WriteString("JOIN languages l ON b.language_id = l.id\n")
	builder.WriteString("JOIN bible_fileset_connections bfc ON b.id = bfc.bible_id\n")
	builder.WriteString("JOIN bible_filesets fs ON bfc.hash_id = fs.hash_id\n")
	fmt.Fprintf(&builder, "JOIN access_group_filesets audio_agf ON audio_agf.hash_id = fs.hash_id AND audio_agf.access_group_id = %d\n", audioAccessGroupID)
	if !opts.duplicate {
		builder.WriteString("JOIN bible_fileset_connections text_bfc ON b.id = text_bfc.bible_id\n")
		builder.WriteString("JOIN bible_filesets text_fs ON text_bfc.hash_id = text_fs.hash_id\n")
		fmt.Fprintf(&builder, "JOIN access_group_filesets text_agf ON text_agf.hash_id = text_fs.hash_id AND text_agf.access_group_id = %d\n", textAccessGroupID)
	}

	builder.WriteString("WHERE ")
	if opts.specificBible {
		builder.WriteString("b.id = ?\n")
		builder.WriteString("AND ")
	}
	builder.WriteString("fs.id LIKE ?\n")
	builder.WriteString("AND fs.content_loaded = 1\n")
	fmt.Fprintf(&builder, "AND fs.asset_id = '%s'\n", dbpAssetID)

	if !opts.duplicate {
		builder.WriteString("AND text_fs.id LIKE ?\n")
		builder.WriteString("AND text_fs.content_loaded = 1\n")
		fmt.Fprintf(&builder, "AND text_fs.asset_id = '%s'\n", dbpAssetID)
		if includeTextExclusion {
			builder.WriteString("AND NOT EXISTS (\n")
			builder.WriteString("\tSELECT 1 FROM bible_filesets excl_fs \n")
			builder.WriteString("\tJOIN bible_fileset_connections excl_bfc ON excl_fs.hash_id = excl_bfc.hash_id\n")
			builder.WriteString("\tWHERE excl_bfc.bible_id = b.id \n")
			builder.WriteString("\tAND excl_fs.id LIKE ?\n")
			builder.WriteString(")\n")
		}
	}

	builder.WriteString("AND NOT EXISTS (\n")
	builder.WriteString("\tSELECT 1 FROM bible_filesets sa_fs \n")
	builder.WriteString("\tJOIN bible_fileset_connections sa_bfc ON sa_fs.hash_id = sa_bfc.hash_id\n")
	builder.WriteString("\tWHERE sa_bfc.bible_id = b.id \n")
	builder.WriteString("\tAND sa_fs.id LIKE ?\n")
	builder.WriteString(")\n")

	if strings.HasPrefix(g.config.Testament, "n2") {
		builder.WriteString("AND NOT EXISTS (\n")
		builder.WriteString("\tSELECT 1 FROM bible_filesets n1sa_fs \n")
		builder.WriteString("\tJOIN bible_fileset_connections n1sa_bfc ON n1sa_fs.hash_id = n1sa_bfc.hash_id\n")
		builder.WriteString("\tWHERE n1sa_bfc.bible_id = b.id \n")
		builder.WriteString("\tAND n1sa_fs.id LIKE '%N1SA%'\n")
		builder.WriteString(")\n")
	}

	builder.WriteString("GROUP BY b.id, l.iso, fs.id ORDER BY b.id, fs.id")

	return builder.String()
}

func formatQueryWithArgs(query string, args []interface{}) string {
	if len(args) == 0 {
		return query
	}

	var builder strings.Builder
	argIdx := 0

	for i := 0; i < len(query); i++ {
		if query[i] == '?' && argIdx < len(args) {
			builder.WriteString(formatQueryArg(args[argIdx]))
			argIdx++
			continue
		}
		builder.WriteByte(query[i])
	}

	return builder.String()
}

func formatQueryArg(arg interface{}) string {
	if arg == nil {
		return "NULL"
	}

	switch v := arg.(type) {
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case []byte:
		return "'" + strings.ReplaceAll(string(v), "'", "''") + "'"
	case fmt.Stringer:
		escaped := strings.ReplaceAll(v.String(), "'", "''")
		return "'" + escaped + "'"
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprint(v)
	}
}

func (g *YAMLGenerator) getAudioPattern() string {
	// Simple approach: match exact endings based on testament and stream type
	switch g.config.Testament {
	case "n1":
		if g.config.StreamType == "hls" {
			return "%N1DA" // Match ABPWBTN1DA
		} else { // dash
			return "%N1DA-opus16" // Match ABPWBTN1DA-opus16
		}
	case "n2":
		if g.config.StreamType == "hls" {
			return "%N2DA" // Match ABPWBTN2DA
		} else { // dash
			return "%N2DA-opus16" // Match ABPWBTN2DA-opus16
		}
	case "o1":
		if g.config.StreamType == "hls" {
			return "%O1DA" // Match ABPWBTO1DA
		} else { // dash
			return "%O1DA-opus16" // Match ABPWBTO1DA-opus16
		}
	case "o2":
		if g.config.StreamType == "hls" {
			return "%O2DA" // Match ABPWBTO2DA
		} else { // dash
			return "%O2DA-opus16" // Match ABPWBTO2DA-opus16
		}
	default:
		return "%"
	}
}

func (g *YAMLGenerator) getTextPattern() string {
	// Match testament scope with text fileset pattern
	var testamentPrefix string
	switch g.config.Testament {
	case "n1", "n2":
		testamentPrefix = "N" // New Testament
	case "o1", "o2":
		testamentPrefix = "O" // Old Testament
	default:
		testamentPrefix = ""
	}

	switch g.config.TextType {
	case "usx":
		if testamentPrefix != "" {
			return "%" + testamentPrefix + "_ET-usx" // Match N_ET-usx or O_ET-usx
		}
		return "%_ET-usx" // Fallback to any USX filesets
	case "plain":
		if testamentPrefix != "" {
			return "%" + testamentPrefix + "_ET" // Match N_ET or O_ET (excludes USX)
		}
		return "%_ET" // Fallback to any plain text filesets
	default:
		return "%"
	}
}

func (g *YAMLGenerator) getTextExclusionPattern() string {
	switch g.config.TextType {
	case "usx":
		return "" // No exclusions for USX
	case "plain":
		return "%_ET-usx" // Exclude USX filesets for plain text
	default:
		return ""
	}
}

func (g *YAMLGenerator) getSAPattern() string {
	switch g.config.Testament {
	case "n1":
		return "%N1SA%"
	case "n2":
		return "%N2SA%"
	case "o1":
		return "%O1SA%"
	case "o2":
		return "%O2SA%"
	default:
		return "%SA%"
	}
}

func (g *YAMLGenerator) generateYAML(bible BibleInfo) error {
	// Generate content based on the -only parameter
	var content string
	var filename string

	if g.config.Duplicate {
		if g.config.Only != "" {
			return fmt.Errorf("duplicate mode does not support -only flag")
		}
		content, skip, err := g.generateDuplicationYAML(bible)
		if err != nil {
			return err
		}
		if skip {
			if g.config.Verbose {
				fmt.Printf("Skipping %s: duplication validation failed\n", bible.FilesetId)
			}
			return nil
		}
		filename = filepath.Join(g.config.OutputDir, bible.FilesetId+".yaml")
		return os.WriteFile(filename, []byte(content), 0644)
	}

	switch g.config.Only {
	case "timings":
		content = g.generateTimingsOnlyYAML(bible)
		filename = filepath.Join(g.config.OutputDir, bible.FilesetId+"_timings.yaml")
	case "streams":
		content = g.generateStreamsOnlyYAML(bible)
		filename = filepath.Join(g.config.OutputDir, bible.FilesetId+"_streams.yaml")
	default:
		// Both (current behavior)
		content = g.generateBothYAML(bible)
		filename = filepath.Join(g.config.OutputDir, bible.FilesetId+".yaml")
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

func (g *YAMLGenerator) generateTimingsOnlyYAML(bible BibleInfo) string {
	content := g.template

	// Remove update_dbp stanza by finding and removing it
	lines := strings.Split(content, "\n")
	var filteredLines []string
	inUpdateDBP := false

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "update_dbp:") {
			inUpdateDBP = true
			continue
		}
		if inUpdateDBP && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "\t") {
			inUpdateDBP = false
		}
		if !inUpdateDBP {
			filteredLines = append(filteredLines, line)
		}
	}

	content = strings.Join(filteredLines, "\n")

	// Replace template placeholders
	content = strings.ReplaceAll(content, "{{DATASET_NAME}}", bible.FilesetId)
	content = strings.ReplaceAll(content, "{{BIBLE_ID}}", bible.BibleId)
	content = strings.ReplaceAll(content, "{{TIMESTAMPS_FILESET}}", bible.AudioFileset)
	content = strings.ReplaceAll(content, "{{SET_TYPE_CODE}}", g.getSetTypeCode())

	// Replace audio type code format
	audioTypeCode := g.getAudioTypeCode()
	content = strings.ReplaceAll(content, "{{AUDIO_TYPE_CODE}}", "set_type_code: "+audioTypeCode)

	// Change output to CSV only - remove JSON line completely
	content = strings.ReplaceAll(content, "  json: yes\n", "")
	content = strings.ReplaceAll(content, "  json: no\n", "")
	content = strings.ReplaceAll(content, "  csv: yes", "  csv: yes")

	// Note: mp3_64 is optional and not included to avoid issues with filesets missing codec/bitrate tags
	// The codec/bitrate check is bypassed when not specified, allowing filesets with missing tags to match

	// Ensure YAML ends with newline
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	return content
}

func (g *YAMLGenerator) generateStreamsOnlyYAML(bible BibleInfo) string {
	// Start with basic YAML structure
	content := `is_new: no
dataset_name: {{DATASET_NAME}}
bible_id: {{BIBLE_ID}}
username: jrstear
notify_ok: [jrstear@fcbhmail.org]
notify_err: [jrstear@fcbhmail.org]
audio_data: 
  bible_brain: 
    set_type_code: audio
update_dbp:
  timestamps: {{TIMESTAMPS_FILESET}}
  {{STREAM_STANZA}}
`

	// Replace template placeholders
	content = strings.ReplaceAll(content, "{{DATASET_NAME}}", bible.FilesetId)
	content = strings.ReplaceAll(content, "{{BIBLE_ID}}", bible.BibleId)
	content = strings.ReplaceAll(content, "{{TIMESTAMPS_FILESET}}", bible.AudioFileset)

	// Replace HLS/DASH stanza based on stream type
	hlsFileset := g.generateHLSFileset(bible.FilesetId)
	if g.config.StreamType == "dash" {
		content = strings.ReplaceAll(content, "{{STREAM_STANZA}}", "dash: "+hlsFileset)
	} else {
		content = strings.ReplaceAll(content, "{{STREAM_STANZA}}", "hls: "+hlsFileset)
	}

	return content
}

func (g *YAMLGenerator) generateBothYAML(bible BibleInfo) string {
	// Replace template placeholders
	content := g.template
	content = strings.ReplaceAll(content, "{{DATASET_NAME}}", bible.FilesetId)
	content = strings.ReplaceAll(content, "{{BIBLE_ID}}", bible.BibleId)
	content = strings.ReplaceAll(content, "{{TIMESTAMPS_FILESET}}", bible.AudioFileset)
	content = strings.ReplaceAll(content, "{{HLS_FILESET}}", g.generateHLSFileset(bible.FilesetId))
	content = strings.ReplaceAll(content, "{{SET_TYPE_CODE}}", g.getSetTypeCode())

	// Replace HLS/DASH stanza based on stream type
	hlsFileset := g.generateHLSFileset(bible.FilesetId)
	if g.config.StreamType == "dash" {
		content = strings.ReplaceAll(content, "{{STREAM_STANZA}}", "dash: "+hlsFileset)
	} else {
		content = strings.ReplaceAll(content, "{{STREAM_STANZA}}", "hls: "+hlsFileset)
	}

	// Replace audio type code format
	audioTypeCode := g.getAudioTypeCode()
	content = strings.ReplaceAll(content, "{{AUDIO_TYPE_CODE}}", "set_type_code: "+audioTypeCode)

	// Change output to CSV only - remove JSON line completely
	content = strings.ReplaceAll(content, "  json: yes\n", "")
	content = strings.ReplaceAll(content, "  json: no\n", "")
	content = strings.ReplaceAll(content, "  csv: yes", "  csv: yes")

	// Note: mp3_64 is optional and not included to avoid issues with filesets missing codec/bitrate tags
	// The codec/bitrate check is bypassed when not specified, allowing filesets with missing tags to match

	// Ensure YAML ends with newline
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	return content
}

func (g *YAMLGenerator) generateHLSFileset(audioFileset string) string {
	// Convert audio fileset to SA fileset based on stream type
	switch g.config.StreamType {
	case "hls":
		// e.g., ABPWBTN1DA -> ABPWBTN1SA
		return replaceSuffix(audioFileset, "DA", "SA")
	case "dash":
		// e.g., ABPWBTN1DA-opus16 -> ABPWBTN1SA-opus16
		return replaceSuffix(audioFileset, "DA-opus16", "SA-opus16")
	default:
		return replaceSuffix(audioFileset, "DA", "SA")
	}
}

func (g *YAMLGenerator) generateDuplicationYAML(bible BibleInfo) (string, bool, error) {
	sourceID := g.deriveSourceFileset(bible.AudioFileset)
	if sourceID == "" {
		return "", true, nil
	}

	hasTimestamps, err := g.sourceHasTimestamps(sourceID)
	if err != nil {
		return "", false, err
	}
	if !hasTimestamps {
		if g.config.Verbose {
			fmt.Printf("Skipping %s: source fileset %s has no timestamps\n", bible.FilesetId, sourceID)
		}
		return "", true, nil
	}

	sourceDurations, err := g.fetchChapterDurations(sourceID)
	if err != nil {
		return "", false, err
	}
	targetDurations, err := g.fetchChapterDurations(bible.AudioFileset)
	if err != nil {
		return "", false, err
	}

	matching, mismatches := compareDurationMaps(sourceDurations, targetDurations, g.config.DupTolerance)
	if len(matching) == 0 {
		return "", true, nil
	}
	if len(mismatches) > 0 {
		if g.config.Verbose {
			fmt.Printf("Skipping %s due to mismatched durations:\n", bible.FilesetId)
			for _, mismatch := range mismatches {
				fmt.Printf("  %s\n", mismatch)
			}
		}
		return "", true, nil
	}

	content := g.generateBothYAML(bible)
	content = removeTopLevelSection(content, "output:")
	content = removeTopLevelSection(content, "text_data:")
	content = removeTopLevelSection(content, "timestamps:")

	streamKey := "hls"
	if g.config.StreamType == "dash" {
		streamKey = "dash"
	}
	streamValue := g.generateHLSFileset(bible.FilesetId)
	target := fmt.Sprintf("update_dbp:\n  timestamps: %s\n  %s: %s", bible.AudioFileset, streamKey, streamValue)
	replacement := fmt.Sprintf("update_dbp:\n  copy_timestamps_from: %s\n  timestamps: %s\n  %s: %s", sourceID, bible.AudioFileset, streamKey, streamValue)
	content = strings.Replace(content, target, replacement, 1)

	return content, false, nil
}

func (g *YAMLGenerator) deriveSourceFileset(targetID string) string {
	upper := strings.ToUpper(targetID)
	sourceSuffix := strings.ToUpper(g.config.DupSource)
	switch {
	case strings.Contains(upper, "N2") && sourceSuffix == "N1":
		return strings.Replace(targetID, "N2", "N1", 1)
	case strings.Contains(upper, "O2") && sourceSuffix == "O1":
		return strings.Replace(targetID, "O2", "O1", 1)
	default:
		return ""
	}
}

func (g *YAMLGenerator) fetchChapterDurations(filesetID string) (map[string]map[int]float64, error) {
	durations := make(map[string]map[int]float64)

	query := `
		SELECT bf.book_id, bf.chapter_start, bf.duration
		FROM bible_filesets fs
		JOIN bible_files bf ON fs.hash_id = bf.hash_id
		WHERE fs.id = ? AND bf.duration IS NOT NULL
	`

	rows, err := g.db.QueryContext(g.ctx, query, filesetID)
	if err != nil {
		return nil, fmt.Errorf("fetching durations for %s: %w", filesetID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			bookID   sql.NullString
			chapter  sql.NullInt64
			duration sql.NullInt64
		)

		if err := rows.Scan(&bookID, &chapter, &duration); err != nil {
			return nil, fmt.Errorf("scanning duration row for %s: %w", filesetID, err)
		}

		if !bookID.Valid || !chapter.Valid || !duration.Valid {
			continue
		}

		bookMap, ok := durations[bookID.String]
		if !ok {
			bookMap = make(map[int]float64)
			durations[bookID.String] = bookMap
		}
		// Convert int to float64 for consistency with existing code
		bookMap[int(chapter.Int64)] = float64(duration.Int64)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating durations for %s: %w", filesetID, err)
	}

	return durations, nil
}

func (g *YAMLGenerator) sourceHasTimestamps(filesetID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM bible_filesets fs
		JOIN bible_files bf ON fs.hash_id = bf.hash_id
		JOIN bible_file_timestamps ts ON ts.bible_file_id = bf.id
		WHERE fs.id = ?
	`

	var count int
	if err := g.db.QueryRowContext(g.ctx, query, filesetID).Scan(&count); err != nil {
		return false, fmt.Errorf("checking timestamps for %s: %w", filesetID, err)
	}

	return count > 0, nil
}

func compareDurationMaps(source, target map[string]map[int]float64, tolerance float64) ([]string, []string) {
	const maxDiffLog = 5
	matching := make([]string, 0)
	mismatches := make([]string, 0)

	for bookID, chapters := range source {
		for chapter, src := range chapters {
			tgt, ok := target[bookID][chapter]
			if !ok {
				if len(mismatches) < maxDiffLog {
					mismatches = append(mismatches, fmt.Sprintf("%s %d missing target duration", bookID, chapter))
				}
				continue
			}
			if math.Abs(src-tgt) > tolerance {
				if len(mismatches) < maxDiffLog {
					mismatches = append(mismatches, fmt.Sprintf("%s %d src=%.2fs tgt=%.2fs", bookID, chapter, src, tgt))
				}
				continue
			}
			matching = append(matching, fmt.Sprintf("%s %d", bookID, chapter))
		}
	}

	return matching, mismatches
}

func removeTopLevelSection(content, key string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))
	skip := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, key) && !strings.HasPrefix(line, "  ") {
			skip = true
			continue
		}
		if skip {
			if strings.HasPrefix(line, " ") || line == "" {
				continue
			}
			skip = false
		}
		result = append(result, line)
	}

	return strings.TrimSpace(strings.Join(result, "\n")) + "\n"
}

func (g *YAMLGenerator) getSetTypeCode() string {
	switch g.config.TextType {
	case "usx":
		return "text_usx_edit"
	case "plain":
		return "text_plain"
	default:
		return "text_usx_edit"
	}
}

func (g *YAMLGenerator) getAudioTypeCode() string {
	switch g.config.Testament {
	case "n1", "o1":
		return "audio"
	case "n2", "o2":
		return "audio_drama"
	default:
		return "audio"
	}
}

func getDSN() string {
	// Use DBP_MYSQL_DSN environment variable (same as timestamp/hls insertion)
	dsn := os.Getenv("DBP_MYSQL_DSN")
	if dsn == "" {
		// Fallback to default if not set
		dsn = "root:@tcp(localhost:3306)/dbp_localtest?parseTime=true"
	}
	return dsn
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func replaceSuffix(value, oldSuffix, newSuffix string) string {
	if strings.HasSuffix(value, oldSuffix) {
		return value[:len(value)-len(oldSuffix)] + newSuffix
	}
	return value
}

func loadMMSSupportFromTree(langTree search.LanguageTree) (map[string]bool, error) {
	// Load MMS ASR supported languages from the language tree
	// This uses the same system as the rest of the MMS codebase
	support := make(map[string]bool)

	// Get all languages from the mms_asr.tab file via the language tree
	// We'll iterate through all languages and check if they're supported by MMS ASR
	for _, lang := range langTree.Table {
		if lang.Iso6393 != "" {
			// Check if this language is supported by MMS ASR
			_, distance, err := langTree.Search(lang.Iso6393, "mms_asr")
			if err == nil && distance >= 0 {
				// Language is supported (distance >= 0 means it was found)
				support[lang.Iso6393] = true
			}
		}
	}

	return support, nil
}

func (g *YAMLGenerator) isMMSSupported(languageISO string) bool {
	// Use the language tree to check if a language is supported by MMS ASR
	_, distance, err := g.langTree.Search(languageISO, "mms_asr")
	if err != nil {
		return false
	}
	// Distance >= 0 means the language was found and is supported
	return distance >= 0
}

func loadTemplate(templatePath string) (string, error) {
	if templatePath != "" {
		// Load custom template
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return "", fmt.Errorf("failed to read custom template: %v", err)
		}
		return string(content), nil
	}

	// Use default template
	return getDefaultTemplate(), nil
}

func getDefaultTemplate() string {
	return `is_new: yes
dataset_name: {{DATASET_NAME}}
bible_id: {{BIBLE_ID}}
username: jrstear
notify_ok: [jrstear@fcbhmail.org]
notify_err: [jrstear@fcbhmail.org]
output: 
  json: yes
  csv: yes
text_data: 
  bible_brain:
    {{SET_TYPE_CODE}}: yes
audio_data: 
  bible_brain: 
    {{AUDIO_TYPE_CODE}}
timestamps: 
  mms_align: yes
update_dbp:
  timestamps: {{TIMESTAMPS_FILESET}}
  {{STREAM_STANZA}}
`
}
