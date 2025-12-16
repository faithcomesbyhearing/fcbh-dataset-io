#!/usr/bin/env python3
"""
MMS TTS (Text-to-Speech) script for generating audio from text.
Communicates via stdin/stdout with JSON messages.
"""
import json
import sys
import os
import torch
from transformers import VitsTokenizer, VitsModel, set_seed
import scipy.io.wavfile
import tempfile

def main():
    # Read language code from command line
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Language code required (e.g., 'eng')"}), file=sys.stderr)
        sys.exit(1)
    
    lang = sys.argv[1]
    model_id = f"facebook/mms-tts-{lang}"
    
    # Initialize model (lazy loading - only load when first request comes)
    tokenizer = None
    model = None
    device = "cuda" if torch.cuda.is_available() else "cpu"
    
    # Process requests from stdin
    for line in sys.stdin:
        try:
            request = json.loads(line.strip())
            text = request.get("text", "")
            output_path = request.get("output_path", "")
            seed = request.get("seed", 555)  # Default seed for reproducibility
            
            if not text:
                response = {"error": "Text is required"}
                print(json.dumps(response))
                sys.stdout.flush()
                continue
            
            # Lazy load model on first request
            if tokenizer is None or model is None:
                try:
                    tokenizer = VitsTokenizer.from_pretrained(model_id)
                    model = VitsModel.from_pretrained(model_id)
                    model = model.to(device)
                except Exception as e:
                    response = {"error": f"Failed to load model {model_id}: {str(e)}"}
                    print(json.dumps(response))
                    sys.stdout.flush()
                    continue
            
            # Generate audio
            try:
                inputs = tokenizer(text=text, return_tensors="pt")
                inputs = {k: v.to(device) for k, v in inputs.items()}
                set_seed(seed)  # Make deterministic
                
                with torch.no_grad():
                    outputs = model(**inputs)
                
                waveform = outputs.waveform[0].cpu().numpy()
                sample_rate = model.config.sampling_rate
                
                # Write to file (use scipy if available, otherwise soundfile)
                if output_path:
                    final_path = output_path
                else:
                    # Create temp file
                    with tempfile.NamedTemporaryFile(mode='wb', suffix='.wav', delete=False) as f:
                        final_path = f.name
                
                try:
                    scipy.io.wavfile.write(final_path, sample_rate, waveform)
                except NameError:
                    # Fallback to soundfile
                    sf.write(final_path, waveform, sample_rate)
                
                response = {
                    "success": True,
                    "output_path": final_path,
                    "sample_rate": sample_rate,
                    "duration": len(waveform) / sample_rate
                }
                
                print(json.dumps(response))
                sys.stdout.flush()
                
            except Exception as e:
                response = {"error": f"Generation failed: {str(e)}"}
                print(json.dumps(response))
                sys.stdout.flush()
                
        except json.JSONDecodeError as e:
            response = {"error": f"Invalid JSON: {str(e)}"}
            print(json.dumps(response))
            sys.stdout.flush()
        except Exception as e:
            response = {"error": f"Unexpected error: {str(e)}"}
            print(json.dumps(response))
            sys.stdout.flush()

if __name__ == "__main__":
    main()

