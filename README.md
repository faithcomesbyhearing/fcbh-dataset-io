# Artificial Polyglot

An application server that integrates AI technologies into a production system that 
streamlines the creation of audio Bible recordings. This server integrates existing 
AI tools to production workflows, automating and enhancing the process of developing 
high-quality, multilingual audio Bibles.

## Motivation

> "And this gospel of the kingdom will be proclaimed throughout the whole world as a testimony to all nations, and then the end will come." Matt 24:14.
This seems to describe a contingency on the Lord's coming, that the gospel has been proclaimed to all nations.  Certainly, the work that many are doing in translating the scriptures, and the work FCBH (Faith Comes By Hearing) is doing are critical parts.  But is it possible that the word "proclaimed" implies more than the scriptures themselves, but also the preaching of those scriptures?  Recent advances in artificial intelligence (AI) make it conceivable that technology might someday provide the ability to translate spoken words into any language.  And even though this technology might be developed by a company such as OpenAI, Google, or Microsoft; FCBH houses data that is critical to this development by having audio Bibles in a large number of languages.  Because each language is a translation of the same document, these audio translations will be especially useful to this AI task.  So, this project hopes to be a help by preparing data.

## The Opportunity

Meta (formerly Facebook) has open-sourced an AI research project called MMS 
(Massively Multilingual Speech) that supports over 1,100 languages. A significant 
portion of the training data for this model came directly from Faith Comes By 
Hearing (FCBH), specifically their online repository of text and audio Bible translations. 
This creates a strategic advantage for FCBH: since the MMS model was partially 
trained on their own New Testament and Old Testament content, FCBH can now leverage 
this same model to verify the accuracy of their audio recordings against the corresponding 
text.

Introduction to MMS<br>
https://ai.meta.com/blog/multilingual-model-speech-recognition/<br>
Technical Paper on MMS<br>
https://arxiv.org/pdf/2305.13516<br>
MMS Repository on Github<br>
https://github.com/facebookresearch/fairseq/tree/main/examples/mms<br>

## Audio Proofing by Comparison

The audio proofing module works through a multi-step process:<br>

1. **Speech-to-Text Conversion**: The MMS model transcribes the audio Bible recordings into text.
1. **Text Comparison**: The system compares this machine-generated transcript against the original source text using Google's diff-match-patch tool, which implements the Myers difference algorithm.
1. **Prioritized Error Detection**: A custom sorting algorithm organizes the differences, placing likely audio errors at the top of the report for efficient review.
1. **Interactive Results**: The system generates an HTML report that allows reviewers to:
   - View all detected discrepancies
   - Play the audio for each verse directly within the report
   - Quickly identify and address the most problematic sections

This automated approach has dramatically improved efficiency, reducing audio proofing time by 90-95% compared to traditional manual methods.

<!-- Audio Proofing is done by using MMS to perform speech to text on the audio, and then 
comparing that generated text to the original source text.  The comparison is performed using 
a text difference tool using the Myers difference algorithm, named diff-match-patch,
and developed by Google.  A custom algorithm to sort the result diff lines so that audio 
errors are near the top of the report is key. Using these methods has reduced the time to 
proof audio files by 90-95%.  The server delivers results as an HTML report that can be 
opened locally and is able play the audio of each verse.-->

MMS ASR<br>
https://huggingface.co/docs/transformers/v4.50.0/en/model_doc/mms#automatic-speech-recognition-asr<br>
Myers Diff Algorithm<br>
https://neil.fraser.name/writing/diff/myers.pdf<br>
Diff-Match-Patch<br>
https://github.com/google/diff-match-patch?tab=readme-ov-file<br>
A-Polyglot code for Speech to Text<br>
https://github.com/faithcomesbyhearing/fcbh-dataset-io/blob/main/mms/mms_asr.go<br>

## Audio Proofing by Forced Alignment

TorchAudio provides multiple forced alignment methods that synchronize audio recordings 
with their corresponding text. These systems:

1. **Generate Precise Timestamps**: Map each character in the text to its exact start and end times in the audio
2. **Calculate Confidence Scores**: Produce probability scores indicating how likely each character in the text matches what was actually spoken
   - These character-level probabilities can be aggregated to assess word-level accuracy
