#!/usr/bin/env python3
"""
Prosody Matching Module for Audio Revision
Uses DSP-based methods (librosa, pyworld) for prosody adjustment
"""

import os
import sys
import json
import numpy as np
import librosa
import soundfile as sf

# Try to import pyworld for better F0 extraction
try:
    import pyworld as pw
    PYWORLD_AVAILABLE = True
except ImportError:
    PYWORLD_AVAILABLE = False

# Add parent directory to path for error handler
sys.path.insert(0, os.path.abspath(os.path.join(os.environ.get('GOPROJ', '.'), 'logger')))
try:
    from error_handler import setup_error_handler
    setup_error_handler()
except ImportError:
    pass  # Error handler not required

def extract_f0_pyworld(audio, sample_rate, hop_length=160):
    """Extract F0 using pyworld (more accurate)"""
    if not PYWORLD_AVAILABLE:
        return None, None
    
    try:
        # Convert to float64 for pyworld
        audio_f64 = audio.astype(np.float64)
        
        # Extract F0 using DIO + StoneMask
        frame_period = hop_length * 1000 / sample_rate
        f0, timeaxis = pw.dio(audio_f64, sample_rate, frame_period=frame_period)
        f0 = pw.stonemask(audio_f64, f0, timeaxis, sample_rate)
        
        # Create voiced mask
        voiced = f0 > 0
        
        return f0, voiced
    except Exception as e:
        print(f"Warning: pyworld F0 extraction failed: {e}", file=sys.stderr)
        return None, None

def extract_f0_librosa(audio, sample_rate, hop_length=160):
    """Extract F0 using librosa (fallback)"""
    try:
        # Use librosa's pitch tracking
        f0, voiced_flag, voiced_probs = librosa.pyin(
            audio,
            fmin=librosa.note_to_hz('C2'),
            fmax=librosa.note_to_hz('C7'),
            sr=sample_rate,
            hop_length=hop_length
        )
        
        # Replace NaN with 0
        f0 = np.nan_to_num(f0, nan=0.0)
        voiced = voiced_flag.astype(bool)
        
        return f0, voiced
    except Exception as e:
        print(f"Warning: librosa F0 extraction failed: {e}", file=sys.stderr)
        return None, None

def extract_f0(audio, sample_rate, hop_length=160, method="auto"):
    """Extract F0 using best available method"""
    if method == "pyworld" or (method == "auto" and PYWORLD_AVAILABLE):
        f0, voiced = extract_f0_pyworld(audio, sample_rate, hop_length)
        if f0 is not None:
            return f0, voiced
    
    # Fallback to librosa
    f0, voiced = extract_f0_librosa(audio, sample_rate, hop_length)
    if f0 is not None:
        return f0, voiced
    
    # Last resort: return zeros
    length = len(audio) // hop_length
    return np.zeros(length), np.zeros(length, dtype=bool)

def calculate_pitch_shift_semitones(f0_source, f0_ref, voiced_source, voiced_ref):
    """Calculate pitch shift in semitones needed to match reference"""
    # Only use voiced frames
    f0_source_voiced = f0_source[voiced_source]
    f0_ref_voiced = f0_ref[voiced_ref]
    
    if len(f0_source_voiced) == 0 or len(f0_ref_voiced) == 0:
        return 0.0
    
    # Calculate mean F0
    mean_f0_source = np.mean(f0_source_voiced)
    mean_f0_ref = np.mean(f0_ref_voiced)
    
    if mean_f0_source == 0:
        return 0.0
    
    # Convert Hz difference to semitones
    semitones = 12 * np.log2(mean_f0_ref / mean_f0_source)
    
    return semitones

def calculate_time_stretch_ratio(source_duration, ref_duration):
    """Calculate time stretch ratio to match reference duration"""
    if source_duration == 0:
        return 1.0
    
    return ref_duration / source_duration

def normalize_loudness(source_audio, reference_audio):
    """Normalize source audio RMS to match reference"""
    # Calculate RMS
    rms_source = np.sqrt(np.mean(source_audio ** 2))
    rms_ref = np.sqrt(np.mean(reference_audio ** 2))
    
    if rms_source == 0:
        return source_audio
    
    # Scale to match reference RMS
    scale_factor = rms_ref / rms_source
    normalized = source_audio * scale_factor
    
    # Prevent clipping
    max_val = np.max(np.abs(normalized))
    if max_val > 0.95:
        normalized = normalized * (0.95 / max_val)
    
    return normalized

