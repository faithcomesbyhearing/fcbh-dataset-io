#!/usr/bin/env python3
"""
So-VITS-SVC Inference Wrapper
Handles voice conversion using trained So-VITS-SVC models

This wrapper expects So-VITS-SVC to be cloned separately and referenced via
the SO_VITS_SVC_ROOT environment variable.
"""

import os
import sys
import json
import numpy as np
import soundfile as sf
import torch

# Add parent directory to path for error handler
sys.path.insert(0, os.path.abspath(os.path.join(os.environ.get('GOPROJ', '.'), 'logger')))
try:
    from error_handler import setup_error_handler
    setup_error_handler()
except ImportError:
    pass  # Error handler not required

# Add So-VITS-SVC to Python path
SO_VITS_SVC_ROOT = os.environ.get('SO_VITS_SVC_ROOT')
if not SO_VITS_SVC_ROOT:
    print("Error: SO_VITS_SVC_ROOT environment variable not set", file=sys.stderr)
    print("Set it to the path where so-vits-svc is cloned, e.g.:", file=sys.stderr)
    print("  export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc", file=sys.stderr)
    sys.exit(1)

if not os.path.exists(SO_VITS_SVC_ROOT):
    print(f"Error: SO_VITS_SVC_ROOT path does not exist: {SO_VITS_SVC_ROOT}", file=sys.stderr)
    sys.exit(1)

# Add So-VITS-SVC to Python path
sys.path.insert(0, SO_VITS_SVC_ROOT)

# Change to SO_VITS_SVC_ROOT directory so relative paths work
os.chdir(SO_VITS_SVC_ROOT)

# Try to import So-VITS-SVC
SO_VITS_SVC_AVAILABLE = False
try:
    from inference.infer_tool import Svc
    SO_VITS_SVC_AVAILABLE = True
except ImportError as e:
    print(f"Error: Failed to import So-VITS-SVC from {SO_VITS_SVC_ROOT}: {e}", file=sys.stderr)
    print("Make sure So-VITS-SVC is properly installed and dependencies are met.", file=sys.stderr)
except Exception as e:
    print(f"Warning: So-VITS-SVC import failed: {e}", file=sys.stderr)

def convert_voice_so_vits_svc(source_audio_path, model_path, config_path, output_path, config):
    """
    Convert voice using So-VITS-SVC
    
    Args:
        source_audio_path: Path to source audio file
        model_path: Path to trained So-VITS-SVC model checkpoint (.pth file)
        config_path: Path to So-VITS-SVC config.json file
        output_path: Path for output audio file
        config: Configuration dict (speaker, f0_method, etc.)
    
    Returns:
        dict with success status and output path
    """
    if not SO_VITS_SVC_AVAILABLE:
        return {
            "success": False,
            "error": "So-VITS-SVC not available. Check SO_VITS_SVC_ROOT and dependencies.",
            "output_path": None
        }
    
    try:
        # Redirect stdout to stderr during model initialization to prevent JSON corruption
        import contextlib
        original_stdout = sys.stdout
        
        # Configuration parameters
        speaker = config.get("speaker", "speaker0")  # Speaker name from model
        f0_method = config.get("f0_method", "rmvpe")  # rmvpe, pm, harvest, crepe, dio, fcpe
        device = config.get("device", None)  # None = auto-detect
        # Convert "auto" to None for auto-detection
        if device == "auto":
            device = None
        # Auto-detect CUDA if device is None
        if device is None:
            device = "cuda" if torch.cuda.is_available() else "cpu"
        tran = config.get("tran", 0)  # Pitch shift in semitones
        auto_predict_f0 = config.get("auto_predict_f0", False)  # Auto F0 prediction
        noice_scale = config.get("noice_scale", 0.4)  # Noise scale
        pad_seconds = config.get("pad_seconds", 0.5)  # Padding
        slice_db = config.get("slice_db", -40)  # Slice threshold
        cluster_infer_ratio = config.get("cluster_infer_ratio", 0)  # Cluster ratio
        cluster_model_path = config.get("cluster_model_path", "")  # Optional cluster model
        
        # Initialize So-VITS-SVC model (redirect stdout to stderr to avoid JSON corruption)
        with contextlib.redirect_stdout(sys.stderr):
            svc_model = Svc(
            net_g_path=model_path,
            config_path=config_path,
            device=device,
            cluster_model_path=cluster_model_path,
            nsf_hifigan_enhance=config.get("enhance", False),
            diffusion_model_path=config.get("diffusion_model_path", ""),
            diffusion_config_path=config.get("diffusion_config_path", ""),
            shallow_diffusion=config.get("shallow_diffusion", False),
            only_diffusion=config.get("only_diffusion", False),
            spk_mix_enable=config.get("use_spk_mix", False),
            feature_retrieval=config.get("feature_retrieval", False)
            )
        
        # Perform inference (also redirect stdout during inference)
        with contextlib.redirect_stdout(sys.stderr):
            audio = svc_model.slice_inference(
            raw_audio_path=source_audio_path,
            spk=speaker,
            tran=tran,
            slice_db=slice_db,
            cluster_infer_ratio=cluster_infer_ratio,
            auto_predict_f0=auto_predict_f0,
            noice_scale=noice_scale,
            pad_seconds=pad_seconds,
            clip_seconds=config.get("clip_seconds", 0),
            lg_num=config.get("linear_gradient", 0),
            lgr_num=config.get("linear_gradient_retain", 0.75),
            f0_predictor=f0_method,
            enhancer_adaptive_key=config.get("enhancer_adaptive_key", 0),
            cr_threshold=config.get("f0_filter_threshold", 0.05),
            k_step=config.get("k_step", 100),
            use_spk_mix=config.get("use_spk_mix", False),
            second_encoding=config.get("second_encoding", False),
            loudness_envelope_adjustment=config.get("loudness_envelope_adjustment", 1)
            )
        
        # Save output
        wav_format = config.get("wav_format", "wav")
        sf.write(output_path, audio, svc_model.target_sample, format=wav_format)
        
        # Clean up
        with contextlib.redirect_stdout(sys.stderr):
            svc_model.clear_empty()
        
        return {
            "success": True,
            "output_path": output_path,
            "sample_rate": svc_model.target_sample
        }
        
    except Exception as e:
        import traceback
        error_msg = f"{str(e)}\n{traceback.format_exc()}"
        print(f"Error in convert_voice_so_vits_svc: {error_msg}", file=sys.stderr)
        return {
            "success": False,
            "error": str(e),
            "output_path": None
        }

def main():
    # Redirect stdout to stderr to prevent interference with JSON responses
    # So-VITS-SVC prints messages to stdout which would break JSON parsing
    import contextlib
    original_stdout = sys.stdout
    
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
                model_path = config.get("model_path")
                config_path = config.get("config_path")
                output = config.get("output")
                if not source or not model_path or not config_path or not output:
                    result = {"success": False, "error": "Missing required arguments: source, model_path, config_path, output"}
                    print(json.dumps(result))
                    sys.stdout.flush()
                    continue
                result = convert_voice_so_vits_svc(source, model_path, config_path, output, config)
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

