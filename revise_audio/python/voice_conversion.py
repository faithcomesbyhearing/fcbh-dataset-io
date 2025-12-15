#!/usr/bin/env python3
"""
Voice Conversion Module for Audio Revision
Uses FCBH RVC system for voice conversion
"""

import os
import sys
import json
import argparse
import numpy as np
import librosa
import soundfile as sf
import torch

# Add FCBH RVC codebase to path
# Try multiple possible locations
possible_paths = [
    # Relative to arti repo
    os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))), 
                 "..", "FCBH-W2V-Bert-2.0-ASR-trainer"),
    # Absolute path from GOPROJ
    os.path.join(os.environ.get("GOPROJ", ""), "..", "FCBH-W2V-Bert-2.0-ASR-trainer"),
    # Common workspace location
    os.path.join(os.path.expanduser("~"), "git", "FCBH-W2V-Bert-2.0-ASR-trainer"),
]

for fcbh_rvc_path in possible_paths:
    if os.path.exists(fcbh_rvc_path):
        sys.path.insert(0, fcbh_rvc_path)
        break

# Try to import RVC components
RVC_AVAILABLE = False
try:
    from rvc_feature_extractor import RVCFeatureExtractor
    RVC_AVAILABLE = True
except ImportError as e:
    print(f"Warning: RVC feature extractor not available: {e}", file=sys.stderr)
except Exception as e:
    # Handle other import errors (e.g., speechbrain/torchaudio compatibility)
    print(f"Warning: RVC feature extractor import failed: {e}", file=sys.stderr)
    RVC_AVAILABLE = False

# Add parent directory to path for error handler
sys.path.insert(0, os.path.abspath(os.path.join(os.environ.get('GOPROJ', '.'), 'logger')))
try:
    from error_handler import setup_error_handler
    setup_error_handler()
except ImportError:
    pass  # Error handler not required

def extract_speaker_embedding_simple(audio_path, output_path, config):
    """
    Simple speaker embedding extraction using librosa (fallback when RVC not available)
    
    Args:
        audio_path: Path to audio file
        output_path: Path to save embedding (numpy .npy file)
        config: Configuration dict (sample_rate, etc.)
    
    Returns:
        dict with success status and embedding path
    """
    try:
        sample_rate = config.get("sample_rate", 16000)
        
        # Load audio
        audio, sr = librosa.load(audio_path, sr=sample_rate)
        
        # Extract mel spectrogram features
        mels = librosa.feature.melspectrogram(y=audio, sr=sample_rate, n_mels=80, hop_length=160)
        
        # Compute statistics across time to create embedding
        # Mean, std, min, max for each mel band
        embedding = np.concatenate([
            np.mean(mels, axis=1),  # Mean
            np.std(mels, axis=1),   # Std
            np.min(mels, axis=1),   # Min
            np.max(mels, axis=1),   # Max
        ])
        
        # Normalize to unit vector
        embedding = embedding / (np.linalg.norm(embedding) + 1e-10)
        
        # Save embedding
        os.makedirs(os.path.dirname(output_path), exist_ok=True)
        np.save(output_path, embedding.astype(np.float32))
        
        return {
            "success": True,
            "embedding_path": output_path,
            "embedding_dim": len(embedding),
            "method": "simple_librosa"
        }
        
    except Exception as e:
        return {
            "success": False,
            "error": str(e),
            "embedding_path": None
        }

def extract_speaker_embedding(audio_path, output_path, config):
    """
    Extract speaker embedding from audio using RVC feature extractor (or fallback)
    
    Args:
        audio_path: Path to audio file
        output_path: Path to save embedding (numpy .npy file)
        config: Configuration dict (sample_rate, f0_method, etc.)
    
    Returns:
        dict with success status and embedding path
    """
    if RVC_AVAILABLE:
        try:
            sample_rate = config.get("sample_rate", 16000)
            f0_method = config.get("f0_method", "rmvpe")
            device = config.get("device", "auto")
            
            # Initialize feature extractor
            extractor = RVCFeatureExtractor(
                sample_rate=sample_rate,
                f0_method=f0_method,
                device=device
            )
            extractor.log_func = lambda msg: print(f"[RVC] {msg}", file=sys.stderr)
            
            # Load audio
            audio, sr = librosa.load(audio_path, sr=sample_rate)
            
            # Extract speaker embedding
            speaker_emb = extractor.extract_speaker_embedding(audio, method="ecapa")
            
            # Save embedding
            os.makedirs(os.path.dirname(output_path), exist_ok=True)
            np.save(output_path, speaker_emb)
            
            return {
                "success": True,
                "embedding_path": output_path,
                "embedding_dim": len(speaker_emb),
                "method": "rvc_ecapa"
            }
            
        except Exception as e:
            print(f"Warning: RVC extraction failed, using fallback: {e}", file=sys.stderr)
            # Fall through to simple extraction
    
    # Fallback to simple extraction
    return extract_speaker_embedding_simple(audio_path, output_path, config)

