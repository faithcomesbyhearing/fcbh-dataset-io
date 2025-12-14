# Jude Test Case Analysis

**Date**: 2025-12-13  
**Goal**: Convert 1984 audio to match 2011 text for book of Jude

## Files Located

1. **Audio File**: `/Users/jrstear/tmp/arti/files/engniv2011/n1da/B26___01_Jude________ENGNIVN1DA.mp3`
   - Size: 2.2MB
   - Date: Aug 7, 2018
   - Contains: Full chapter of Jude (1984 recording)

2. **1984 USX Text**: `/Users/jrstear/tmp/arti/files/engniv2011/1984usx/065JUD.usx`
   - Source text for the 1984 recording

3. **2011 USX Text**: `/Users/jrstear/tmp/arti/files/engniv2011/usx/JUD.usx`
   - Target text (what audio should say)

4. **Database**: `/Users/jrstear/fcbh/arti/old/engniv2011/00005/database/engniv2011.db`
   - Contains: ASR transcription, word-level timestamps
   - Verse 1 timestamps: 1.96s - 12.97s
   - Script ID: 21153

5. **Compare HTML**: `/Users/jrstear/fcbh/arti/engniv2011_audio_compare.html`
   - Shows differences between ASR transcription and 2011 USX text

## Key Findings

### Database Transcription (ASR Result)
- Verse 1 text: "Jude, a servant of Jesus Christ and a brother of James, To those who have been called, who are loved **in** God the Father and kept **for** Jesus Christ:"
- **Note**: Database transcription already matches 2011 USX text ("loved in", "kept for")

### HTML Compare Shows
- Audio actually says: "jed" (typo) → should be "jude"
- Audio says: "loved **by**" → should be "loved **in**" (per 2011 USX)
- Audio says: "kept **by**" → should be "kept **for**" (per 2011 USX)

### Text Version Differences (1984 USX vs 2011 USX)
- "loved **by**" → "loved **in**"
- "kept **by**" → "kept **for**"
- "godless **men**" → "ungodly **people**"
- "certain **men**" → "certain **individuals**"
- "I felt I **had to**" → "I felt **compelled to**"
- "entrusted to **the saints**" → "entrusted to **God's holy people**"
- "from **falling**" → "from **stumbling**"
- "snatch others from the fire **and save them**" → "**save others by** snatching them from the fire"

## Important Discovery & Clarifications

### Source of Truth
**Human review is the baseline truth**. Key principles:

1. **Recordings don't exactly match target text**: Audio may not match even the source text (1984 USX)
2. **Transcriptions aren't perfect**: ASR results have errors
3. **Human review decides**: Human reviews compare.html and identifies which words actually need changing
4. **Whole-word differences are more significant**: Entire word differences (e.g., "by" → "in") are more likely to need correction than sub-word differences (e.g., "jed" vs "jude" - just a few letters)

### Jude 1:1 Human Review Results
- **"jude" (first word)**: ✅ **NO CHANGE NEEDED** - Audio is fine, transcription error only
- **"loved by" → "loved in"**: ✅ **CHANGE NEEDED** - Valid whole-word difference
- **"kept by" → "kept for"**: ✅ **CHANGE NEEDED** - Valid whole-word difference

**Test Case**: Revise these two word changes in verse 1.

## Test Case: Jude 1:1

### Current State
- **Audio file**: `B26___01_Jude________ENGNIVN1DA.mp3`
- **Verse 1 timestamps**: 1.96s - 12.97s
- **Script ID**: 21153
- **Word count**: 61 words (with timestamps)

### Required Changes (based on human review)
1. ~~Fix typo: "jed" → "jude"~~ ❌ **NO CHANGE** - Audio is fine, transcription error only
2. ✅ Change: "loved by" → "loved in" - **VALID CHANGE**
3. ✅ Change: "kept by" → "kept for" - **VALID CHANGE**

**Test Case Focus**: Revise these two word changes in verse 1.

### Important Discovery: Database vs Audio Mismatch

**The database transcription shows 2011 text, but the audio says 1984 text:**
- Database: "loved **in** God" and "kept **for** Jesus" (2011 text)
- Audio: "loved **by** God" and "kept **by** Jesus" (1984 text)

**Replacement snippets found in nearby verses:**
- "in" found in JUD 1:2 (word_seq 14, 15.46-15.52s) - closest match, same speaker
- "for" found in JUD 1:3 (word_seq 55, 26.54-26.62s) - close match, same speaker

