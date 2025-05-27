import torch
import torch.nn as nn
from torch.utils.data import Dataset, DataLoader
import torchaudio
import pandas as pd
import os
from transformers import Wav2Vec2ForCTC, Wav2Vec2Config, Wav2Vec2Processor
import numpy as np
from transformers.adapters import AdapterConfig, PfeifferConfig

# Define the custom dataset for the low-resource language
class SpeechDataset(Dataset):
    def __init__(self, csv_file, audio_dir, processor, max_len=160000):
        """
        Args:
            csv_file (string): Path to the CSV file with annotations.
            audio_dir (string): Directory with all the audio files.
            processor (Wav2Vec2Processor): Processor for the audio.
            max_len (int): Maximum audio length in samples.
        """
        self.transcription_df = pd.read_csv(csv_file)
        self.audio_dir = audio_dir
        self.processor = processor
        self.max_len = max_len

    def __len__(self):
        return len(self.transcription_df)

    def __getitem__(self, idx):
        # Get audio file path and corresponding text
        audio_path = os.path.join(self.audio_dir, self.transcription_df.iloc[idx, 0])
        text = self.transcription_df.iloc[idx, 1]

        # Load and preprocess audio
        speech, sample_rate = torchaudio.load(audio_path)
        speech = speech.squeeze().numpy()

        # Resample if needed
        if sample_rate != 16000:
            resampler = torchaudio.transforms.Resample(sample_rate, 16000)
            speech = resampler(torch.tensor(speech)).numpy()
            sample_rate = 16000

        # Pad or trim audio
        if len(speech) < self.max_len:
            # Pad
            padded_speech = np.zeros(self.max_len)
            padded_speech[:len(speech)] = speech
            speech = padded_speech
        else:
            # Trim
            speech = speech[:self.max_len]

        # Process audio
        input_values = self.processor(speech, sampling_rate=16000, return_tensors="pt").input_values.squeeze()

        # Process text
        with self.processor.as_target_processor():
            labels = self.processor(text, return_tensors="pt").input_ids.squeeze()

        return {
            "input_values": input_values,
            "labels": labels,
            "text": text
        }

# Define training function
def train_mms_adapter(model, dataloader, optimizer, device, epochs=5):
    model.train()

    for epoch in range(epochs):
        total_loss = 0
        for batch in dataloader:
            # Move batch to device
            input_values = batch["input_values"].to(device)
            labels = batch["labels"].to(device)

            # Forward pass
            outputs = model(input_values=input_values, labels=labels)
            loss = outputs.loss

            # Backward pass
            optimizer.zero_grad()
            loss.backward()
            optimizer.step()

            total_loss += loss.item()

        print(f"Epoch {epoch+1}/{epochs}, Loss: {total_loss/len(dataloader)}")

# Main execution
def main():
    # Load MMS model and processor
    model_name = "facebook/mms-1b-all"
    processor = Wav2Vec2Processor.from_pretrained(model_name)
    model = Wav2Vec2ForCTC.from_pretrained(model_name)

    # Setup adapter for new language
    adapter_name = "new_language_adapter"
    adapter_config = PfeifferConfig(reduction_factor=16)

    # Add adapter to model
    model.add_adapter(adapter_name, config=adapter_config)

    # Activate adapter
    model.set_active_adapters(adapter_name)

    # Freeze base model parameters and only train the adapter
    for param in model.base_model.parameters():
        param.requires_grad = False

    # Ensure output layer is correctly sized for target language
    # Assuming you have already determined the vocabulary size of your target language
    target_vocab_size = 100  # Replace with actual vocab size
    model.resize_output_embeddings(target_vocab_size)

    # Create dataset and dataloader
    train_dataset = SpeechDataset(
        csv_file="path/to/train.csv",
        audio_dir="path/to/audio",
        processor=processor
    )

    train_dataloader = DataLoader(
        train_dataset,
        batch_size=4,
        shuffle=True,
        collate_fn=lambda batch: {
            "input_values": torch.stack([item["input_values"] for item in batch]),
            "labels": torch.stack([item["labels"] for item in batch]),
            "text": [item["text"] for item in batch]
        }
    )

    # Setup device, optimizer
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    model.to(device)

    optimizer = torch.optim.AdamW(model.parameters(), lr=1e-4)

    # Train the adapter
    train_mms_adapter(model, train_dataloader, optimizer, device)

    # Save the adapter
    model.save_adapter("./saved_adapters/", adapter_name)

    # For evaluation
    # Implement an evaluation function as needed

if __name__ == "__main__":
    main()