#!/usr/bin/env python3
"""
Extract training data from corpus for So-VITS-SVC training
Extracts audio segments for a specific speaker/book
"""

import os
import sys
import json
import sqlite3
import subprocess
from pathlib import Path

def extract_verse_audio(db_path, base_path, book_id, chapter_num, output_dir, speaker_name="speaker0"):
    """
    Extract all verse audio segments from a book/chapter for training
    """
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    # Query for all verses in the chapter
    query = """
    SELECT DISTINCT s.verse_str, s.script_begin_ts, s.script_end_ts, s.audio_file
    FROM scripts s
    WHERE s.book_id = ? AND s.chapter_num = ? AND s.audio_file != ''
    ORDER BY s.verse_num
    """
    
    cursor.execute(query, (book_id, chapter_num))
    verses = cursor.fetchall()
    
    if not verses:
        print(f"No verses found for {book_id} {chapter_num}")
        return []
    
    # Create speaker directory
    speaker_dir = Path(output_dir) / speaker_name
    speaker_dir.mkdir(parents=True, exist_ok=True)
    
    extracted_files = []
    
    for verse_str, begin_ts, end_ts, audio_file in verses:
        # Find the audio file
        audio_path = None
        possible_paths = [
            os.path.join(base_path, "engniv2011/n1da", audio_file),
            os.path.join(base_path, audio_file),
        ]
        
        for path in possible_paths:
            if os.path.exists(path):
                audio_path = path
                break
        
        if not audio_path:
            # Search recursively
            for root, dirs, files in os.walk(base_path):
                if audio_file in files:
                    audio_path = os.path.join(root, audio_file)
                    break
        
        if not audio_path:
            print(f"Warning: Could not find audio file: {audio_file}")
            continue
        
        # Extract segment using ffmpeg
        output_filename = f"{book_id}_{chapter_num:02d}_v{verse_str.replace('-', '_')}.wav"
        output_path = speaker_dir / output_filename
        
        # Use ffmpeg to extract segment
        cmd = [
            "ffmpeg", "-i", audio_path,
            "-ss", str(begin_ts),
            "-t", str(end_ts - begin_ts),
            "-ar", "44100",  # Resample to 44.1kHz
            "-ac", "1",      # Mono
            "-y",            # Overwrite
            str(output_path)
        ]
        
        try:
            subprocess.run(cmd, check=True, capture_output=True)
            extracted_files.append(str(output_path))
            print(f"Extracted: {output_filename} ({begin_ts:.2f}s - {end_ts:.2f}s)")
        except subprocess.CalledProcessError as e:
            print(f"Error extracting {verse_str}: {e}")
            continue
    
    conn.close()
    return extracted_files

def main():
    if len(sys.argv) < 6:
        print("Usage: prepare_training_data.py <db_path> <base_path> <book_id> <chapter_num> <output_dir> [speaker_name]")
        sys.exit(1)
    
    db_path = sys.argv[1]
    base_path = sys.argv[2]
    book_id = sys.argv[3]
    chapter_num = int(sys.argv[4])
    output_dir = sys.argv[5]
    speaker_name = sys.argv[6] if len(sys.argv) > 6 else "speaker0"
    
    print(f"Extracting training data for {book_id} {chapter_num}...")
    print(f"Output directory: {output_dir}")
    print(f"Speaker name: {speaker_name}\n")
    
    files = extract_verse_audio(db_path, base_path, book_id, chapter_num, output_dir, speaker_name)
    
    print(f"\nExtracted {len(files)} audio segments")
    print(f"Training data ready in: {output_dir}/{speaker_name}/")

if __name__ == "__main__":
    main()