def extract_spectral_envelope(audio, sample_rate, n_fft=2048, hop_length=512, n_mels=80):
    """Extract spectral envelope using mel spectrogram"""
    mel_spec = librosa.feature.melspectrogram(
        y=audio, 
        sr=sample_rate, 
        n_fft=n_fft, 
        hop_length=hop_length, 
        n_mels=n_mels
    )
    # Convert to log scale
    mel_spec_db = librosa.power_to_db(mel_spec, ref=np.max)
    return mel_spec_db

def apply_spectral_envelope(source_audio, target_envelope, sample_rate, n_fft=2048, hop_length=512, n_mels=80):
    """Apply target spectral envelope to source audio"""
    from scipy import interpolate
    
    # Get source mel spectrogram
    source_mel = librosa.feature.melspectrogram(
        y=source_audio,
        sr=sample_rate,
        n_fft=n_fft,
        hop_length=hop_length,
        n_mels=n_mels
    )
    
    # Handle different lengths by interpolating target envelope to match source
    target_frames = target_envelope.shape[1]
    source_frames = source_mel.shape[1]
    
    if target_frames != source_frames:
        # Interpolate target envelope to match source length
        target_interp = np.zeros((n_mels, source_frames))
        for i in range(n_mels):
            # Create interpolation function
            x_old = np.linspace(0, 1, target_frames)
            x_new = np.linspace(0, 1, source_frames)
            f = interpolate.interp1d(x_old, target_envelope[i, :], kind='linear', 
                                     bounds_error=False, fill_value='extrapolate')
            target_interp[i, :] = f(x_new)
        target_envelope = target_interp
    
    # Calculate ratio between target and source envelopes
    # Add small epsilon to avoid division by zero
    eps = 1e-10
    source_mel_db = librosa.power_to_db(source_mel + eps, ref=np.max)
    
    # Calculate gain to apply
    gain_db = target_envelope - source_mel_db
    gain_linear = librosa.db_to_power(gain_db)
    
    # Clamp gain to reasonable range to avoid artifacts
    gain_linear = np.clip(gain_linear, 0.1, 10.0)
    
    # Apply gain
    modified_mel = source_mel * gain_linear
    
    # Convert back to waveform using Griffin-Lim
    # First convert mel to linear spectrogram
    mel_basis = librosa.filters.mel(sr=sample_rate, n_fft=n_fft, n_mels=n_mels)
    linear_spec = np.dot(mel_basis.T, modified_mel)
    
    # Use Griffin-Lim to reconstruct audio
    audio_reconstructed = librosa.griffinlim(
        linear_spec,
        n_iter=32,
        hop_length=hop_length,
        n_fft=n_fft
    )
    
    return audio_reconstructed

