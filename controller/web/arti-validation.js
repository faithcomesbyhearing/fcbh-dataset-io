/**
 * Artie Folder Validation Library
 * JavaScript implementation of Go validation logic from input/utility.go
 */

// USFM Book sequence mapping (from db/book_sequence.go)
const BOOK_SEQ_MAP = {
    'GEN': 1, 'EXO': 2, 'LEV': 3, 'NUM': 4, 'DEU': 5, 'JOS': 6, 'JDG': 7, 'RUT': 8,
    '1SA': 9, '2SA': 10, '1KI': 11, '2KI': 12, '1CH': 13, '2CH': 14, 'EZR': 15, 'NEH': 16,
    'EST': 17, 'JOB': 18, 'PSA': 19, 'PRO': 20, 'ECC': 21, 'SNG': 22, 'ISA': 23, 'JER': 24,
    'LAM': 25, 'EZK': 26, 'DAN': 27, 'HOS': 28, 'JOL': 29, 'AMO': 30, 'OBA': 31, 'JON': 32,
    'MIC': 33, 'NAM': 34, 'HAB': 35, 'ZEP': 36, 'HAG': 37, 'ZEC': 38, 'MAL': 39,
    'MAT': 41, 'MRK': 42, 'LUK': 43, 'JHN': 44, 'ACT': 45, 'ROM': 46, '1CO': 47, '2CO': 48,
    'GAL': 49, 'EPH': 50, 'PHP': 51, 'COL': 52, '1TH': 53, '2TH': 54, '1TI': 55, '2TI': 56,
    'TIT': 57, 'PHM': 58, 'HEB': 59, 'JAS': 60, '1PE': 61, '2PE': 62, '1JN': 63, '2JN': 64,
    '3JN': 65, 'JUD': 66, 'REV': 67, 'TOB': 68, 'JDT': 69, 'ESG': 70, 'WIS': 71, 'SIR': 72,
    'BAR': 73, 'LJE': 74, 'S3Y': 75, 'SUS': 76, 'BEL': 77, '1MA': 78, '2MA': 79, '3MA': 80,
    '4MA': 81, '1ES': 82, '2ES': 83, 'MAN': 84, 'PS2': 85, 'ODA': 86, 'PSS': 87, 'EZA': 88,
    '5EZ': 89, '6EZ': 90, 'DAG': 91, 'PS3': 92, '2BA': 93, 'LBA': 94, 'JUB': 95, 'ENO': 96,
    '1MQ': 97, '2MQ': 98, '3MQ': 100, 'REP': 101, '4BA': 102, 'LAO': 103, 'FRT': 104, 'BAK': 105,
    'OTH': 106, 'INT': 107, 'CNC': 108, 'GLO': 109, 'TDX': 110, 'NDX': 111, 'XXA': 112, 'XXB': 113,
    'XXC': 114, 'XXD': 115, 'XXE': 116, 'XXF': 117, 'XXG': 118
};

// Book code corrections (from input/utility.go lines 347-369)
const BOOK_CORRECTIONS = {
    "PSM": "PSA", "PRV": "PRO", "SOS": "SNG", "EZE": "EZK", "JOE": "JOL", "NAH": "NAM",
    "MRC": "MRK", "LUC": "LUK", "JUA": "JHN", "HEC": "ACT", "EFE": "EPH", "FHP": "PHP",
    "1TE": "1TH", "2TE": "2TH", "TTO": "TIT", "TTL": "TIT", "TTS": "TIT", "FHM": "PHM",
    "JMS": "JAS", "SNT": "JAS", "APO": "REV"
};

/**
 * Validate and correct book ID (equivalent to validateBookId in Go)
 */
function validateBookId(bookId) {
    // Apply corrections
    const corrected = BOOK_CORRECTIONS[bookId];
    if (corrected) {
        bookId = corrected;
    }
    
    // Check if book exists in sequence map
    if (!BOOK_SEQ_MAP.hasOwnProperty(bookId)) {
        return {
            valid: false,
            bookId: bookId,
            error: `BookId "${bookId}" is not known. Available corrections: ${JSON.stringify(BOOK_CORRECTIONS)}`
        };
    }
    
    return {
        valid: true,
        bookId: bookId
    };
}

/**
 * Determine testament from book ID (equivalent to db.Testament function)
 */
function getTestament(bookId) {
    const seq = BOOK_SEQ_MAP[bookId];
    if (!seq) return 'None';
    if (seq < 40) return 'OT';
    if (seq < 68) return 'NT';
    return 'AP';
}

