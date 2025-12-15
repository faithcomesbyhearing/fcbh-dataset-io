# Next Steps: So-VITS-SVC Integration

**Created**: 2024-12-14  
**Status**: Ready to begin  
**Bead**: arti-2ps

---

## Overview

Integrate So-VITS-SVC (mature open-source RVC implementation) for neural voice conversion to replace the current DSP-based approach that produces intelligible but unacceptable results.

---

## Implementation Plan

### Phase 1: Research and Setup

1. **Research So-VITS-SVC**:
   - Review GitHub repository: https://github.com/svc-develop-team/so-vits-svc
   - Understand installation requirements
   - Review documentation and examples
   - Identify inference API/interface

2. **Environment Setup**:
   - Add So-VITS-SVC to conda environment (`revise_audio/build_env.sh`)
   - Install dependencies
   - Test basic installation

### Phase 2: Training Pipeline

3. **Create Training Data Pipeline**:
   - Extract audio segments for target speaker from corpus
   - Prepare training data format (WAV files, 16kHz, mono)
   - Create training script/configuration
   - Document training process

4. **Train Initial Model**:
   - Train on Jude narrator corpus (test case)
   - Validate training completes successfully
   - Evaluate model quality

### Phase 3: Integration

5. **Integrate Inference**:
   - Create So-VITS-SVC inference wrapper (Python)
   - Update `voice_conversion.py` to use So-VITS-SVC
   - Maintain Go adapter interface (no changes needed)
   - Test inference pipeline

6. **Quality Validation**:
   - Test voice conversion on TTS-generated segments
   - Compare quality to DSP approach
   - Iterate on training parameters if needed

### Phase 4: Re-enable Prosody Matching

7. **Re-enable Prosody Matching**:
   - Once voice conversion quality is acceptable
   - Fine-tune prosody matching parameters
   - Test combined pipeline (TTS → VC → Prosody → Stitch)

---

## Key Files to Modify

- `revise_audio/python/voice_conversion.py` - Replace DSP approach with So-VITS-SVC
- `revise_audio/build_env.sh` - Add So-VITS-SVC dependencies
- `revise_audio/vc_adapter.go` - May need minor updates for model path config

## New Files to Create

- `revise_audio/python/so_vits_svc_inference.py` - So-VITS-SVC inference wrapper
- `revise_audio/python/train_so_vits_svc.py` - Training script (optional, may use CLI)
- `revise_audio/history/SO_VITS_SVC_INTEGRATION.md` - Integration notes

---

## Success Criteria

- Voice conversion produces natural-sounding speech (not robotic)
- Quality matches or exceeds original speaker characteristics
- Integration maintains existing Go-Python interface
- Training pipeline is documented and repeatable
- End-to-end workflow (TTS → VC → Prosody → Stitch) produces acceptable results

---

## References

- So-VITS-SVC GitHub: https://github.com/svc-develop-team/so-vits-svc
- RVC Explanation: `revise_audio/history/RVC_EXPLANATION.md`
- Alternatives Analysis: `revise_audio/history/VOICE_CONVERSION_ALTERNATIVES.md`
- Session Summary: `revise_audio/history/SESSION_SUMMARY_2024-12-14.md`

