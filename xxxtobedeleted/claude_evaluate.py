

def evaluate_model(model, dataloader, processor, device):
    model.eval()

    # Track metrics
    total_wer = 0
    total_cer = 0
    total_samples = 0

    with torch.no_grad():
        for batch in dataloader:
            # Move batch to device
            input_values = batch["input_values"].to(device)
            labels = batch["labels"]
            texts = batch["text"]

            # Get model predictions
            outputs = model(input_values=input_values)
            logits = outputs.logits

            # Decode predictions
            predicted_ids = torch.argmax(logits, dim=-1)
            predicted_texts = processor.batch_decode(predicted_ids)

            # Calculate metrics
            for pred_text, ref_text in zip(predicted_texts, texts):
                # Clean up text (remove special tokens, etc.)
                pred_text = pred_text.lower().strip()
                ref_text = ref_text.lower().strip()

                # Calculate Word Error Rate
                wer = calculate_wer(pred_text, ref_text)
                total_wer += wer

                # Calculate Character Error Rate
                cer = calculate_cer(pred_text, ref_text)
                total_cer += cer

                total_samples += 1

                # Print a few examples
                if total_samples <= 5:  # Just show first 5 examples
                    print(f"Reference: {ref_text}")
                    print(f"Prediction: {pred_text}")
                    print(f"WER: {wer:.4f}, CER: {cer:.4f}")
                    print("---")

    # Print overall results
    avg_wer = total_wer / total_samples
    avg_cer = total_cer / total_samples
    print(f"Average WER: {avg_wer:.4f}")
    print(f"Average CER: {avg_cer:.4f}")

    return avg_wer, avg_cer

# Helper functions for metrics
def calculate_wer(predicted_text, reference_text):
    # Split into words
    pred_words = predicted_text.split()
    ref_words = reference_text.split()

    # Calculate edit distance
    distance = levenshtein_distance(pred_words, ref_words)

    # Calculate WER
    if len(ref_words) > 0:
        return distance / len(ref_words)
    else:
        return 0 if len(pred_words) == 0 else 1

def calculate_cer(predicted_text, reference_text):
    # Convert to list of characters
    pred_chars = list(predicted_text)
    ref_chars = list(reference_text)

    # Calculate edit distance
    distance = levenshtein_distance(pred_chars, ref_chars)

    # Calculate CER
    if len(ref_chars) > 0:
        return distance / len(ref_chars)
    else:
        return 0 if len(pred_chars) == 0 else 1

def levenshtein_distance(seq1, seq2):
    """Calculate the Levenshtein distance between two sequences."""
    size_x = len(seq1) + 1
    size_y = len(seq2) + 1
    matrix = np.zeros((size_x, size_y))

    for x in range(size_x):
        matrix[x, 0] = x
    for y in range(size_y):
        matrix[0, y] = y

    for x in range(1, size_x):
        for y in range(1, size_y):
            if seq1[x-1] == seq2[y-1]:
                matrix[x, y] = matrix[x-1, y-1]
            else:
                matrix[x, y] = min(
                    matrix[x-1, y] + 1,     # deletion
                    matrix[x, y-1] + 1,     # insertion
                    matrix[x-1, y-1] + 1    # substitution
                )

    return matrix[size_x-1, size_y-1]