/**
 * Parse VOX audio filename (equivalent to parseVOXAudioFilename in Go)
 * Format: {N1/N2/O1/O2}_{ISO}_{VERSION}_{BOOKSEQ}_{BOOKCODE}_{CHAPTER}_{VERSE}_VOX.{mp3|wav}
 * Example: N2_ANL_BSM_001_MAT_001_VOX.mp3
 */
function parseVOXAudioFilename(filename) {
    const parts = filename.split('_');
    if (parts.length !== 7) {
        return {
            valid: false,
            error: `VOX filename expected to have 7 parts, got ${parts.length}: ${filename}`
        };
    }
    
    const [drama, langCode, versionCode, bookSeq, bookCode, chapter, verseWithExt] = parts;
    const verse = verseWithExt.replace(/\.(mp3|wav)$/, '');
    
    // Validate book code
    const bookValidation = validateBookId(bookCode);
    if (!bookValidation.valid) {
        return {
            valid: false,
            error: bookValidation.error
        };
    }
    
    // Determine testament from drama prefix
    let testament;
    if (drama[0] === 'N') {
        testament = 'NT';
    } else if (drama[0] === 'O') {
        testament = 'OT';
    } else if (drama[0] === 'P') {
        testament = getTestament(bookValidation.bookId);
    } else {
        return {
            valid: false,
            error: `Unknown media type: ${drama}`
        };
    }
    
    // Validate chapter number
    const chapterNum = parseInt(chapter);
    if (isNaN(chapterNum) || chapterNum < 1) {
        return {
            valid: false,
            error: `Invalid chapter number: ${chapter}`
        };
    }
    
    return {
        valid: true,
        drama: drama,
        langCode: langCode,
        versionCode: versionCode,
        bookSeq: bookSeq,
        bookId: bookValidation.bookId,
        chapter: chapterNum,
        verse: verse,
        testament: testament,
        mediaId: langCode + versionCode + drama + "DA"
    };
}

/**
 * Parse V4 audio filename (equivalent to parseV4AudioFilename in Go)
 * Format: {MEDIAID}_{A/B}{BOOKSEQ}_{BOOKCODE}_{CHAPTER}[_{VERSE}-{CHAPTER_END}_{VERSE_END}].{mp3|wav}
 * Example: ENGESVN2DA_B001_MAT_001.mp3
 */
function parseV4AudioFilename(filename) {
    const fileExt = filename.match(/\.(mp3|wav)$/)?.[0] || '';
    const nameWithoutExt = filename.replace(fileExt, '');
    const cleanName = nameWithoutExt.replace(/-/g, '_');
    const parts = cleanName.split('_');
    
    if (parts.length < 4) {
        return {
            valid: false,
            error: `V4 filename expected at least 4 parts, got ${parts.length}: ${filename}`
        };
    }
    
    const [mediaId, abSeq, bookCode, chapter, ...remainingParts] = parts;
    
    // Determine testament from A/B prefix
    let testament;
    if (abSeq[0] === 'A') {
        testament = 'OT';
    } else if (abSeq[0] === 'B') {
        testament = 'NT';
    } else {
        return {
            valid: false,
            error: `Invalid A/B prefix: ${abSeq[0]}`
        };
    }
    
    const bookSeq = abSeq.substring(1);
    
    // Validate book code
    const bookValidation = validateBookId(bookCode);
    if (!bookValidation.valid) {
        return {
            valid: false,
            error: bookValidation.error
        };
    }
    
    // Validate chapter number
    const chapterNum = parseInt(chapter);
    if (isNaN(chapterNum) || chapterNum < 1) {
        return {
            valid: false,
            error: `Invalid chapter number: ${chapter}`
        };
    }
    
    // Parse optional verse range
    let verse, chapterEnd, verseEnd;
    if (remainingParts.length >= 1) {
        verse = remainingParts[0];
    }
    if (remainingParts.length >= 3) {
        chapterEnd = parseInt(remainingParts[1]);
        verseEnd = remainingParts[2];
        if (isNaN(chapterEnd) || chapterEnd < 1) {
            return {
                valid: false,
                error: `Invalid chapter end number: ${remainingParts[1]}`
            };
        }
    }
    
    return {
        valid: true,
        mediaId: mediaId,
        testament: testament,
        bookSeq: bookSeq,
        bookId: bookValidation.bookId,
        chapter: chapterNum,
        verse: verse,
        chapterEnd: chapterEnd,
        verseEnd: verseEnd
    };
}

/**
 * Parse USX text filename (equivalent to logic in lines 147-166)
 * Format: {BOOKSEQ}{BOOKCODE}.usx or {BOOKCODE}.usx
 * Example: 040MAT.usx, 001GEN.usx, or MAT.usx
 */