def match_prosody(source_audio_path, reference_audio_path, output_path, config):
    """
    Adjust source audio to match reference audio's prosody
    
    Args:
        source_audio_path: Path to source audio file (to be adjusted)
        reference_audio_path: Path to reference audio file (prosody source)
        output_path: Path for output audio file
        config: Configuration dict (F0 method, pitch shift range, etc.)
    
    Returns:
        dict with success status and output path
    """
    try:
        sample_rate = config.get("sample_rate", 16000)
        f0_method = config.get("f0_method", "auto")
        pitch_shift_range = config.get("pitch_shift_range", 2.0)  # Max semitones
        time_stretch_range = config.get("time_stretch_range", 1.2)  # Max stretch factor
        hop_length = config.get("hop_length", 160)
        
        # Load audio files
        source_audio, sr_source = librosa.load(source_audio_path, sr=sample_rate)
        reference_audio, sr_ref = librosa.load(reference_audio_path, sr=sample_rate)
        
        # Extract F0 from both
        f0_source, voiced_source = extract_f0(source_audio, sample_rate, hop_length, f0_method)
        f0_ref, voiced_ref = extract_f0(reference_audio, sample_rate, hop_length, f0_method)
        
        # Calculate pitch shift
        pitch_shift = calculate_pitch_shift_semitones(f0_source, f0_ref, voiced_source, voiced_ref)
        
        # Clamp pitch shift to allowed range
        pitch_shift = np.clip(pitch_shift, -pitch_shift_range, pitch_shift_range)
        
        # Calculate time stretch
        source_duration = len(source_audio) / sample_rate
        ref_duration = len(reference_audio) / sample_rate
        time_stretch = calculate_time_stretch_ratio(source_duration, ref_duration)
        
        # Clamp time stretch to allowed range
        time_stretch = np.clip(time_stretch, 1.0 / time_stretch_range, time_stretch_range)
        
        # Apply pitch shift
        if abs(pitch_shift) > 0.1:  # Only shift if significant
            adjusted_audio = librosa.effects.pitch_shift(
                source_audio,
                sr=sample_rate,
                n_steps=pitch_shift
            )
        else:
            adjusted_audio = source_audio.copy()
        
        # Apply time stretch
        if abs(time_stretch - 1.0) > 0.05:  # Only stretch if significant
            adjusted_audio = librosa.effects.time_stretch(
                adjusted_audio,
                rate=time_stretch
            )
        
        # Normalize loudness to match reference
        adjusted_audio = normalize_loudness(adjusted_audio, reference_audio)
        
        # Ensure output directory exists
        os.makedirs(os.path.dirname(output_path), exist_ok=True)
        
        # Save output
        sf.write(output_path, adjusted_audio, sample_rate)
        
        return {
            "success": True,
            "output_path": output_path,
            "pitch_shift_semitones": float(pitch_shift),
            "time_stretch_ratio": float(time_stretch),
            "source_duration": float(source_duration),
            "reference_duration": float(ref_duration)
        }
        
    except Exception as e:
        import traceback
        error_msg = f"{str(e)}\n{traceback.format_exc()}"
        print(f"Error in match_prosody: {error_msg}", file=sys.stderr)
        return {
            "success": False,
            "error": str(e),
            "output_path": None
        }

def extract_prosody_features(audio_path, config):
    """
    Extract prosody features from audio
    
    Args:
        audio_path: Path to audio file
        config: Configuration dict
    
    Returns:
        dict with prosody features (F0, energy, timing)
    """
    try:
        sample_rate = config.get("sample_rate", 16000)
        f0_method = config.get("f0_method", "auto")
        hop_length = config.get("hop_length", 160)
        
        # Load audio
        audio, sr = librosa.load(audio_path, sr=sample_rate)
        duration = len(audio) / sample_rate
        
        # Extract F0
        f0, voiced = extract_f0(audio, sample_rate, hop_length, f0_method)
        f0_voiced = f0[voiced] if len(f0) > 0 else np.array([])
        
        # Calculate F0 statistics
        f0_mean = float(np.mean(f0_voiced)) if len(f0_voiced) > 0 else 0.0
        f0_std = float(np.std(f0_voiced)) if len(f0_voiced) > 0 else 0.0
        
        # Calculate energy (RMS)
        rms = librosa.feature.rms(y=audio, hop_length=hop_length)[0]
        energy_mean = float(np.mean(rms))
        energy_std = float(np.std(rms))
        
        # Calculate speaking rate (approximate: duration / number of voiced frames)
        speaking_rate = len(f0_voiced) / duration if duration > 0 else 0.0
        
        return {
            "success": True,
            "features": {
                "f0_mean": f0_mean,
                "f0_std": f0_std,
                "f0_contour": f0.tolist() if len(f0) > 0 else [],
                "energy_mean": energy_mean,
                "energy_std": energy_std,
                "speaking_rate": speaking_rate,
                "duration": duration
            }
        }
        
    except Exception as e:
        import traceback
        error_msg = f"{str(e)}\n{traceback.format_exc()}"
        print(f"Error in extract_prosody_features: {error_msg}", file=sys.stderr)
        return {
            "success": False,
            "error": str(e),
            "features": None
        }

def main():
    # Handle multiple requests in a loop (like MMS ASR script)
    for line in sys.stdin:
        try:
            if not line.strip():
                continue
            
            # Parse request
            config = json.loads(line)
            mode = config.get("mode")
            
            if not mode:
                result = {"success": False, "error": "Missing 'mode' in config"}
                print(json.dumps(result))
                sys.stdout.flush()
                continue
            
            if mode == "match":
                source = config.get("source")
                reference = config.get("reference")
                output = config.get("output")
                if not source or not reference or not output:
                    result = {"success": False, "error": "Missing required arguments: source, reference, output"}
                    print(json.dumps(result))
                    sys.stdout.flush()
                    continue
                result = match_prosody(source, reference, output, config)
            elif mode == "extract":
                source = config.get("source")
                if not source:
                    result = {"success": False, "error": "Missing required argument: source"}
                    print(json.dumps(result))
                    sys.stdout.flush()
                    continue
                result = extract_prosody_features(source, config)
            else:
                result = {"success": False, "error": f"Invalid mode: {mode}"}
            
            print(json.dumps(result))
            sys.stdout.flush()
            
        except json.JSONDecodeError as e:
            result = {"success": False, "error": f"Failed to parse JSON: {e}"}
            print(json.dumps(result))
            sys.stdout.flush()
        except Exception as e:
            import traceback
            error_msg = f"{str(e)}\n{traceback.format_exc()}"
            result = {"success": False, "error": error_msg}
            print(json.dumps(result))
            sys.stdout.flush()

if __name__ == "__main__":
    main()
