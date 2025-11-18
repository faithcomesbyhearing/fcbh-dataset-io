# Audio Chopper Utility

A simple utility to chop an MP3 file (Bible chapter) into verse segments using timestamps from a CSV file.

## Usage

```bash
go run chop_audio.go -audio <mp3_file> -timestamps <csv_file> [options]
```

### Required Arguments

- `-audio`: Path to the input MP3 file (Bible chapter)
- `-timestamps`: Path to CSV file containing verse timestamps

### Optional Arguments

- `-output`: Output directory for verse segments (default: same directory as audio file)
- `-book`: Book ID (e.g., MAT, MRK, LUK) - required if not in CSV
- `-chapter`: Chapter number - required if not in CSV

## CSV Format

The CSV file should have a header row with the following columns:

**Required columns:**
- `verse_str`: Verse identifier (e.g., "1", "2", "1-2")
- `script_begin_ts` or `begin_ts`: Start timestamp in seconds (float)
- `script_end_ts` or `end_ts`: End timestamp in seconds (float)

Note: Arti's CSV output uses `script_begin_ts` and `script_end_ts`, which are automatically detected.

**Optional columns:**
- `book_id`: Book ID (e.g., MAT, MRK) - can also use `-book` flag
- `chapter_num`: Chapter number - can also use `-chapter` flag

### Example CSV

```csv
book_id,chapter_num,verse_str,begin_ts,end_ts
MAT,1,1,0.0,3.5
MAT,1,2,3.5,8.2
MAT,1,3,8.2,12.1
```

Or a simpler format (using flags for book/chapter):

```csv
verse_str,begin_ts,end_ts
1,0.0,3.5
2,3.5,8.2
3,8.2,12.1
```

## Examples

### Using CSV with book/chapter columns

```bash
go run chop_audio.go \
  -audio /path/to/MAT_1.mp3 \
  -timestamps /path/to/timestamps.csv \
  -output /path/to/output/
```

### Using CSV without book/chapter (using flags)

```bash
go run chop_audio.go \
  -audio /path/to/MAT_1.mp3 \
  -timestamps /path/to/timestamps.csv \
  -book MAT \
  -chapter 1 \
  -output /path/to/output/
```

## Output

The utility will create individual MP3 files for each verse segment in the output directory. Files are named according to the [Bible Brain V4 naming convention](https://docs.google.com/document/d/1ytVKiyzTXmPsEz170UHTlZ8NXeXxUOrqNgpPurm_AGE/edit?tab=t.0#heading=h.y3ecefloodui):

**Format:** `{mediaid}_{A/B}{ordering}_{USFM book code}_{chapter start}[_{verse start}-{chapter stop}_{verse stop}].mp3`

Where:
- `{mediaid}` - Media ID extracted from the input filename (e.g., `ENGNIVN1DA`, `ENGESVN2DA`)
- `{A/B}{ordering}` - Testament prefix (A for Old Testament, B for New Testament) followed by 3-digit book ordering number
- `{USFM book code}` - 3-letter USFM book code (e.g., `MAT`, `JUD`, `1TH`)
- `{chapter start}` - 3-digit chapter number (e.g., `001`, `002`)
- `[_{verse start}-{chapter stop}_{verse stop}]` - Optional verse range (2-digit verse numbers)

**Examples:**
- `ENGNIVN1DA_B026_JUD_001_01-001_01.mp3` - Single verse (Jude 1:1)
- `ENGNIVN1DA_B026_JUD_001_01-001_10.mp3` - Verse range (Jude 1:1-10)
- `ENGESVN2DA_B001_MAT_001.mp3` - Full chapter (Matthew 1, no verse suffix)
- `IRUNLCP1DA_B013_1TH_001_01-001_010.mp3` - Partial chapter (1 Thessalonians 1:1-10)

The media ID is automatically extracted from the input audio filename. The book ordering is calculated based on the book's position in the Bible (using the standard USFM book sequence).

## Requirements

- Go environment
- ffmpeg installed and available in PATH
- CSV file with timestamps (can be exported from arti using the CSV output feature)

## Getting Timestamps from Arti

If you've run arti and have timestamps in the database, you can export them to CSV using:

```bash
# Using the convert_timestamps_to_csv utility
go run cli_misc/convert_timestamps_to_csv/simple_csv_converter.go <book_id> <chapter_num> output.csv
```

Or configure arti to output CSV by adding to your YAML:

```yaml
output:
  csv: yes
```

