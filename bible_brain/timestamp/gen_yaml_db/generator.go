package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type BibleInfo struct {
	BibleId      string
	LanguageISO  string
	FilesetId    string
	AudioFileset string
	TextFileset  string
}

type YAMLGenerator struct {
	config     Config
	db         *sql.DB
	template   string
	mmssupport map[string]bool
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

	// Load MMS supported languages
	mmsSupport, err := loadMMSSupport()
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
	var query string
	var args []interface{}

	if g.config.BibleId != "" {
		// Generate for specific Bible ID
		query = g.buildSpecificBibleQuery()
		args = []interface{}{
			g.config.BibleId,
			g.getAudioPattern(),
			g.getTextPattern(),
		}
		// Add exclusion pattern if needed
		if g.getTextExclusionPattern() != "" {
			args = append(args, g.getTextExclusionPattern())
		}
		args = append(args, g.getSAPattern())
	} else {
		// Generate for all matching languages
		query = g.buildDiscoveryQuery()
		args = []interface{}{
			g.getAudioPattern(),
			g.getTextPattern(),
		}
		// Add exclusion pattern if needed
		if g.getTextExclusionPattern() != "" {
			args = append(args, g.getTextExclusionPattern())
		}
		args = append(args, g.getSAPattern())
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

		// Check MMS support
		if !g.mmssupport[bible.LanguageISO] {
			if g.config.Verbose {
				fmt.Printf("Skipping %s: language %s not supported by MMS\n", bible.BibleId, bible.LanguageISO)
			}
			continue
		}

		bibles = append(bibles, bible)
	}

	return bibles, nil
}

func (g *YAMLGenerator) buildDiscoveryQuery() string {
	query := `
		SELECT 
			b.id as bible_id, 
			l.iso as language_iso,
			fs.id as fileset_id,
			fs.id as audio_fileset,
			MIN(text_fs.id) as text_fileset
		FROM bibles b
		JOIN languages l ON b.language_id = l.id
		JOIN bible_fileset_connections bfc ON b.id = bfc.bible_id
		JOIN bible_filesets fs ON bfc.hash_id = fs.hash_id
		JOIN bible_fileset_connections text_bfc ON b.id = text_bfc.bible_id
		JOIN bible_filesets text_fs ON text_bfc.hash_id = text_fs.hash_id
		WHERE fs.id LIKE ? 
		AND fs.content_loaded = 1`

	query += `
		AND text_fs.id LIKE ?
		AND text_fs.content_loaded = 1`

	// Add exclusion pattern for text filesets if needed
	exclusionPattern := g.getTextExclusionPattern()
	if exclusionPattern != "" {
		query += `
		AND NOT EXISTS (
			SELECT 1 FROM bible_filesets excl_fs 
			JOIN bible_fileset_connections excl_bfc ON excl_fs.hash_id = excl_bfc.hash_id
			WHERE excl_bfc.bible_id = b.id 
			AND excl_fs.id LIKE ?
		)`
	}

	query += `
		AND NOT EXISTS (
			SELECT 1 FROM bible_filesets sa_fs 
			JOIN bible_fileset_connections sa_bfc ON sa_fs.hash_id = sa_bfc.hash_id
			WHERE sa_bfc.bible_id = b.id 
			AND sa_fs.id LIKE ?
		)`

	// Add N2 special case: no corresponding N1SA
	if strings.HasPrefix(g.config.Testament, "n2") {
		query += `
		AND NOT EXISTS (
			SELECT 1 FROM bible_filesets n1sa_fs 
			JOIN bible_fileset_connections n1sa_bfc ON n1sa_fs.hash_id = n1sa_bfc.hash_id
			WHERE n1sa_bfc.bible_id = b.id 
			AND n1sa_fs.id LIKE '%N1SA%'
		)`
	}

	query += ` GROUP BY b.id, l.iso, fs.id ORDER BY b.id, fs.id`

	return query
}