def convert_voice_simple_conservative(source_audio_path, reference_audio_path, output_path, config):
    """
    Conservative voice conversion: pitch matching and loudness normalization only
    Avoids aggressive spectral manipulation that can introduce artifacts
    """
    try:
        sample_rate = config.get("sample_rate", 16000)
        
        # Load source audio (TTS output)
        source_audio, sr = librosa.load(source_audio_path, sr=sample_rate)
        
        # Load reference audio (original speaker)
        reference_audio, sr_ref = librosa.load(reference_audio_path, sr=sample_rate)
        
        # Extract a representative segment from reference (use middle portion, avoid silence)
        # Take middle 2-3 seconds if available
        ref_duration = len(reference_audio) / sample_rate
        if ref_duration > 3.0:
            start_sample = int(sample_rate * (ref_duration - 2.5) / 2)
            end_sample = int(sample_rate * (ref_duration + 2.5) / 2)
            reference_segment = reference_audio[start_sample:end_sample]
        else:
            reference_segment = reference_audio
        
        # Extract F0 from both using librosa (more stable than pyworld for this)
        f0_ref, voiced_flag_ref, _ = librosa.pyin(
            reference_segment,
            fmin=librosa.note_to_hz('C2'),
            fmax=librosa.note_to_hz('C7'),
            sr=sample_rate,
            frame_length=2048
        )
        f0_source, voiced_flag_source, _ = librosa.pyin(
            source_audio,
            fmin=librosa.note_to_hz('C2'),
            fmax=librosa.note_to_hz('C7'),
            sr=sample_rate,
            frame_length=2048
        )
        
        # Calculate pitch shift - use median instead of mean for robustness
        f0_ref_voiced = f0_ref[voiced_flag_ref]
        f0_source_voiced = f0_source[voiced_flag_source]
        
        pitch_shift_semitones = 0.0
        if len(f0_ref_voiced) > 10 and len(f0_source_voiced) > 10:
            # Use median for robustness against outliers
            median_f0_ref = np.median(f0_ref_voiced)
            median_f0_source = np.median(f0_source_voiced)
            
            if median_f0_source > 0:
                pitch_shift_semitones = 12 * np.log2(median_f0_ref / median_f0_source)
                # Clamp to reasonable range - be conservative
                pitch_shift_semitones = np.clip(pitch_shift_semitones, -3.0, 3.0)
        
        # Apply pitch shift if significant
        if abs(pitch_shift_semitones) > 0.2:
            converted_audio = librosa.effects.pitch_shift(
                source_audio,
                sr=sample_rate,
                n_steps=pitch_shift_semitones,
                bins_per_octave=12
            )
        else:
            converted_audio = source_audio.copy()
        
        # Normalize loudness to match reference
        rms_ref = np.sqrt(np.mean(reference_segment ** 2))
        rms_converted = np.sqrt(np.mean(converted_audio ** 2))
        if rms_converted > 1e-6:
            converted_audio = converted_audio * (rms_ref / rms_converted)
        
        # Prevent clipping
        max_val = np.max(np.abs(converted_audio))
        if max_val > 0.95:
            converted_audio = converted_audio * (0.95 / max_val)
        
        # Ensure output directory exists
        os.makedirs(os.path.dirname(output_path), exist_ok=True)
        
        # Save output
        sf.write(output_path, converted_audio, sample_rate)
        
        return {
            "success": True,
            "output_path": output_path,
            "pitch_shift_semitones": float(pitch_shift_semitones),
            "method": "conservative_pitch_loudness"
        }
        
    except Exception as e:
        import traceback
        error_msg = f"{str(e)}\n{traceback.format_exc()}"
        print(f"Error in convert_voice_simple_conservative: {error_msg}", file=sys.stderr)
        return {
            "success": False,
            "error": str(e),
            "output_path": None
        }

