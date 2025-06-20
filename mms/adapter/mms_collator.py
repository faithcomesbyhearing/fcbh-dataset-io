import torch
from typing import Dict, List

class SimpleMMSCollator:

    def __init__(self, processor):
        self.processor = processor

    def __call__(self, features: List[Dict[str, torch.Tensor]]) -> Dict[str, torch.Tensor]:
        # Extract sequences
        input_values = [f["input_values"] for f in features]
        labels = [f["labels"] for f in features]

        # Pad audio to max length in batch
        max_audio_len = max(seq.shape[0] for seq in input_values)
        batch_input_values = torch.zeros(len(input_values), max_audio_len)

        for i, seq in enumerate(input_values):
            batch_input_values[i, :seq.shape[0]] = seq

        # Pad labels to max length in batch (use -100 for padding)
        max_label_len = max(seq.shape[0] for seq in labels)
        batch_labels = torch.full((len(labels), max_label_len), -100, dtype=torch.long)

        for i, seq in enumerate(labels):
            batch_labels[i, :seq.shape[0]] = seq

        return {
            "input_values": batch_input_values,
            "labels": batch_labels,
        }