func (g *YAMLGenerator) buildSpecificBibleQuery() string {
	query := `
		SELECT 
			b.id as bible_id, 
			l.iso as language_iso,
			fs.id as fileset_id,
			fs.id as audio_fileset,
			MIN(text_fs.id) as text_fileset
		FROM bibles b
		JOIN languages l ON b.language_id = l.id
		JOIN bible_fileset_connections bfc ON b.id = bfc.bible_id
		JOIN bible_filesets fs ON bfc.hash_id = fs.hash_id
		JOIN bible_fileset_connections text_bfc ON b.id = text_bfc.bible_id
		JOIN bible_filesets text_fs ON text_bfc.hash_id = text_fs.hash_id
		WHERE b.id = ?
		AND fs.id LIKE ? 
		AND fs.content_loaded = 1`

	query += `
		AND text_fs.id LIKE ?
		AND text_fs.content_loaded = 1`

	// Add exclusion pattern for text filesets if needed
	exclusionPattern := g.getTextExclusionPattern()
	if exclusionPattern != "" {
		query += `
		AND NOT EXISTS (
			SELECT 1 FROM bible_filesets excl_fs 
			JOIN bible_fileset_connections excl_bfc ON excl_fs.hash_id = excl_bfc.hash_id
			WHERE excl_bfc.bible_id = b.id 
			AND excl_fs.id LIKE ?
		)`
	}

	query += `
		AND NOT EXISTS (
			SELECT 1 FROM bible_filesets sa_fs 
			JOIN bible_fileset_connections sa_bfc ON sa_fs.hash_id = sa_bfc.hash_id
			WHERE sa_bfc.bible_id = b.id 
			AND sa_fs.id LIKE ?
		)`

	// Add N2 special case: no corresponding N1SA
	if strings.HasPrefix(g.config.Testament, "n2") {
		query += `
		AND NOT EXISTS (
			SELECT 1 FROM bible_filesets n1sa_fs 
			JOIN bible_fileset_connections n1sa_bfc ON n1sa_fs.hash_id = n1sa_bfc.hash_id
			WHERE n1sa_bfc.bible_id = b.id 
			AND n1sa_fs.id LIKE '%N1SA%'
		)`
	}

	query += ` GROUP BY b.id, l.iso, fs.id ORDER BY b.id, fs.id`

	return query
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

	// Write to file
	filename := filepath.Join(g.config.OutputDir, bible.FilesetId+".yaml")
	return os.WriteFile(filename, []byte(content), 0644)
}

func (g *YAMLGenerator) generateHLSFileset(audioFileset string) string {
	// Convert audio fileset to SA fileset based on stream type
	switch g.config.StreamType {
	case "hls":
		// e.g., ABPWBTN1DA -> ABPWBTN1SA
		return strings.ReplaceAll(audioFileset, "DA", "SA")
	case "dash":
		// e.g., ABPWBTN1DA-opus16 -> ABPWBTN1SA-opus16
		return strings.ReplaceAll(audioFileset, "DA-opus16", "SA-opus16")
	default:
		return strings.ReplaceAll(audioFileset, "DA", "SA")
	}
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

func loadMMSSupport() (map[string]bool, error) {
	// For now, return a hardcoded list of common MMS-supported languages
	// In a full implementation, this would load from the language tree
	support := make(map[string]bool)

	// Common MMS-supported languages (from mms_asr.tab)
	mmssupported := []string{
		"eng", "spa", "fra", "deu", "ita", "por", "rus", "jpn", "kor", "cmn",
		"ara", "hin", "ben", "urd", "tam", "tel", "mar", "guj", "kan", "mal",
		"abp", "abi", "acr", "adx", "aeu", "agd", "agu", "amh", "asm", "aze",
		"bam", "ben", "bod", "bul", "cat", "ceb", "ces", "cmn", "cym", "dan",
		"deu", "ell", "eng", "est", "eus", "fas", "fin", "fra", "gle", "glg",
		"guj", "hat", "hau", "heb", "hin", "hrv", "hun", "ibo", "ind", "isl",
		"ita", "jav", "jpn", "kan", "kat", "kaz", "khm", "kin", "kir", "kor",
		"kur", "lao", "lat", "lav", "lit", "lug", "mal", "mar", "mkd", "mlg",
		"mlt", "mon", "mri", "msa", "mya", "nep", "nld", "nor", "nso", "nya",
		"ori", "orm", "pan", "pol", "por", "pus", "que", "ron", "run", "rus",
		"sin", "slk", "slv", "sna", "som", "sot", "spa", "sqi", "srp", "ssw",
		"sun", "swa", "swe", "tam", "tel", "tgk", "tha", "tir", "ton", "tsn",
		"tso", "tur", "ukr", "umb", "urd", "uzb", "ven", "vie", "xho", "yor",
		"zul",
	}

	for _, lang := range mmssupported {
		support[lang] = true
	}

	return support, nil
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