function parseUSXFilename(filename) {
    if (!filename.endsWith('.usx')) {
        return {
            valid: false,
            error: `USX file must end with .usx: ${filename}`
        };
    }
    
    const nameWithoutExt = filename.replace('.usx', '');
    let bookId, bookSeq;
    
    if (nameWithoutExt.length === 10) {
        // Format: 001GEN.usx (3-digit sequence + 3-letter code + .usx = 10 chars)
        bookSeq = nameWithoutExt.substring(0, 3);
        bookId = nameWithoutExt.substring(3, 6);
    } else if (nameWithoutExt.length === 7) {
        // Format: GEN.usx (3-letter code + .usx = 7 chars)
        bookId = nameWithoutExt.substring(0, 3);
        bookSeq = BOOK_SEQ_MAP[bookId]?.toString() || '';
    } else if (nameWithoutExt.length === 6) {
        // Format: 040MAT.usx (3-digit sequence + 3-letter code = 6 chars, common format)
        bookSeq = nameWithoutExt.substring(0, 3);
        bookId = nameWithoutExt.substring(3, 6);
    } else {
        return {
            valid: false,
            error: `USX files expected in format 001GEN.usx, 040MAT.usx, or GEN.usx, got: ${filename}`
        };
    }
    
    // Validate book code
    const bookValidation = validateBookId(bookId);
    if (!bookValidation.valid) {
        return {
            valid: false,
            error: bookValidation.error
        };
    }
    
    return {
        valid: true,
        bookId: bookValidation.bookId,
        bookSeq: bookSeq,
        testament: getTestament(bookValidation.bookId)
    };
}

/**
 * Extract book code from any filename by looking for 3-letter patterns
 */
function extractBookCodeFromFilename(filename) {
    // Remove file extension
    const nameWithoutExt = filename.replace(/\.(mp3|wav|usx)$/i, '');
    
    // Look for 3-letter book codes in the filename
    // Common patterns: MAT_001, _MAT_, MAT001, 040MAT, etc.
    const bookCodePatterns = [
        /[^A-Za-z]([A-Z]{3})[^A-Za-z]/,  // _MAT_ or _MAT001
        /^([A-Z]{3})[^A-Za-z]/,          // MAT_001 (start of string)
        /[^A-Za-z]([A-Z]{3})$/,          // 040MAT (end of string)
        /([A-Z]{3})/,                    // Any 3 uppercase letters
    ];
    
    for (const pattern of bookCodePatterns) {
        const match = nameWithoutExt.match(pattern);
        if (match) {
            const potentialBookCode = match[1];
            // Validate that this is actually a valid book code
            if (BOOK_SEQ_MAP.hasOwnProperty(potentialBookCode)) {
                return potentialBookCode;
            }
        }
    }
    
    return null;
}

/**
 * Validate audio file based on filename patterns - ROBUST VERSION
 */
function validateAudioFile(filename) {
    if (!filename.match(/\.(mp3|wav)$/i)) {
        return {
            valid: false,
            error: `Audio file must end with .mp3 or .wav: ${filename}`
        };
    }
    
    // Try specific patterns first (for better error messages)
    if (filename.includes('_VOX.') && (filename.endsWith('.mp3') || filename.endsWith('.wav'))) {
        const voxResult = parseVOXAudioFilename(filename);
        if (voxResult.valid) {
            return voxResult;
        }
    }
    
    if (filename.includes('_') && (filename.endsWith('.mp3') || filename.endsWith('.wav'))) {
        const v4Result = parseV4AudioFilename(filename);
        if (v4Result.valid) {
            return v4Result;
        }
    }
    
    // Fallback: Extract book code from any pattern
    const bookCode = extractBookCodeFromFilename(filename);
    if (bookCode) {
        return {
            valid: true,
            bookId: bookCode,
            testament: getTestament(bookCode),
            mediaType: 'robust_audio'
        };
    }
    
    return {
        valid: false,
        error: `No valid book code found in audio filename: ${filename}`
    };
}

/**
 * Validate text file (USX) - ROBUST VERSION
 */
function validateTextFile(filename) {
    if (!filename.toLowerCase().endsWith('.usx')) {
        return {
            valid: false,
            error: `Text file must be .usx format: ${filename}`
        };
    }
    
    // Try specific USX patterns first
    const usxResult = parseUSXFilename(filename);
    if (usxResult.valid) {
        return usxResult;
    }
    
    // Fallback: Extract book code from any pattern
    const bookCode = extractBookCodeFromFilename(filename);
    if (bookCode) {
        return {
            valid: true,
            bookId: bookCode,
            bookSeq: BOOK_SEQ_MAP[bookCode]?.toString() || '',
            testament: getTestament(bookCode),
            mediaType: 'robust_usx'
        };
    }
    
    return {
        valid: false,
        error: `No valid book code found in USX filename: ${filename}`
    };
}

