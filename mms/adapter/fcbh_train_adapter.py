#import os
#import numpy as np
#import torch
#import torch.nn as nn
#from torch.utils.data import Dataset, DataLoader
#import torchaudio
#import pandas as pd
#from transformers import Wav2Vec2ForCTC, Wav2Vec2Config, Wav2Vec2Processor
#from transformers.adapters import AdapterConfig, PfeifferConfig

MMS_ADAPTERS_DIR = os.path(os.getenv("FCBH_DATASET_DB"), "mms_adapters")
# batch size is recommended to be 4 to 8, try 8 to see if there is enough memory
# num workers is recommended to be 2 to 4, start with 2
# try 30-50 epochs, for overlearning do 60-100 epochs
# Consider data augmentation techniques (speed perturbation, noise addition)
# Use a smaller learning rate (1e-5 instead of 1e-4) for more stable training
# Implement learning rate scheduling (reduce on plateau)


if len(sys.argv) < 7:
    print("Usage: fcbh_train_adapter.py {iso639-3} {vocabSize} {databasePath} {audioDirectory} {batchSize} {numWorkers}")
    sys.exit(1)
adapterName = sys.argv[1]
vocabularySize = sys.argv[2]
databasePath = sys.argv[3]
audioDirectory = sys.argv[4]
batchSize = sys.argv[5]
numWorkers = sys.argv[6]

# Load MMS model and processor
modelName = "facebook/mms-1b-all"
processor = Wav2Vec2Processor.from_pretrained(modelName)
model = Wav2Vec2ForCTC.from_pretrained(modelName)

# Setup adapter for new language
adapterConfig = PfeifferConfig(reduction_factor=16)
model.add_adapter(adapterName, config=adapterConfig)
model.set_active_adapters(adapterName)

# Freeze base model parameters and only train the adapter
for param in model.base_model.parameters():
    param.requires_grad = False

# Ensure output layer is correctly sized for target language
model.resize_output_embeddings(vocabularySize)

# Set up train and test datasets
wav2Vec2Processor = Wav2Vec2Processor.from_pretrained(modelName)
data = FCBHDataset(databasePath, audioDirectory, wav2Vec2Processor)
kFold = 5
trainDataset, testDataset = FCBHDataLoader(data, kFold, batchSize, numWorkers)

# Setup device, optimizer
device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
model.to(device)
optimizer = torch.optim.AdamW(model.parameters(), lr=1e-4)

# Train the adapter
trainMMSAdapter(model, trainDataset, optimizer, device)

model.save_adapter(MMS_ADAPTERS_DIR)

    # For evaluation
    # Implement an evaluation function as needed





def trainMMSAdapter(model, dataloader, optimizer, device, epochs=5):
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

