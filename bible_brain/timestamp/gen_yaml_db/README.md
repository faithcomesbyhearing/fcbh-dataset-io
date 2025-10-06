# YAML Generator for Database

A CLI utility to generate YAML files for Bible Brain dataset processing, specifically targeting languages that have audio and text filesets but are missing corresponding SA (stream audio) filesets.

## Purpose

This tool generates YAML configuration files for processing languages that:
1. Have N1DA/N2DA/O1DA/O2DA audio filesets with `content_loaded=1`
2. Have corresponding text filesets (USX or plain text)
3. Are missing the corresponding SA filesets (which will be created)
4. Are supported by MMS (Massively Multilingual Speech)

## Usage

```bash
# Basic usage
./yaml_generator -testament n1 -text usx -output ./n1_usx/

# Generate N2DA filesets (with additional check for no N1SA)
./yaml_generator -testament n2 -text plain -output ./n2_plain/

# Use custom template
./yaml_generator -testament n1 -text usx -template ./my_template.yaml -output ./custom/

# Generate for specific Bible ID
./yaml_generator -testament n1 -text usx -bible ABPWBT -output ./single/

# Verbose output
./yaml_generator -testament n1 -text usx -output ./n1_usx/ -verbose
```

## Arguments

- `-testament`: Testament scope (required)
  - `n1`: New Testament 1 (N1DA audio filesets)
  - `n2`: New Testament 2 (N2DA audio filesets) 
  - `o1`: Old Testament 1 (O1DA audio filesets)
  - `o2`: Old Testament 2 (O2DA audio filesets)

- `-text`: Text format (required)
  - `usx`: USX text filesets (`_ET-usx`)
  - `plain`: Plain text filesets (`_ET`)

- `-stream`: Stream format (optional, default: hls)
  - `hls`: HLS streaming (processes filesets ending in `DA`)
  - `dash`: DASH streaming (processes filesets ending in `DA-opus16`)

- `-output`: Output directory (required)
  - Directory where YAML files will be written
  - Will be created if it doesn't exist

- `-template`: Custom template file (optional)
  - Path to custom YAML template
  - Uses default template if not specified

- `-bible`: Specific Bible ID (optional)
  - Generate YAML for only this Bible ID
  - If not specified, generates for all matching languages

- `-verbose`: Verbose output (optional)
  - Shows detailed progress information

## Output

- Files are named as `{filesetid}.yaml` (e.g., `ABPWBTN1DA.yaml`)
- Output directory structure: `{testament}_{text}/` (e.g., `n1_usx/`, `n2_plain/`)

## Template System

### Default Template
The default template includes:
- Basic dataset configuration
- MMS align for timestamps
- Update DBP section for both timestamps and HLS creation
- Placeholders for dynamic content

### Template Placeholders
- `{{DATASET_NAME}}`: Generated dataset name (e.g., `ABPWBTN1DA_TS`)
- `{{BIBLE_ID}}`: Bible ID (e.g., `ABPWBT`)
- `{{TIMESTAMPS_FILESET}}`: Audio fileset for timestamps (e.g., `ABPWBTN1DA`)
- `{{HLS_FILESET}}`: SA fileset for HLS (e.g., `ABPWBTN1SA`)
- `{{SET_TYPE_CODE}}`: Text type code (`text_usx_edit` or `text_plain`)
- `{{AUDIO_TYPE_CODE}}`: Audio type code (`audio` or `audio_drama`)

### Custom Templates
You can provide custom templates using the `-template` argument. Custom templates should include the same placeholders for dynamic content replacement.

## Database Requirements

- MySQL database with Bible Brain schema
- Tables: `bibles`, `languages`, `bible_filesets`, `bible_fileset_connections`
- Environment variable for database connection:
  - `DBP_MYSQL_DSN`: MySQL connection string (e.g., `user:password@tcp(hostname:port)/database`)
  - Falls back to `root:@tcp(localhost:3306)/dbp_localtest?parseTime=true` if not set

## Special Cases

### N2 Filesets
When generating N2 filesets (`-testament n2`), the tool includes an additional check to ensure no corresponding N1SA fileset exists. This prevents conflicts in the system.

### MMS Language Support
Only languages supported by MMS ASR (as defined in the language tree) will have YAML files generated. Languages not supported by MMS are skipped with a warning in verbose mode.

## Validation

The tool validates that:
1. Audio filesets exist and have `content_loaded=1`
2. Text filesets exist and have `content_loaded=1`
3. Corresponding SA filesets do NOT exist
4. Languages are supported by MMS
5. For N2 cases, no corresponding N1SA filesets exist

## Examples

### Generate N1DA + USX YAMLs (HLS)
```bash
./yaml_generator -testament n1 -text usx -output ./n1_usx/
```

This will find all languages with N1DA audio filesets and USX text filesets but no N1SA filesets, and generate YAML files for MMS-supported languages.

### Generate N2DA + Plain text YAMLs (DASH)
```bash
./yaml_generator -testament n2 -text plain -stream dash -output ./n2_plain/
```

This will find all languages with N2DA audio filesets and plain text filesets but no N2SA filesets, and also ensure no N1SA filesets exist.

## Template Example

```yaml
is_new: yes
dataset_name: {{DATASET_NAME}}
bible_id: {{BIBLE_ID}}
username: jrstear
notify_ok: [jrstear@fcbhmail.org]
notify_err: [jrstear@fcbhmail.org]
output: 
  json: yes
  csv: yes
testament: 
  nt_books: [MAT, MRK, LUK, JHN]
text_data: 
  bible_brain:
    {{SET_TYPE_CODE}}: yes
audio_data: 
  bible_brain: 
    {{AUDIO_TYPE_CODE}}: yes
timestamps: 
  mms_align: yes
update_dbp:
  timestamps: {{TIMESTAMPS_FILESET}}
  hls: {{HLS_FILESET}}
```

## Building

```bash
cd bible_brain/timestamp/gen_yaml_db/
go mod tidy
go build -o yaml_generator
```
