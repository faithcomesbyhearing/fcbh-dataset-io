import torch
from transformers import Wav2Vec2ForCTC, Wav2Vec2Processor

def get_word_timestamps_from_ctc(audio_path, model, processor, blank_id=0):
    """
    Get word-level timestamps from CTC output without forced alignment
    """
    import librosa

    # Load audio
    audio, sr = librosa.load(audio_path, sr=16000)
    inputs = processor(audio, sampling_rate=16000, return_tensors="pt", padding=True)

    # Get CTC outputs
    with torch.no_grad():
        logits = model(inputs.input_values).logits

    # Get predicted token IDs (greedy decoding)
    predicted_ids = torch.argmax(logits, dim=-1)[0]

    # CTC collapse: remove blanks and repeated tokens
    frame_duration = 0.02  # 20ms per frame for wav2vec2

    words = []
    current_word_tokens = []
    current_word_start = None
    prev_token = blank_id

    for frame_idx, token_id in enumerate(predicted_ids.tolist()):
        if token_id == blank_id:
            # Blank token - potential word boundary
            if current_word_tokens:
                # End of word
                word_text = processor.decode(current_word_tokens)
                if word_text.strip():
                    words.append({
                        'word': word_text.strip(),
                        'start': current_word_start * frame_duration,
                        'end': frame_idx * frame_duration
                    })
                current_word_tokens = []
                current_word_start = None
            prev_token = token_id

        elif token_id != prev_token:
            # New token (CTC collapse rule)
            if current_word_start is None:
                current_word_start = frame_idx
            current_word_tokens.append(token_id)
            prev_token = token_id

    # Handle final word
    if current_word_tokens:
        word_text = processor.decode(current_word_tokens)
        if word_text.strip():
            words.append({
                'word': word_text.strip(),
                'start': current_word_start * frame_duration,
                'end': len(predicted_ids) * frame_duration
            })

    return words

# Usage
words_with_timestamps = get_word_timestamps_from_ctc(
    "audio.wav",
    your_trained_model,
    your_processor
)

for w in words_with_timestamps:
    print(f"{w['word']}: {w['start']:.2f}s - {w['end']:.2f}s")