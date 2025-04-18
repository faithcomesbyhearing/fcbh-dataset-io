#import os
#import numpy as np
#import torch
#import torch.nn as nn
#from torch.utils.data import Dataset, DataLoader
#import torchaudio
#import pandas as pd
#from transformers import Wav2Vec2ForCTC, Wav2Vec2Config, Wav2Vec2Processor
#from transformers.adapters import AdapterConfig, PfeifferConfig
from torch.optim.lr_scheduler import CosineAnnealingLR

MMS_ADAPTERS_DIR = os.path(os.getenv("FCBH_DATASET_DB"), "mms_adapters")
# batch size is recommended to be 4 to 8, try 8 to see if there is enough memory
# num workers is recommended to be 2 to 4, start with 2
# try 30-50 epochs, for overlearning do 60-100 epochs
# Consider data augmentation techniques (speed perturbation, noise addition)
# Use a smaller learning rate (1e-5 instead of 1e-4) for more stable training
# Implement learning rate scheduling (reduce on plateau)


if len(sys.argv) < 7:
    print("Usage: fcbh_train_adapter.py {iso639-3} {databasePath} {audioDirectory} {batchSize} {numWorkers} {numEpochs}")
    sys.exit(1)
adapterName = sys.argv[1]
databasePath = sys.argv[2]
audioDirectory = sys.argv[3]
batchSize = sys.argv[4]
numWorkers = sys.argv[5]
numEpochs = sys.argv[6]

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
model.resize_output_embeddings(dataset.getVocabularySize())

# Set up train and test datasets
wav2Vec2Processor = Wav2Vec2Processor.from_pretrained(modelName)
dataset = FCBHDataset(databasePath, audioDirectory, wav2Vec2Processor)
#kFold = 5
dataLoader = FCBHDataLoader(dataset, "train", batchSize, numWorkers)

# Setup device, optimizer
device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
model.to(device)
#optimizer = torch.optim.AdamW(model.parameters(), lr=1e-4)

# Train the adapter
#trainMMSAdapter(model, trainDataset, optimizer, device)

#model.save_adapter(MMS_ADAPTERS_DIR)

    # For evaluation
    # Implement an evaluation function as needed
optimizer = torch.optim.AdamW(model.parameters(), lr=2e-5)
#num_epochs = 80  # More epochs for memorization
scheduler = CosineAnnealingLR(optimizer, T_max=numEpochs, eta_min=1e-6)

# Training loop
bestLoss = float('inf')
for epoch in range(numEpochs):
    model.train()
    train_loss = 0

    for audioBatch, labelBatch, texts in dataLoader:
        inputValues = audioBatch.to(device)
        labels = labelBatch.to(device)

        # Now you can use input_values and labels in your model
        outputs = model(input_values=inputValues, labels=labels)
        loss = outputs.loss

        # And you have access to original texts if needed
        print(texts[0])  # Print first text in batch

        optimizer.zero_grad()
        loss.backward()
        optimizer.step()

        trainLoss += loss.item()

    avgTrainLoss = trainLoss / len(dataloader)
    scheduler.step()

    print(f"Epoch {epoch+1}/{numEpochs}, Train Loss: {avgTrainLoss:.4f}, LR: {scheduler.get_last_lr()[0]:.2e}")

    # Save best model
    if avgTrainLoss < bestLoss:
        bestLoss = avgTrainLoss
        model.save_adapter(os.path.join(MMS_ADAPTERS_DIR, adapterName))

    # Periodically check CER/WER on a few examples
    if (epoch + 1) % 10 == 0:
        model.eval()
        with torch.no_grad():
            for i in range(min(5, len(dataset))):
                #sample = dataset[i]
                inputValues, labelValues, text = dataset[i]
                #input_values = sample["input_values"].unsqueeze(0).to(device)
                inputValues = inputValues.unsqueeze(0).to(device)

                outputs = model(input_values=inputValues)
                predictedIds = torch.argmax(outputs.logits, dim=-1)

                transcription = processor.batch_decode(predictedIds)[0]
                #reference = sample["text"]

                print(f"Example {i+1}:")
                print(f"  Reference: {text}")
                print(f"  Predicted: {transcription}")


#def trainMMSAdapter(model, dataloader, optimizer, device, epochs=5):
#    model.train()
#
#    for epoch in range(epochs):
#        total_loss = 0
#        for batch in dataloader:
#            # Move batch to device
#            input_values = batch["input_values"].to(device)
#            labels = batch["labels"].to(device)
#
#            # Forward pass
#            outputs = model(input_values=input_values, labels=labels)
#            loss = outputs.loss
#
#            # Backward pass
#            optimizer.zero_grad()
#            loss.backward()
#            optimizer.step()
#
#            total_loss += loss.item()
#
#        print(f"Epoch {epoch+1}/{epochs}, Loss: {total_loss/len(dataloader)}")
#