**Challenge**: The word "by" is not in the database (because transcription was corrected). We need to:
1. Run word-level alignment on verse 1 with 1984 text to find "by" timestamps
2. Extract "in" from 1:2 and "for" from 1:3
3. Replace "by" segments with extracted snippets

### Workflow for Revision

1. **Identify words to replace**:
   - "jed" → "jude"
   - "by" (after "loved") → "in"
   - "by" (after "kept") → "for"

2. **Search corpus** for replacement snippets:
   - Find "jude" elsewhere in corpus
   - Find "loved in" phrase
   - Find "kept for" phrase

3. **Extract audio snippets** from corpus matches

4. **Voice conversion** (if different speaker):
   - Extract speaker embedding from original Jude audio
   - Convert replacement snippets to match original speaker

5. **Prosody matching**:
   - Extract prosody from surrounding context (verses 1-2)
   - Adjust replacement snippets to match prosody

6. **Stitch revised audio** back into chapter file

## Questions & Next Steps

### Decisions Made

1. **Source of truth**: Human review of compare.html is baseline
   - Human identifies which differences actually need correction
   - Whole-word differences prioritized over sub-word differences
   - Transcription errors (like "jed" vs "jude") don't require audio changes

2. **Corpus for search**: 
   - ✅ Use same `engniv2011` dataset (1984 audio) as source corpus
   - Search for replacement words/phrases in existing 1984 recordings

3. **Word-level timestamps**: 
   - ✅ Database has word timestamps for verse 1
   - Use these to identify exact word boundaries for replacement

4. **Actor/Speaker info**:
   - Not available yet (will come later)
   - For now, assume same speaker for all Jude audio

### Next Steps

1. **Verify audio content**:
   - Listen to verse 1 audio segment (1.96s - 12.97s)
   - Confirm what it actually says vs database vs HTML

2. **Create revision request**:
   - Build YAML or JSON structure for Jude 1:1 revisions
   - Specify exact words/phrases to replace

3. **Test corpus search**:
   - Use `CorpusSearcher.FindReplacementSnippets()` to find:
     - "jude" (to fix typo)
     - "loved in" (phrase match)
     - "kept for" (phrase match)

4. **Test snippet extraction**:
   - Extract audio segments using word timestamps
   - Use `utility/ffmpeg.ChopOneSegment()`

5. **Test voice conversion** (if needed):
   - Extract speaker embedding from original Jude audio
   - Convert replacement snippets

6. **Test prosody matching**:
   - Extract prosody from surrounding verses
   - Adjust replacement snippets

7. **Stitch and validate**:
   - Replace segments in chapter audio
   - Generate revised audio file
   - Compare with original

## Database Schema Notes

- Database location: `/Users/jrstear/fcbh/arti/old/engniv2011/00005/database/engniv2011.db`
- Tables: `scripts`, `words`, `chars`, `ident`, `script_mfcc`, `word_mfcc`
- Verse 1 has 61 words with timestamps
- Audio file path stored in `scripts.audio_file`
- Word timestamps: `word_begin_ts`, `word_end_ts`

## Test Results

### Corpus Search Test (Jude 1:1)

**Test Date**: 2025-12-13

**Findings**:
1. **Phrase matching**: No exact phrase matches found for "loved in" or "kept for" in corpus (excluding JUD)
   - System correctly falls back to single word matches
   - Found "loved" in MAL 1:2, 1TH 1:4, etc.
   - Found "kept" in 1PE 1:4, 1SA 1:6, etc.

2. **Single word search**: Working correctly
   - Found "in" in JUD 1:2 (distance: 1) - nearby verse, good candidate
   - Found "for" in JUD 1:3 (distance: 2) - nearby verse, good candidate

3. **Target verse exclusion**: Fixed - JUD 1:1 now excluded from search results

**Strategy Decision Needed**:
- **Option A**: Use single word matches from nearby verses (JUD 1:2, 1:3)
  - Pros: Same speaker, very close context
  - Cons: Need to extract individual words and assemble
- **Option B**: Use phrase matches from other books (MAL, 1PE, etc.)
  - Pros: Full phrases available
  - Cons: Different context, may need voice conversion
- **Option C**: Hybrid - prefer nearby single words, fallback to distant phrases

**Recommendation**: Start with Option A (nearby verses) for Jude 1:1 test case.

## Implementation Status

- ✅ Corpus search implemented (`corpus_search.go`)
- ✅ Target verse exclusion working
- ⏳ Snippet extraction (next: `arti-7qf`)
- ⏳ Voice conversion (next: `arti-lmh`)
- ⏳ Prosody matching (next: `arti-a2g`)
- ⏳ Audio stitching (next: `arti-bue`)

