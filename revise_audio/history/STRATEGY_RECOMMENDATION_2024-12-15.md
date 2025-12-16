# Strategy Recommendation: FCBH RVC vs So-VITS-SVC

**Date**: 2024-12-15  
**Context**: So-VITS-SVC tested on AWS, quality unacceptable. Considering pivot to FCBH RVC strategy from TOOLING_RESEARCH.md

---

## Current Situation

### What We Know

**So-VITS-SVC Results (from AWS testing)**:
- ❌ **Quality**: Unacceptable with 26 segments (~4 min training data)
- ❌ **Voice Conversion**: Introduces artifacts, quality degradation
- ❌ **Training Data**: Insufficient (need 30+ minutes, have ~4 minutes)
- ✅ **Infrastructure**: Integration complete, preprocessing works

**TOOLING_RESEARCH.md Original Recommendation**:
- **Phase 1**: Use FCBH RVC (existing codebase) + DSP prosody matching
- **Phase 2**: Benchmark FCBH RVC vs So-VITS-SVC
- **Rationale**: Pragmatic - use what exists, validate later

---

## Analysis: Should We Pivot to FCBH RVC?

### Arguments FOR Pivoting to FCBH RVC

1. **Already Tested So-VITS-SVC**: We now have empirical evidence it doesn't work well with limited data
   - ✅ Validates TOOLING_RESEARCH.md's Phase 2 recommendation (benchmark)
   - ✅ We've done the comparison, So-VITS-SVC failed

2. **FCBH RVC May Be Better for Low-Resource**:
   - Claims "optimized for low-resource" (unverified, but worth testing)
   - May handle limited training data better than So-VITS-SVC
   - Already in codebase, designed for this use case

3. **Pragmatic Approach**:
   - Code already exists in FCBH-W2V-Bert-2.0-ASR-trainer
   - Faster to test than collecting 30+ minutes of training data
   - Can validate if it's actually better before investing more time

4. **Lessons Learned Support This**:
   - AWS testing showed insufficient data is a fundamental issue
   - FCBH RVC might be more tolerant of limited data
   - Worth trying before collecting more training data

### Arguments AGAINST Pivoting

1. **So-VITS-SVC Infrastructure Complete**:
   - Integration code done
   - Preprocessing pipeline works
   - Would need to duplicate effort for FCBH RVC

2. **Unknown if FCBH RVC is Better**:
   - "Optimized for low-resource" is an unverified claim
   - May have same issues with limited data
   - Could be wasted effort

3. **Training Data is Still the Issue**:
   - Both systems likely need more data
   - FCBH RVC might not solve the fundamental problem
   - May be better to collect more data first

4. **Time Investment**:
   - Need to port/adapt FCBH RVC code
   - Need to set up training pipeline
   - May take similar time as collecting more training data

---

## Recommendation: **YES, Pivot to FCBH RVC** (with caveats)

### Rationale

1. **Empirical Evidence**: We've tested So-VITS-SVC and it failed. This is exactly what TOOLING_RESEARCH.md recommended (Phase 2: benchmark). We've done the benchmark, So-VITS-SVC lost.

2. **Low Risk, High Reward**:
   - FCBH RVC code already exists
   - Can test quickly (few days of work)
   - If it works better, we've solved the problem
   - If it doesn't, we've validated the data issue is fundamental

3. **Matches Original Strategy**: TOOLING_RESEARCH.md recommended FCBH RVC as Phase 1, with So-VITS-SVC as comparison. We did it backwards (So-VITS-SVC first), but now we can do it right.

4. **Data Collection Can Happen in Parallel**:
   - Test FCBH RVC with current data
   - Collect more training data in parallel
   - If FCBH RVC works, great. If not, we'll have more data for retraining.

### Implementation Strategy

**Phase 1: Quick FCBH RVC Test** (1-2 days)
1. Port/adapt FCBH RVC inference code
2. Test with same 26 segments (or extract more from corpus)
3. Compare quality to So-VITS-SVC results
4. **Decision Point**: If better, proceed. If same/worse, pivot to data collection.

**Phase 2: Data Collection** (ongoing, in parallel)
1. Extract more training data from corpus
2. Target: 30+ minutes of clean, diverse speech
3. Can be used for either system

**Phase 3: Optimize Based on Results**
- If FCBH RVC works: Continue with it, collect more data for fine-tuning
- If FCBH RVC fails: Collect more data, retrain So-VITS-SVC, or consider alternatives

### What to Keep from So-VITS-SVC Work

✅ **Keep**:
- Integration patterns (Go → Python subprocess)
- Preprocessing pipeline concepts
- Test infrastructure
- Documentation approach

❌ **Don't Keep**:
- So-VITS-SVC-specific code (can be archived)
- Training pipeline (if FCBH RVC works better)

---

## Alternative: Hybrid Approach

**Option**: Use FCBH RVC for voice conversion, keep So-VITS-SVC infrastructure for future

1. **Short Term**: Implement FCBH RVC, test with current data
2. **Medium Term**: Collect more training data
3. **Long Term**: Retrain So-VITS-SVC with more data, compare again

This gives us:
- Immediate solution (FCBH RVC)
- Future option (So-VITS-SVC with more data)
- Empirical comparison data

---

## Risk Assessment

### Low Risk
- ✅ FCBH RVC code exists, just needs porting
- ✅ Can test quickly without major investment
- ✅ So-VITS-SVC work isn't wasted (patterns reusable)

### Medium Risk
- ⚠️ FCBH RVC may have same data limitations
- ⚠️ Porting may reveal unexpected issues
- ⚠️ May need to collect more data anyway

### Mitigation
- Test FCBH RVC quickly (1-2 days max)
- Set clear success criteria (must be better than So-VITS-SVC)
- Collect training data in parallel
- Keep So-VITS-SVC code for future use

---

## Success Criteria

**FCBH RVC should be considered successful if**:
1. Quality is better than So-VITS-SVC (subjective, but clear improvement)
2. Works with current training data (26 segments)
3. Integration is straightforward (similar effort to So-VITS-SVC)

**If FCBH RVC fails**:
- Collect more training data (30+ minutes)
- Retrain So-VITS-SVC with more data
- Consider alternative approaches (corpus snippets, TTS-only)

---

## Conclusion

**Recommendation**: **Pivot to FCBH RVC** following TOOLING_RESEARCH.md strategy.

**Why**:
1. We've validated So-VITS-SVC doesn't work with limited data
2. FCBH RVC is the original recommendation
3. Low risk, quick test
4. Matches pragmatic approach from TOOLING_RESEARCH.md

**Next Steps**:
1. Port FCBH RVC inference code to `revise_audio/`
2. Test with same training data (26 segments)
3. Compare quality to So-VITS-SVC results
4. Make go/no-go decision based on results
5. Collect more training data in parallel

**Timeline**: 1-2 days for FCBH RVC test, then decide on path forward.

---

## References

- TOOLING_RESEARCH.md: Original strategy recommendation
- LESSONS_LEARNED_2024-12-15.md: AWS testing results
- SESSION_SUMMARY_2024-12-15_TTS_VITS_PROSOODY_EVALUATION.md: Detailed test results
- FCBH RVC Code: `/Users/jrstear/git/FCBH-W2V-Bert-2.0-ASR-trainer/`