3. **Implementation Options**:
   - **Multilingual Model**: Offers superior accuracy but requires significant GPU resources
   - **CTC Model**: Used for processing longer audio files that would exceed GPU memory with the multilingual model

While these forced alignment techniques have sometimes identified errors with remarkable 
precision, they have proven less consistent than the "Audio Proofing by Comparison" 
method described earlier. 


<!-- Torchaudio includes a few different methods of forced alignment, by which one enters
audio and text into a model that provides audio timestamps for the beginning and ending
of each character in the audio.  In addition to timestamps, it also delivers a probability
that the character in the text was correct in the audio.  These character probabilities
can be summarized to word probabilities of correctness.  These probabilities have been
very useful for identifying errors in the audio, in some cases they have been very precise,
but have also been less consistent than the method above "Audio Proofing by Comparison".  
Two methods of forced alignment are used.  The multilingual model is most accurate but requires more
GPU.  The other is used when a audio requires too much GPU to process. -->

Multilingual Forced Alignment<br>
https://pytorch.org/audio/main/tutorials/forced_alignment_for_multilingual_data_tutorial.html<br>
CTC Forced Alignment<br>
https://pytorch.org/audio/main/tutorials/ctc_forced_alignment_api_tutorial.html<br>
A-Polyglot code for Multilingual Forced Alignment<br>
https://github.com/faithcomesbyhearing/fcbh-dataset-io/blob/main/mms/mms_align.go<br>

## Audio Timestamps Generation

Audio timestamps are essential for multiple purposes in audio Bible production:<br>
1. **Core Applications**:
   - Dividing audio chapters into verse and script segments for speech-to-text processing
   - Pinpointing specific words for audio correction
   - Validating character-level accuracy in recordings
2. **Available Methods**:
   - Aeneas forced alignment
   - Standard forced alignment
   - Multilingual forced alignment
3. **Comparative Analysis**: We evaluated the accuracy of all three methods by performing speech-to-text conversion and analyzing positional errors. The multilingual forced alignment consistently delivered the highest precision.
4. **Output Formats**: The system can export timestamp data in multiple formats:
   - JSON
   - CSV
   - SQLite database
   
This flexibility allows timestamps to be integrated into various workflow systems and downstream applications.

<!-- Audio timestamps are needed to chop audio chapters into verse segments and script segments for speech to text processing,
or to locate the position of specific words for audio correction, or to identify the probability that
specific characters are correctly presented in the audio.  This server is able to compute timestamps
using Aeneas and by the two forced alignment methods described above.  The accuracy of these
three methods was compared by doing speech to text and looking for errors of position.  
The Multilingual Forced Alignment method was the most accurate.  Output of timestamp data
can be delivered in json, csv, or as a sqlite database. -->

## Text Input Adapters

The server supports multiple Bible text formats through specialized adapters that:
1. **Process Various File Types**:
   - USX (Unified Scripture XML) an XML-based Bible encoding standard
   - Plain text format used by FCBH Bible Brain system
   - Excel (XLSX) production scripts used by FCBH for audio recording
2. **Standardize Content**: Each adapter converts its specific format into a common internal schema
3. **Store Data Efficiently**: All processed text is organized and stored in a SQLite database for unified access across the system

This flexible input system allows the server to work with whatever text resources are available without requiring format conversion by users.

<!-- Bible text can be read into the server from a number of formats, including: USX,
a plain text format used by the FCBH Bible Brain system, and an xlxs format that
FCBH uses as a script for audio production.  The system has adapters for each of
these formats that loads the text into an internal schema and stored in a sqlite database. -->

## Textual Comparison Tools

The text comparison system extends beyond audio verification to support several critical workflows:

1. **Version Control**:
   - Compare different revisions of the same USX file to identify changes
   - Track modifications between iterations of XLSX production scripts
2. **Cross-Format Validation**:
   - Verify that production scripts accurately reflect the source USX content
   - Ensure consistency between different representations of the same biblical text
3. **Intelligent Processing**:
   - When comparing USX files with scripts or audio files, the system automatically filters the USX content to include only:
       - Verse text
       - Relevant heading text

