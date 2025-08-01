https://docs.pytorch.org/audio/stable/tutorials/forced_alignment_tutorial.html


Perfect! That PyTorch tutorial is exactly what you need. The `torchaudio.functional.forced_align` 
function implements the CTC forward-backward algorithm for forced alignment, which is precisely 
what we were discussing.

## Key Advantages for Your Use Case

This approach will work well with your MMS adapter because:

1. **Model agnostic**: It works with any CTC-based model that outputs frame-level token probabilities
1. **Handles misalignments**: Built-in support for insertions, deletions, and the alignment issues you mentioned
1. **Word-level scoring**: You can aggregate the character/token-level alignments to get word probabilities

## Integration with Your MMS Adapter

You’ll need to:

1. **Extract emissions**: Get the CTC logits from your trained MMS model

```python
emissions = your_mms_model(waveform)  # Shape: [batch, time, vocab]
log_probs = torch.nn.functional.log_softmax(emissions, dim=-1)
```

1. **Tokenize your script**: Convert to the same vocabulary your MMS adapter uses

```python
tokens = tokenize_script_text(script, your_mms_tokenizer)
```

1. **Run forced alignment**:

```python
alignment = torchaudio.functional.forced_align(
    log_probs, targets=tokens, input_lengths, target_lengths
)
```

## Getting Word-Level Probabilities

The tutorial shows how to extract timing and confidence scores. For your validation use case, you can 
aggregate the token-level scores within each word boundary to get the word-level probabilities you want.

This is much cleaner than implementing CTC forward-backward from scratch! The tutorial should give 
you everything you need to build your audio-script validation system on top of your MMS adapters