def convert_voice_simple(source_audio_path, target_speaker_embedding_path, output_path, config):
    """
    Voice conversion using spectral envelope transfer and pitch matching
    Uses reference audio from the embedding source to extract voice characteristics
    
    Args:
        source_audio_path: Path to source audio file (TTS output)
        target_speaker_embedding_path: Path to target speaker embedding
        output_path: Path for output audio file
        config: Configuration dict
    
    Returns:
        dict with success status and output path
    """
    try:
        sample_rate = config.get("sample_rate", 16000)
        reference_audio_path = config.get("reference_audio_path", None)
        
        # Load source audio (TTS output)
        source_audio, sr = librosa.load(source_audio_path, sr=sample_rate)
        
        # If we have a reference audio path, use it directly
        # Otherwise, try to infer from embedding path (embedding is usually saved near the audio)
        if not reference_audio_path:
            # Try to find reference audio near the embedding file
            embedding_dir = os.path.dirname(target_speaker_embedding_path)
            # Look for common audio file patterns
            for ext in ['.wav', '.mp3', '.flac']:
                potential_ref = target_speaker_embedding_path.replace('.npy', ext)
                if os.path.exists(potential_ref):
                    reference_audio_path = potential_ref
                    break
        
        if not reference_audio_path:
            return {
                "success": False,
                "error": "Reference audio path required for voice conversion. Set 'reference_audio_path' in config.",
                "output_path": None
            }
        
        # Use conservative approach: simple pitch + loudness matching
        return convert_voice_simple_conservative(source_audio_path, reference_audio_path, output_path, config)
        
        # OLD CODE - Disabled due to quality issues with Griffin-Lim reconstruction
        # Load reference audio (original speaker)
        reference_audio, sr_ref = librosa.load(reference_audio_path, sr=sample_rate)
        
        # Extract F0 from both
        # Use pyworld if available, otherwise librosa
        try:
            import pyworld as pw
            # Extract F0 from reference
            f0_ref, sp_ref, ap_ref = pw.wav2world(
                reference_audio.astype(np.float64), 
                sample_rate, 
                frame_period=1000 * 512 / sample_rate
            )
            # Extract F0 from source
            f0_source, sp_source, ap_source = pw.wav2world(
                source_audio.astype(np.float64),
                sample_rate,
                frame_period=1000 * 512 / sample_rate
            )
            
            # Calculate mean F0 shift
            f0_ref_voiced = f0_ref[f0_ref > 0]
            f0_source_voiced = f0_source[f0_source > 0]
            
            if len(f0_ref_voiced) > 0 and len(f0_source_voiced) > 0:
                mean_f0_ref = np.mean(f0_ref_voiced)
                mean_f0_source = np.mean(f0_source_voiced)
                
                # Calculate pitch shift in semitones
                if mean_f0_source > 0:
                    pitch_shift_semitones = 12 * np.log2(mean_f0_ref / mean_f0_source)
                    # Clamp to reasonable range
                    pitch_shift_semitones = np.clip(pitch_shift_semitones, -4.0, 4.0)
                else:
                    pitch_shift_semitones = 0.0
            else:
                pitch_shift_semitones = 0.0
            
            # Apply pitch shift
            if abs(pitch_shift_semitones) > 0.1:
                source_audio = librosa.effects.pitch_shift(
                    source_audio,
                    sr=sample_rate,
                    n_steps=pitch_shift_semitones
                )
            
            # Extract spectral envelope from reference
            reference_envelope = extract_spectral_envelope(reference_audio, sample_rate)
            
            # Apply spectral envelope to source
            converted_audio = apply_spectral_envelope(
                source_audio,
                reference_envelope,
                sample_rate
            )
            
        except ImportError:
            # Fallback: use librosa for F0 and simpler spectral matching
            print("Warning: pyworld not available, using librosa fallback", file=sys.stderr)
            
            # Extract F0 using librosa
            f0_ref, voiced_flag_ref, _ = librosa.pyin(
                reference_audio,
                fmin=librosa.note_to_hz('C2'),
                fmax=librosa.note_to_hz('C7'),
                sr=sample_rate
            )
            f0_source, voiced_flag_source, _ = librosa.pyin(
                source_audio,
                fmin=librosa.note_to_hz('C2'),
                fmax=librosa.note_to_hz('C7'),
                sr=sample_rate
            )
            
            # Calculate pitch shift
            f0_ref_voiced = f0_ref[voiced_flag_ref]
            f0_source_voiced = f0_source[voiced_flag_source]
            
            if len(f0_ref_voiced) > 0 and len(f0_source_voiced) > 0:
                mean_f0_ref = np.mean(f0_ref_voiced)
                mean_f0_source = np.mean(f0_source_voiced)
                
                if mean_f0_source > 0:
                    pitch_shift_semitones = 12 * np.log2(mean_f0_ref / mean_f0_source)
                    pitch_shift_semitones = np.clip(pitch_shift_semitones, -4.0, 4.0)
                else:
                    pitch_shift_semitones = 0.0
            else:
                pitch_shift_semitones = 0.0
            
            # Apply pitch shift
            if abs(pitch_shift_semitones) > 0.1:
                source_audio = librosa.effects.pitch_shift(
                    source_audio,
                    sr=sample_rate,
                    n_steps=pitch_shift_semitones
                )
            
            # Simple spectral envelope matching using mel spectrogram
            from scipy import interpolate
            
            reference_mel = librosa.feature.melspectrogram(
                y=reference_audio,
                sr=sample_rate,
                n_mels=80,
                hop_length=512
            )
            source_mel = librosa.feature.melspectrogram(
                y=source_audio,
                sr=sample_rate,
                n_mels=80,
                hop_length=512
            )
            
            # Handle different lengths by interpolating
            ref_frames = reference_mel.shape[1]
            src_frames = source_mel.shape[1]
            
            if ref_frames != src_frames:
                # Interpolate reference to match source length
                ref_interp = np.zeros((80, src_frames))
                for i in range(80):
                    x_old = np.linspace(0, 1, ref_frames)
                    x_new = np.linspace(0, 1, src_frames)
                    f = interpolate.interp1d(x_old, reference_mel[i, :], kind='linear',
                                           bounds_error=False, fill_value='extrapolate')
                    ref_interp[i, :] = f(x_new)
                reference_mel = ref_interp
            
            # Calculate gain
            eps = 1e-10
            gain = np.sqrt(reference_mel / (source_mel + eps))
            gain = np.clip(gain, 0.1, 10.0)  # Prevent extreme gains
            
            # Apply gain
            modified_mel = source_mel * gain
            
            # Reconstruct audio
            mel_basis = librosa.filters.mel(sr=sample_rate, n_fft=2048, n_mels=80)
            linear_spec = np.dot(mel_basis.T, modified_mel)
            converted_audio = librosa.griffinlim(linear_spec, n_iter=32, hop_length=512, n_fft=2048)
        
        # Normalize loudness to match reference
        rms_ref = np.sqrt(np.mean(reference_audio ** 2))
        rms_converted = np.sqrt(np.mean(converted_audio ** 2))
        if rms_converted > 0:
            converted_audio = converted_audio * (rms_ref / rms_converted)
        
        # Prevent clipping
        max_val = np.max(np.abs(converted_audio))
        if max_val > 0.95:
            converted_audio = converted_audio * (0.95 / max_val)
        
        # Ensure output directory exists
        os.makedirs(os.path.dirname(output_path), exist_ok=True)
        
        # Save output
        sf.write(output_path, converted_audio, sample_rate)
        
        return {
            "success": True,
            "output_path": output_path,
            "pitch_shift_semitones": float(pitch_shift_semitones) if 'pitch_shift_semitones' in locals() else 0.0,
            "method": "spectral_envelope_transfer"
        }
        
    except Exception as e:
        import traceback
        error_msg = f"{str(e)}\n{traceback.format_exc()}"
        print(f"Error in convert_voice_simple: {error_msg}", file=sys.stderr)
        return {
            "success": False,
            "error": str(e),
            "output_path": None
        }