This adaptable comparison framework helps maintain content integrity across the various stages of Bible production and translation.

<!-- ## Textual Comparison

The text comparison tool that is used for audio proofing can be used in a variety of
ways.  For example, two revisions of a USX file could be compared to see changes.
Or, two revisions of a xlxs script could be compare to see changes. Or, a USX file
could be compared with a script to verify that the script properly the USX file.
When doing the USX to script comparison or USX to audio file comparison the USX file
is parsed to include only verse text, and the expected heading text. -->

## Language Selection

Most of the audios to be proofed are not one of the languages supported by MMS.  The 
current work-around of this problem is a module that selects a related language 
by doing a search of the glottolog tree to find a related language that is supported
by the MMS model.

Glottolog<br>
https://glottolog.org<br>

## Prerequisites

**System Dependencies:**
- FFmpeg and Sox (see MMS documentation for installation)
- Python 3.8+ with virtual environment

**For MMS Forced Alignment:**
- See `mms/forced_align/README.md` for detailed setup instructions
- Includes PyTorch, model download, and platform-specific commands

## Current and Future Development

* Identification and removal of false positives from proofing reports
* Ability to learn languages not supported by the MMS model
* Ability to correct selected words in audio files

## System Architecture

The server is architected as a collection of reasonably independent modules that all read 
their inputs and write their outputs to one unified schema in a sqlite database.

The user's primary input to the server is a configuration file in .yaml format where the
user specifies their inputs, the tasks they would like performed, and how their output will be presented.

## YAML Configuration

The server accepts configuration through YAML request files that specify inputs, processing tasks, and output formats. For comprehensive documentation of all available configuration options, see [README_YAML.md](yaml.md).

## Methodology

Text adapters for handling a variety of formats, and converting them all to a single format for further processing.
So, 

modules for timestamp, speech to text, and various kinds of encoding that work
with the internal format.

The FCBH audio production process breaks text Bible chapters into script 
segments (called lines) that are often a sentence long, but always include one speaker.
The audio is recorded from these scripts into chapter files with timestamps 
that mark the beginning and end of each script segment.
Text is also sourced from plain_text and USX data sources.

Sqlite is the data store to hold the text, and encodings.
It is much higher performance than any server based relational database
as long a multiple concurrent writes are not needed.
And they are not because each project has its own database.

Using the python module **Aeneas**, we process each chapter to obtain a list of
timestamps that mark the beginning and end of each script line, verse, or word.

Using a speech to text module for the language being processed, 
the generated text is used to test the correctness of the audio.
Possible speech to text tools include: **Whisper** - from OpenAI and
**MMS** - from Meta.

Comparison of source text to generated speech to text is done using
**Diff-Match-Patch** from Google.

The audio data is converted into Mel-Frequency Cepstral Coefficients (MFCCs) 
using the python module **librosa**.  This output is then broken up into script line length,
or verse length, or word length segments using the timestamps found by Aeneas.

**FastText** from Meta is used to create a word encoding of all of the available 
text in each language to be processed. **BERT** from Google and
**Word2Vec** from Gemsys are two other tools that could also be used.

Using a lexicon that provides equivalent meanings in each language, 
and other languages, multiple language encodings will be used to create a 
single multilingual encoding that will be used for both languages.
There are a few possible solutions to Facebook's **MUSE**, Google's **mBERT**, 
Google's Universal Sentence Encoder (USE), or Byte Pair Encoding (BPE) 
and Sentence Piece.

The MFCC data for each word, and the corresponding multilingual word encoding 
of both the language and the source language are used to create a tensor 
as a timeseries with the corresponding MFCC, and target language encoding.

Then the MFCC encoded word data is normalized and padded to be the same length in the
time dimension to prepare it for use by a neural net.

The tensor containing MFCC data, and encoded text, or multilingual encoded text
is loaded into a LLM (Large Language Model) or 
Recurrent Neural Net (RNN), or Natural Language Processor (NLP). 
It is likely the model would be designed to predict the next audio word.

## Database Structure

The data is organized into three tables.  An Ident table, which has only a single row in a Sqlite3 database.
This is a denormalized table that contains important identifying information about the data collected for one Bible.

