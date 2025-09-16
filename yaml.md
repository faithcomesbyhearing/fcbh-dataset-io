# YAML Configuration Reference

This document provides comprehensive documentation for all available configuration options in the FCBH Dataset IO YAML request files.

## Table of Contents

- [Example Configurations](#example-configurations)
- [Required Fields](#required-fields)
- [Optional Configuration Sections](#optional-configuration-sections)
  - [Audio Data Sources](#audio-data-sources)
  - [Text Data Sources](#text-data-sources)
  - [Testament Selection](#testament-selection)
  - [Processing Detail](#processing-detail)
  - [Output Configuration](#output-configuration)
  - [Speech-to-Text Options](#speech-to-text-options)
  - [Timestamp Generation](#timestamp-generation)
  - [Audio Proofing](#audio-proofing)
  - [Text Comparison](#text-comparison)
  - [Training Configuration](#training-configuration)
  - [Audio Encoding](#audio-encoding)
  - [Text Encoding](#text-encoding)
  - [Database Configuration](#database-configuration)
  - [Update DBP (Planned Feature)](#update-dbp-planned-feature)
- [Validation Rules](#validation-rules)
- [Default Values](#default-values)

## Example Configurations

We'll start with some handy examples to get you up and running quickly, then provide detailed explanations of all available options.

### Basic Audio Proofing
```yaml
is_new: yes
dataset_name: AudioProof_Example
username: JohnDoe
bible_id: ATIWBT

testament:
  nt: yes

audio_data:
  bible_brain:
    mp3_64: yes

text_data:
  bible_brain:
    text_usx_edit: yes

timestamps:
  mms_align: yes

speech_to_text:
  mms_asr: yes

audio_proof:
  html_report: yes

output:
  directory: /output
  html_report: yes
```

### Text Comparison
```yaml
is_new: no
dataset_name: Compare_Example
username: JohnDoe
bible_id: ATIWBT

compare:
  html_report: yes
  base_dataset: Original_Dataset
  gordon_filter: 4
  compare_settings:
    lower_case: yes
    remove_punctuation: yes
    double_quotes:
      normalize: yes
```

### Training Configuration
```yaml
is_new: yes
dataset_name: Training_Example
username: JohnDoe
bible_id: ATIWBT

testament:
  nt: yes

audio_data:
  bible_brain:
    mp3_64: yes

text_data:
  bible_brain:
    text_usx_edit: yes

training:
  mms_adapter:
    batch_mb: 16
    num_epochs: 100
    learning_rate: 1e-4

output:
  directory: /output
  sqlite: yes
```

## Required Fields

These fields must be present in every YAML configuration file:

```yaml
is_new: yes                    # Answer yes to start a new project, answer no to do further processing (avoiding unnecessary recomputation)
dataset_name: Test1_ATIWBT     # unique name for this dataset (new if is_new: yes, reuse existing value if is_new: no)
username: JohnDoe              # Your username for the system (used as a top-level folder for results)
bible_id: ATIWBT               # One of these two fields is required (bible_id OR language_iso)
language_iso: ati              # Alternative to bible_id - ISO 639-3 language code
```

**Note:** Either `bible_id` or `language_iso` must be provided, but not both.

**Optional Language Override:**
```yaml
alt_language: eng                 # Force use of a specific language code (bypasses automatic language selection)
```

**Note:** `alt_language` overrides automatic language selection for MMS ASR and other AI tools. Use this when you want to force a specific language instead of letting the system find the closest supported language.

## Optional Configuration Sections

### Audio Data Sources

Choose one audio data source (only one can be selected):

```yaml
audio_data:
  bible_brain:                 # If using Bible Brain, mark ONE desired type only
    mp3_64: yes                # Mark yes for 64 kbps MP3
    mp3_16: yes                # Mark yes for 16 kbps MP3
    opus: yes                   # Mark yes for OPUS format
  file: /directory/{mediaId}/*.wav          # Local file path (include twice for OT and NT)
  aws_s3: s3://bucket/audio/{bibleId}/{mediaId}/*.mp3  # S3 path (include twice for OT and NT)
  post: {mediaId}_{A/Bseq}_{book}_{chapter}_{verse}-{chapter_end}_{verse_end}  # File path pattern for multipart uploads (see [Multipart Uploads](#multipart-uploads) section)
  no_audio: yes                # If no audio processing is needed
```

**Default:** `no_audio: yes`

**Note:** If multiple Bible Brain options are specified, the system uses the first one given (in the order listed above).

### Text Data Sources

Choose one text data source (only one can be selected):

```yaml
text_data:
  bible_brain:                 # If using Bible Brain, mark ONE desired type only
    text_usx_edit: yes         # Mark yes for USX with text not in audio removed
    text_plain_edit: yes       # Mark yes for plain text with headings added to match audio
    text_plain: yes            # Mark yes for DBP plain text
  file: /directory/{mediaId}/*.usx          # Local file path (include twice for OT and NT)
  aws_s3: s3://bucket/text/{bibleId}/{mediaId}/*.usx  # S3 path (include twice for OT and NT)
  post: {mediaId}_{A/Bseq}_{book}_{chapter}_{verse}-{chapter_end}_{verse_end}  # File path pattern for multipart uploads (see [Multipart Uploads](#multipart-uploads) section)
  no_text: yes                 # If no text processing is needed
```

**Default:** `no_text: yes`

**Note:** If multiple Bible Brain options are specified, the system uses the first one given (in the order listed above).

### Testament Selection

Choose which portions of the Bible to process:

```yaml
testament:
  nt: yes                      # Mark yes for entire New Testament
  nt_books: [MAT,MRK,LUK,JHN]  # To process part of the NT, list specific USFM NT book codes
  ot: yes                      # Mark yes for entire Old Testament
  ot_books: [GEN,EXO,LEV,NUM]  # To process part of the OT, list specific USFM OT book codes
```

**Default:** If no testament is specified, `nt: yes` is assumed.

**Note:** `nt` and `nt_books` are **not mutually exclusive** - you can specify both. If `nt: yes` is set, the entire New Testament will be processed regardless of `nt_books`. Similarly, `ot` and `ot_books` can be used together, with `ot: yes` taking precedence over `ot_books`.

### Processing Detail

Choose processing granularity:

```yaml
detail:
  lines: yes                   # Mark yes to process script lines
  words: yes                   # Mark yes to process individual words
```

**Default:** `lines: yes`

### Output Configuration

Specify the format and location of output files. **Multiple output formats can be generated in a single run**:

```yaml
output:
  directory: /path/to/output   # Server directory path where output should be written
  csv: yes                     # Mark yes for CSV output
  json: yes                    # Mark yes for JSON output
  sqlite: yes                  # Mark yes for SQLite database output
```

**Multiple Formats:** You can enable any combination of CSV, JSON, and SQLite outputs simultaneously. Each format will be generated as a separate file in the specified output directory.

### Speech-to-Text Options

Choose speech-to-text method (only one can be selected):

```yaml
speech_to_text:
  mms_asr: yes                 # Use Meta's multilingual MMS model (supports 1,100+ languages)
  adapter_asr: yes             # Use language-specific MMS adapter (requires prior training)
  whisper:                     # Use OpenAI's Speech-to-Text model
    model:                     # Choose one Whisper model size
      large: yes               # Most accurate, slowest
      medium: yes              # Good balance
      small: yes               # Faster, less accurate
      base: yes                # Fast, basic accuracy
      tiny: yes                # Fastest, least accurate
  no_speech_to_text: yes       # If speech-to-text is not needed
```

**Default:** `no_speech_to_text: yes`

**Note:** `no_speech_to_text` operates similarly to `no_training` - speech_to_text is enabled if options are given, but then disabled/ignored if "no_speech_to_text: yes" is also given.

**MMS ASR vs Adapter ASR:**

- **`mms_asr`**: Uses Meta's pre-trained multilingual MMS model (`facebook/mms-1b-all`) that supports 1,100+ languages out-of-the-box. The system automatically selects the best supported language if your target language isn't directly supported. (Note: This is different from Meta's zero-shot MMS model, which is not currently exposed through the YAML configuration.)

**Language Selection Process:**
1. If `alt_language` is specified in your YAML, that language is used directly (bypassing automatic selection)
2. If your target language isn't supported by MMS, the system uses a **Glottolog language graph** to find the closest supported language
3. The search traverses up and down the language hierarchy graph to find the nearest supported language
4. Distance is measured by the number of intervening nodes in the graph
5. The system logs which language was selected and the distance from your target language

- **`adapter_asr`**: Uses a language-specific adapter trained on your audio data. The adapter files must exist from a previous training run (either in the same submission with `training.mms_adapter` or from a prior submission). The adapter is stored locally and provides better accuracy for your specific language/dialect but only works for languages you've trained.

**When to use each:**
- Use `mms_asr` for quick testing or languages already well-supported by MMS
- Use `adapter_asr` for better accuracy on specific languages/dialects after training

**Workflow Options:**
- **Train and use in same submission**: Include both `training.mms_adapter` and `speech_to_text.adapter_asr` in the same YAML
- **Use pre-existing adapter**: Include only `speech_to_text.adapter_asr` if adapter files already exist from a previous training run (and use "is_new: no")
- **Training only**: Include only `training.mms_adapter` to train an adapter for future use

### Timestamp Generation

Choose timestamp generation method (only one can be selected):

```yaml
timestamps:
  bible_brain: yes             # Use Bible Brain timestamps (not recommended - last verse has no ending timestamp)
  aeneas: yes                  # Compute timestamps using Aeneas forced alignment (requires audio and text)
  ts_bucket: yes               # Pull timestamp data from Sandeep's bucket
  mms_fa_verse: yes            # Compute timestamps using MMS forced alignment
  mms_align: yes               # Second method for computing timestamps with word/verse scores
  no_timestamps: yes           # If timestamps are not needed
```

**Default:** `no_timestamps: yes`

**Note:** Multiple timestamp options cannot be specified - the system will return a validation error if more than one is selected.

**Note:** `mms_align` automatically sets `detail.words: yes` as a prerequisite (see [Processing Detail](#processing-detail) section).

### Audio Proofing

Generate audio proofing reports that compare speech-to-text results against original text:

```yaml
audio_proof:
  html_report: yes             # Mark yes to receive HTML proof report
  base_dataset: dataset_name   # Required when "is_new: no" to specify which dataset contains the original USX text
```

**What audio proofing does:**
- Compares speech-to-text results against original USX text
- Generates an interactive HTML report showing alignment scores and differences
- Allows audio playback validation of specific verses/lines

**Requirements:**
- For new datasets: requires `timestamps.mms_align: yes` and `speech_to_text.mms_asr: yes`
- For existing datasets: requires `base_dataset` to be specified

**How `base_dataset` works:**
- **If `is_new: yes`**: Uses current dataset's text data for comparison (any `base_dataset` value is ignored)
- **If `is_new: no`**: Uses `base_dataset` to reference a **different existing dataset** that contains the original USX text to compare against
- This allows comparing speech-to-text results against known good text from another dataset

### Text Comparison

Compare text between two different datasets:

```yaml
compare:
  html_report: yes             # Mark yes to receive HTML comparison report
  base_dataset: dataset_name   # Name of dataset to compare to this one
  gordon_filter: 4             # Optional filter: ignore differences that occur more than N times (4 = ignore if same difference appears >4 times)
  compare_settings:            # Text normalization settings
    lower_case: yes            # Convert to lowercase
    remove_prompt_chars: yes   # Remove prompt characters found in audio transcript
    remove_punctuation: yes    # Remove punctuation
    double_quotes:             # Choose no more than one
      remove: yes              # Remove all double quotes (", ", », «)
      normalize: yes           # Convert all double quotes to ASCII " (", ", », « → ")
    apostrophe:                # Choose no more than one
      remove: yes              # Remove all apostrophes (', ', ', ')
      normalize: yes           # Convert all apostrophes to ASCII ' (', ', ', ' → ')
    hyphen:                    # Choose no more than one
      remove: yes              # Remove all hyphens (–, —, ‐, ‑, etc.)
      normalize: yes           # Convert all hyphens to ASCII - (–, —, ‐, ‑ → -)
    diacritical_marks:         # Choose no more than one
      remove: yes              # Remove all diacritical marks (é → e, ñ → n)
      normalize_nfc: yes       # Unicode NFC: Canonical composition (é → é, most compact)
      normalize_nfd: yes       # Unicode NFD: Canonical decomposition (é → e + ́)
      normalize_nfkc: yes      # Unicode NFKC: Compatibility composition (² → 2, then compose)
      normalize_nfkd: yes      # Unicode NFKD: Compatibility decomposition (² → 2, separate marks)
```

**Relationship between Audio Proofing and Text Comparison:**

Both generate HTML reports, but serve different purposes:

- **Audio Proofing (`audio_proof.html_report`)**: Compares speech-to-text results against original USX text from the same or different dataset. Focuses on audio-to-text accuracy validation.

- **Text Comparison (`compare.html_report`)**: Compares text between two different datasets. Focuses on text-to-text differences and variations.

**When to use each:**
- Use **Audio Proofing** to validate speech-to-text accuracy against known good text
- Use **Text Comparison** to find differences between different text versions or translations

**Gordon Filter Details:**
- **Purpose**: Reduces false positive differences by filtering out common patterns
- **How it works**: 
  1. Identifies character-level differences between texts
  2. Counts how many times each identical difference pattern occurs
  3. If a difference pattern occurs more than the threshold, it's considered a "false positive" and removed from the report
  4. Uses a database of known words to avoid filtering legitimate differences
- **Example**: If "the" vs "The" appears 50 times, and threshold is 4, this difference is ignored as a systematic formatting issue
- **Only for text comparison**: Gordon filter is only available in the `compare` section, not in `audio_proof`

**Unicode Normalization Forms Explained:**

The four Unicode normalization forms handle character encoding differently:

**NFC vs NFD (Canonical):**
- **NFC**: Combines characters into composed forms (é stays as é) - most compact
- **NFD**: Separates characters into decomposed forms (é becomes e + ́) - separates base from marks

**NFKC vs NFKD (Compatibility):**
- **NFKC**: Converts compatibility characters (² → 2) then composes - changes appearance
- **NFKD**: Converts compatibility characters (² → 2) then decomposes - most decomposed

**Key Differences:**
- **NFC vs NFKC**: Both compose, but NFKC converts compatibility characters (² → 2) while NFC doesn't
- **NFD vs NFKD**: Both decompose, but NFKD converts compatibility characters while NFD doesn't
- **NFC vs NFD**: NFC composes (compact), NFD decomposes (separated)
- **NFKC vs NFKD**: NFKC composes after conversion, NFKD decomposes after conversion

**When to use each:**
- **NFC**: General text storage (most common)
- **NFD**: Text processing that needs to analyze individual components
- **NFKC**: Identifier matching where different representations should be equivalent
- **NFKD**: Text analysis that needs fundamental character components

### Training Configuration

Configure MMS adapter training:

```yaml
training:
  mms_adapter:                 # Do training using the MMS language adapter method
    batch_mb: 32               # Maximum size of batch in MB
    num_epochs: 50            # Number of epochs to run
    learning_rate: 1e-3       # Learning rate for training
    warmup_pct: 12.0          # Warmup percentage
    grad_norm_max: 0.4        # Maximum gradient norm
  no_training: yes             # Explicitly disable training
```

**Training Enablement:** Training is enabled if **any** `mms_adapter` parameter is specified (even if set to 0). Training is disabled if `no_training: yes` is specified or if no training parameters are provided.

**Default:** `no_training: yes` (training disabled)

**Default Values:** When training is enabled, unspecified parameters use these defaults:
- `batch_mb: 8`
- `num_epochs: 32`
- `learning_rate: 5e-5`
- `warmup_pct: 1.0`
- `grad_norm_max: 1.0`

### Database Configuration

**Advanced users only** - Configure database access and storage:

```yaml
database:
  aws_s3: s3://bucket/path/database_name.db  # Import existing database from S3 (no wildcards allowed)
  file: /local/path/to/database.db           # Use local database file
```

**Note:** When using `database.aws_s3`, `is_new` must be set to `no`.

**Advanced Usage:** These options are **not mutually exclusive**. If both are specified, the S3 database is downloaded to the local file system. Most users can omit this section entirely - the system will automatically create and manage the database locally.

### Update DBP (Planned Feature)

**⚠️ This feature is planned but not fully implemented yet**

```yaml
update_dbp:
  timestamps: []               # Array of timestamp sources to update in DBP
```

## Validation Rules

The system enforces several validation rules:

### Database Rules
- When `database.aws_s3` is set, `is_new` must be `false`

### Timestamp Rules
- Timestamps require both audio and text data
- Aeneas, MMS forced alignment methods require text data
- `mms_align` automatically enables word-level processing

### Speech-to-Text Rules
- Speech-to-text requires audio data
- Audio proofing requires MMS ASR and MMS align for new datasets
- Audio proofing requires `base_dataset` for existing datasets

### Encoding Rules
- MFCC encoding requires timestamps
- Text encoding requires text data

### Mutual Exclusivity
- Only one option can be selected from each category (audio_data, text_data, timestamps, etc.)

## Default Values

When options are not specified, the following defaults apply:

- `testament.nt: yes`
- `audio_data.no_audio: yes`
- `text_data.no_text: yes`
- `timestamps.no_timestamps: yes`
- `training.no_training: yes`
- `speech_to_text.no_speech_to_text: yes`
- `detail.lines: yes`
- `audio_encoding.no_encoding: yes`
- `text_encoding.no_encoding: yes`

## Additional Notes

- **Tab characters are NOT allowed** in YAML files
- Dataset names are automatically sanitized (spaces replaced with underscores)
- The system automatically determines language ISO codes from bible_id if not explicitly provided
- POST file uploads support multipart form data for both YAML configuration and content files
- All processing steps are logged and can be monitored through the system's logging interface

## Multipart Uploads

The system supports multipart file uploads for scenarios where you need to upload content files (audio or text) along with your YAML configuration. This is useful when files are not available via Bible Brain, folders on server, or S3.

### How It Works

When you specify `post:` in your `audio_data` or `text_data` configuration, the client automatically switches to multipart upload mode:

1. **Client Behavior**: The client reads the POST pattern and uploads files to the `/upload` endpoint
2. **Server Processing**: The server receives both the YAML configuration and uploaded files
3. **File Processing**: Uploaded files are processed according to the YAML configuration

### API Endpoints

- **`/request`** - For YAML-only requests (no file uploads)
- **`/upload`** - For multipart uploads with files

### Multipart Form Structure

The multipart form must contain these fields:

| Field Name | Content Type | Description |
|------------|--------------|-------------|
| `yaml` | `application/x-yaml` | Your YAML configuration file |
| `audio` | `audio/mpeg` or `audio/wav` | Audio file(s) (when `audio_data.post` is specified) |
| `text` | `text/plain` or `application/xml` | Text file(s) (when `text_data.post` is specified) |

### Example Usage

#### Using curl:

```bash
# Upload audio file with YAML configuration
curl -X POST http://localhost:8080/upload \
  -F "audio=@/path/to/audio.mp3;type=audio/mpeg" \
  -F "yaml=@/path/to/config.yaml;type=application/x-yaml" \
  -H "Accept: application/json"

# Upload text file with YAML configuration  
curl -X POST http://localhost:8080/upload \
  -F "text=@/path/to/text.usx;type=application/xml" \
  -F "yaml=@/path/to/config.yaml;type=application/x-yaml" \
  -H "Accept: application/json"
```

#### YAML Configuration:

```yaml
is_new: yes
dataset_name: Upload_Test
username: YourName
bible_id: ENGWEB

# Use POST pattern for audio upload
audio_data:
  post: {mediaId}_{A/Bseq}_{book}_{chapter}_{verse}-{chapter_end}_{verse_end}

# Use POST pattern for text upload  
text_data:
  post: {mediaId}_{A/Bseq}_{book}_{chapter}_{verse}-{chapter_end}_{verse_end}

speech_to_text:
  whisper:
    model:
      tiny: yes

output:
  directory: /output
  json: yes
```

### POST Pattern Format

The POST pattern uses placeholders to help the client identify which files to upload:

- `{mediaId}` - Media identifier (e.g., `ENGWEBN2DA`)
- `{A/Bseq}` - Testament sequence (`A` for OT, `B` for NT) + book sequence
- `{book}` - USFM book code (e.g., `MAT`, `GEN`)
- `{chapter}` - Chapter number
- `{verse}` - Starting verse
- `{chapter_end}` - Ending chapter (optional)
- `{verse_end}` - Ending verse (optional)

**Example patterns:**
- `{mediaId}_{A/Bseq}_{book}_{chapter}_{verse}` - Single verse
- `{mediaId}_{A/Bseq}_{book}_{chapter}_{verse}-{chapter_end}_{verse_end}` - Verse range

### File Naming Requirements

Uploaded files must follow the naming conventions expected by the system's filename parser:

**Audio files:**
- `{mediaId}_{A/Bseq}_{book}_{chapter}_{verse}.mp3`
- `{mediaId}_{A/Bseq}_{book}_{chapter}_{verse}-{chapter_end}_{verse_end}.wav`

**Text files:**
- `{book}.usx` (e.g., `MAT.usx`, `GEN.usx`)
- `{mediaId}_{A/Bseq}_{book}_{chapter}_{verse}.txt`

### Limitations

- **File size limit**: 10MB maximum per upload
- **Single file per upload**: Currently supports one content file per request
- **Temporary storage**: Uploaded files are stored temporarily and cleaned up after processing

### Error Handling

Common upload errors:

- **400 Bad Request**: Invalid YAML configuration or unsupported file format
- **413 Payload Too Large**: File exceeds 10MB limit
- **415 Unsupported Media Type**: Incorrect content type specified
- **500 Internal Server Error**: Server-side processing error

### Client Implementation

The system includes a Go client (`controller/FCBHDataset/client.go`) that automatically handles multipart uploads when POST patterns are detected. For custom implementations, ensure your client:

1. Detects POST patterns in YAML configuration
2. Constructs multipart form with correct field names
3. Sets appropriate content types
4. Handles server responses and error cases
