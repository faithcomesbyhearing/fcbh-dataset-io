#!/usr/bin/env python3
"""
Measure silence gap durations at boundaries in audio file.
"""
import json
import sys
import librosa
import numpy as np

def measure_gap_durations(audio_file, start_time, end_time, look_window=0.2):
    """Measure gap durations before and after a segment"""
    y, sr = librosa.load(audio_file, sr=None, mono=True)
    
    def time_to_samples(t):
        return int(t * sr)
    
    def find_silence_gap(y_segment, threshold_db=-50):
        """Find silence gap in audio segment - returns gap before and after audio content"""
        if len(y_segment) == 0:
            return 0.0, 0.0
        
        # Use a more sensitive threshold
        y_db = librosa.amplitude_to_db(np.abs(y_segment), ref=1.0)
        silent = y_db < threshold_db
        
        non_silent = np.where(~silent)[0]
        if len(non_silent) == 0:
            # All silence - return full duration as gap
            return len(y_segment) / sr, 0.0
        
        first_audio = non_silent[0]
        last_audio = non_silent[-1]
        
        gap_before = first_audio / sr
        gap_after = (len(y_segment) - last_audio - 1) / sr
        
        return gap_before, gap_after
    
    # Measure gap before segment (gap between previous content and segment start)
    # Look for the gap that starts at start_time and continues until next audio content
    # This is the gap we need to preserve when inserting replacement
    look_after = min(len(y) / sr, start_time + 1.0)  # Look up to 1s after start_time
    after_samples_start = time_to_samples(start_time)
    after_samples_end = time_to_samples(look_after)
    if after_samples_end > after_samples_start:
        # Find first audio content after start_time
        # Use a more lenient threshold - look for significant audio (not just any signal)
        after_segment = y[after_samples_start:after_samples_end]
        after_db = librosa.amplitude_to_db(np.abs(after_segment), ref=1.0)
        
        # Use a threshold that distinguishes between silence/gap and actual speech
        # Speech typically has peaks above -40dB, gaps are usually below -50dB
        # But we want to find where speech STARTS, so look for sustained audio above threshold
        threshold = -45  # More lenient threshold
        after_silent = after_db < threshold
        
        # Find first sustained audio (not just a brief spike)
        # Look for at least 50ms of audio above threshold
        min_audio_duration = int(0.05 * sr)  # 50ms
        after_non_silent = np.where(~after_silent)[0]
        
        if len(after_non_silent) > 0:
            # Find first sustained audio segment
            first_audio_start = None
            for i in range(len(after_non_silent) - min_audio_duration):
                if after_non_silent[i + min_audio_duration] - after_non_silent[i] < min_audio_duration:
                    # Found sustained audio
                    first_audio_start = after_non_silent[i]
                    break
            
            if first_audio_start is not None:
                gap_before = first_audio_start / sr
            else:
                # Fallback: use first non-silent sample
                gap_before = after_non_silent[0] / sr
        else:
            # No audio found in window - gap is at least the window size
            gap_before = look_after - start_time
    else:
        gap_before = 0.0
    
    # Measure gap after segment (silence just after end_time)
    # Use a longer window to capture the full gap
    after_start = end_time
    after_end = min(len(y) / sr, end_time + 0.5)  # Look up to 500ms after
    after_samples_start = time_to_samples(after_start)
    after_samples_end = time_to_samples(after_end)
    if after_samples_end > after_samples_start:
        gap_after, _ = find_silence_gap(y[after_samples_start:after_samples_end])
    else:
        gap_after = 0.0
    
    return gap_before, gap_after

def main():
    if len(sys.argv) < 4:
        print(json.dumps({"error": "Usage: measure_gaps.py <audio_file> <start_time> <end_time>"}))
        sys.exit(1)
    
    audio_file = sys.argv[1]
    start_time = float(sys.argv[2])
    end_time = float(sys.argv[3])
    
    try:
        gap_before, gap_after = measure_gap_durations(audio_file, start_time, end_time)
        result = {
            "gap_before": gap_before,
            "gap_after": gap_after
        }
        print(json.dumps(result))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)

if __name__ == "__main__":
    main()