The Script is a normalized table that contains one record for each script line of an audio recording,
or one verse of a plain text Bible.

The Word is a normalized table that contains one record for each word of an audio recording.

### Ident Record

**dataset_id** - A unique integer identifier for a dataset.  In this sqlite implementation, it is always 1.  But, in a central database implementation it would uniquely identify each dataset.

**bible_id** - The FCBH bible_id, often 3 char of ISO + 3 char of Version.  It is the unique identifier of a Bible, and is the common identifier given to all text, audio, and video filesets.

**audio_OT_id** - The audio OT media_id if needed

**audio_NT_id** - The audio NT media_id if needed

**text_OT_id** - The text OT media_id if needed

**text_NT_id** - The text NT media_id if needed

**text_source** - This code defines the source of the text data collected.  Possible values include: script, text_plain, text_plain_edit, text_usx_edit.

**language_iso** - The ISO language code using the ISO 639-3 standard.

**version_code** - The 3 character version code.  This is almost always the same as the last 3 digits of the bible_id.

**language_id** - The FCBH language id, which takes into account oral dialect

**rolv_id** - To be written

**alphabet** - The 4 digit code of the ISO 15924 standard.  It is also called script code.

**language_name** - The ISO 639-3 name of the language.

**version_name** - The name associated with the version_code.

### Script Record

**script_id** - A surrogate primary key.  It is an integer that begins with 1 for the first record, and increments for each record inserted.  It is present primarily to make table updates efficient.

**dataset_id** - A foreign key to the Ident table

**book_id** - The USFM 3 character book code.

**chapter_num** - An integer that defines the chapter number.

**chapter_end** - The end chapter of a piece text, almost always the same as chapter_num.

**script_num** - An integer that defines the line of a script.
For non-script text, like USX or plain text, it is an identifier that
uniquely identifies a line within a chapter.
The three fields (book_id, chapter_num, script_num) together uniquely identify a script line.

**usfm_style** - The USFM style code of the text is available for text loaded from USX. Some AI researchers might consider the style information to be a useful input for their AI model.

**person** - This is the person or character who is speaking in a script segment. It is available from text loaded from audio scripts. Narrator is the most frequent person.  It is included here because some AI researchers might find this information useful for the analysis of text language, since different people have different grammars and styles of speech.

**actor** - This is a number that identifies the actor who is speaking this script segment.  It is available from text loaded from audio scripts.  Since the Bible has more persons speaking than the number of actors available to record a Bible, actors will need to play many parts.  This data item is included because some AI researchers might find this information useful for the analysis of audio data.

**verse_str** - The starting verse number (string) of this piece of text.

**verse_end** - The ending verse number (string) of this piece of text.

**script_text** - This is the text of the script.

**script_begin_ts** -  The timestamp that marks the beginning of the text in the audio chapter file.

**script_end_ts** - The timestamp that marks the end of the script in the audio chapter file.

**mfcc_json** - Mel-Frequency Cepstral Coefficients of the audio as produced by the python library librosa, and broken into word segments using the timestamps.

### Word Record

**word_id** - A surrogate primary key.  It is an integer that begins with 1 for the first record, and increments for each record inserted.  It is present primarily to make table updates efficient.

**script_id** - A foreign key to the script table.

**word_seq** - An integer that defines the position of a word in the specific script line that it belongs to.  The columns (script_id, word_seq) are a unique index.

**verse_num** - This is typically a number, but can be a value like 2a. This column will be '0' when the word is part of a heading, reference, note, or other non-verse text.

**ttype** - A code that identifies the type of data in word. It values are (W, S, P) meaning (Word, Space, Punctuation)

**word** - The word in UTF-8 format.  This could be more than one word if needed to correctly correspond to a word in the source language.

**word_begin_ts** - The timestamp for the start of a word in an audio script segment.

**word_end_ts** - The timestamp for the end of a word in an audio script segment.

**word_enc** - A numeric encoding that uniquely represents a word, and also carries some semantic meaning.

**word_multi_enc** - An encoding of the word which uses an encoding that is common to both languages.

**mfcc_json** - Mel-Frequency Cepstral Coefficients of the audio as produced by the python library librosa, and broken into word segments using the timestamps.

-->
