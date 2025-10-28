def check_audio_quality(audio_path, reference_text, model, processor):
    """
    Compare spoken audio to reference text to find errors
    """
    # Step 1: Get what was ACTUALLY spoken (with timestamps)
    spoken_words = get_word_timestamps_from_ctc(audio_path, model, processor)

    # Step 2: Align spoken words to reference text
    reference_words = reference_text.split()

    # Step 3: Find differences using sequence alignment
    from difflib import SequenceMatcher

    spoken_text = [w['word'] for w in spoken_words]
    matcher = SequenceMatcher(None, reference_words, spoken_text)

    errors = []

    for tag, i1, i2, j1, j2 in matcher.get_opcodes():
        if tag == 'replace':
            # Wrong word(s) spoken
            errors.append({
                'type': 'substitution',
                'expected': ' '.join(reference_words[i1:i2]),
                'actual': ' '.join(spoken_text[j1:j2]),
                'timestamp': spoken_words[j1]['start'] if j1 < len(spoken_words) else None
            })
        elif tag == 'delete':
            # Word(s) missing from audio
            errors.append({
                'type': 'deletion',
                'expected': ' '.join(reference_words[i1:i2]),
                'actual': '[missing]',
                'timestamp': None
            })
        elif tag == 'insert':
            # Extra word(s) in audio
            errors.append({
                'type': 'insertion',
                'expected': '[none]',
                'actual': ' '.join(spoken_text[j1:j2]),
                'timestamp': spoken_words[j1]['start']
            })

    return errors, spoken_words

# Usage
errors, all_words = check_audio_quality(
    "chapter1.wav",
    "In the beginning God created the heaven and the earth",
    your_model,
    your_processor
)

for error in errors:
    print(f"{error['type']} at {error['timestamp']}s:")
    print(f"  Expected: {error['expected']}")
    print(f"  Got: {error['actual']}")