def convert_voice_rvc(source_audio_path, target_speaker_embedding_path, rvc_model_path, output_path, config):
    """
    Voice conversion using trained RVC model
    
    Args:
        source_audio_path: Path to source audio file
        target_speaker_embedding_path: Path to target speaker embedding
        rvc_model_path: Path to trained RVC model checkpoint
        output_path: Path for output audio file
        config: Configuration dict
    
    Returns:
        dict with success status and output path
    """
    if not RVC_AVAILABLE:
        return {
            "success": False,
            "error": "RVC feature extractor not available",
            "output_path": None
        }
    
    try:
        sample_rate = config.get("sample_rate", 16000)
        f0_method = config.get("f0_method", "rmvpe")
        device = config.get("device", "auto")
        
        # Initialize feature extractor
        extractor = RVCFeatureExtractor(
            sample_rate=sample_rate,
            f0_method=f0_method,
            device=device
        )
        extractor.log_func = lambda msg: print(f"[RVC] {msg}", file=sys.stderr)
        
        # Load source audio
        source_audio, sr = librosa.load(source_audio_path, sr=sample_rate)
        
        # Extract features from source
        f0, voiced = extractor.extract_f0(source_audio)
        hubert_features = extractor.extract_hubert_features(source_audio)
        
        # Load target speaker embedding
        target_emb = np.load(target_speaker_embedding_path)
        
        # Load RVC model
        # TODO: Implement model loading and inference
        # This requires:
        # 1. Load model checkpoint
        # 2. Convert features to mel spectrogram
        # 3. Run model forward pass with target embedding
        # 4. Use vocoder to convert mel to audio
        
        return {
            "success": False,
            "error": "RVC model inference not yet implemented",
            "output_path": None
        }
        
    except Exception as e:
        return {
            "success": False,
            "error": str(e),
            "output_path": None
        }

def convert_voice(source_audio_path, target_speaker_embedding_path, output_path, config):
    """
    Convert source audio to target speaker's voice using RVC
    
    Args:
        source_audio_path: Path to source audio file
        target_speaker_embedding_path: Path to target speaker embedding
        output_path: Path for output audio file
        config: Configuration dict (model path, F0 method, etc.)
    
    Returns:
        dict with success status and output path
    """
    # Check if trained RVC model is available
    rvc_model_path = config.get("rvc_model_path", None)
    
    if rvc_model_path and os.path.exists(rvc_model_path):
        return convert_voice_rvc(
            source_audio_path, 
            target_speaker_embedding_path, 
            rvc_model_path,
            output_path, 
            config
        )
    else:
        # Use simple conversion as fallback
        return convert_voice_simple(
            source_audio_path,
            target_speaker_embedding_path,
            output_path,
            config
        )

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
            
            if mode == "convert":
                source = config.get("source")
                target_embedding = config.get("target_embedding")
                output = config.get("output")
                if not source or not target_embedding or not output:
                    result = {"success": False, "error": "Missing required arguments: source, target_embedding, output"}
                    print(json.dumps(result))
                    sys.stdout.flush()
                    continue
                result = convert_voice(source, target_embedding, output, config)
            elif mode == "extract":
                source = config.get("source")
                output = config.get("output")
                if not source or not output:
                    result = {"success": False, "error": "Missing required arguments: source, output"}
                    print(json.dumps(result))
                    sys.stdout.flush()
                    continue
                result = extract_speaker_embedding(source, output, config)
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