/**
 * Extract books from file list
 */
function extractBooksFromFiles(files, validator) {
    const books = new Set();
    const errors = [];
    
    for (const file of files) {
        const result = validator(file.name);
        if (result.valid) {
            books.add(result.bookId);
        } else {
            errors.push(`${file.name}: ${result.error}`);
        }
    }
    
    return { books, errors };
}

/**
 * Main validation function for folder structure
 */
function validateFolderStructure(folderData) {
    const errors = [];
    const warnings = [];
    
    // Check for required folders
    if (!folderData.audioFiles || folderData.audioFiles.length === 0) {
        errors.push('No audio files found. Expected audio files in .mp3 or .wav format.');
    }
    
    if (!folderData.textFiles || folderData.textFiles.length === 0) {
        errors.push('No text files found. Expected USX files (.usx format).');
    }
    
    if (errors.length > 0) {
        return {
            valid: false,
            errors: errors,
            warnings: warnings
        };
    }
    
    // Validate audio files
    const audioResult = extractBooksFromFiles(folderData.audioFiles, validateAudioFile);
    if (audioResult.errors.length > 0) {
        errors.push(...audioResult.errors);
    }
    
    // Validate text files
    const textResult = extractBooksFromFiles(folderData.textFiles, validateTextFile);
    if (textResult.errors.length > 0) {
        errors.push(...textResult.errors);
    }
    
    // Check book matching between audio and text
    const audioBooks = Array.from(audioResult.books);
    const textBooks = Array.from(textResult.books);
    
    const audioOnlyBooks = audioBooks.filter(book => !textBooks.includes(book));
    const textOnlyBooks = textBooks.filter(book => !audioBooks.includes(book));
    
    if (audioOnlyBooks.length > 0) {
        errors.push(`Audio files contain books not found in text: ${audioOnlyBooks.join(', ')}`);
    }
    
    if (textOnlyBooks.length > 0) {
        errors.push(`Text files contain books not found in audio: ${textOnlyBooks.join(', ')}`);
    }
    
    // Check for minimum book count
    if (audioBooks.length === 0) {
        errors.push('No valid audio files found after validation');
    }
    
    if (textBooks.length === 0) {
        errors.push('No valid text files found after validation');
    }
    
    return {
        valid: errors.length === 0,
        errors: errors,
        warnings: warnings,
        audioBooks: audioBooks,
        textBooks: textBooks,
        totalAudioFiles: folderData.audioFiles.length,
        totalTextFiles: folderData.textFiles.length
    };
}

/**
 * Parse folder name to extract components
 * Pattern: {N1|N2|O1|O2}{ISO}{BIBLEID}
 * Example: N2ANLBSM -> drama=N2, iso=ANL, bibleId=ANLBSM
 * Handles folder names with extra text like "N2ANLBSM Khongso (ANL)"
 */
function parseFolderName(folderName) {
    // Remove any trailing slashes or spaces
    const cleanName = folderName.replace(/[/\\\s]+$/, '');
    
    // Extract the core pattern from the beginning of the folder name
    // This handles cases like "N2ANLBSM Khongso (ANL)" by taking just "N2ANLBSM"
    const coreMatch = cleanName.match(/^([NO][12][A-Z]{3}[A-Z0-9]+)/);
    
    if (!coreMatch) {
        return {
            valid: false,
            error: `Folder name does not match expected pattern. Expected: N1/N2/O1/O2 + 3-letter ISO + Bible ID. Got: ${cleanName}`
        };
    }
    
    const coreName = coreMatch[1];
    
    // Pattern: N1/N2/O1/O2 + ISO (3 chars) + BIBLEID (rest)
    const match = coreName.match(/^([NO][12])([A-Z]{3})(.+)$/);
    
    if (!match) {
        return {
            valid: false,
            error: `Core folder name does not match expected pattern. Expected: N1/N2/O1/O2 + 3-letter ISO + Bible ID. Got: ${coreName}`
        };
    }
    
    const [, drama, iso, bibleId] = match;
    
    return {
        valid: true,
        drama: drama,
        iso: iso,
        bibleId: bibleId,
        datasetName: `${iso}${bibleId}${drama}DA`
    };
}
