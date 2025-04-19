import os
import sys
import torch
from transformers import Wav2Vec2ForCTC, Wav2Vec2Config, Wav2Vec2Processor
from adapters import AdapterConfig, SeqBnConfig
from torch.optim.lr_scheduler import CosineAnnealingLR
from peft import PeftConfig, LoraConfig, get_peft_model
import jiwer
from fcbh_dataset import *
from fcbh_dataloader import *

MMS_ADAPTERS_DIR = os.path.join(os.getenv("FCBH_DATASET_DB"), "mms_adapters")
# batch size is recommended to be 4 to 8, try 8 to see if there is enough memory
# num workers is recommended to be 2 to 4, start with 2
# try 30-50 epochs, for over-learning do 60-100 epochs
# Consider data augmentation techniques (speed perturbation, noise addition)
# see audiomentations or torch-audiomentations

#
# https://docs.adapterhub.ml/training.html
#

if len(sys.argv) < 7:
    print("Usage: fcbh_train_adapter.py {iso639-3} {databasePath} {audioDirectory} {numWorkers} {batchSize} {numEpochs}")
    sys.exit(1)
adapterName = sys.argv[1]
databasePath = sys.argv[2]
audioDirectory = sys.argv[3]
numWorkers = int(sys.argv[4])
batchSize = int(sys.argv[5])
numEpochs = int(sys.argv[6])

# Load MMS model and processor
#modelName = "facebook/mms-1b-all"
#modelName = "facebook/wav2vec2-base-960h"
#modelName = "facebook/wav2vec2-base"
modelName = "facebook/mms-1b-fl102"
processor = Wav2Vec2Processor.from_pretrained(modelName)
model = Wav2Vec2ForCTC.from_pretrained(modelName)

wav2Vec2Processor = Wav2Vec2Processor.from_pretrained(modelName)
dataset = FCBHDataset(databasePath, audioDirectory, wav2Vec2Processor)

vocabSize = dataset.getVocabularySize() + 100
print("vocabSize", vocabSize)
model.lm_head = torch.nn.Linear(model.config.hidden_size, vocabSize)
model.config.vocab_size = vocabSize

adapter_config = LoraConfig(
    r=16,  # reduction_factor
    lora_alpha=32,
    target_modules=["q_proj", "v_proj"],
    lora_dropout=0.05,
    bias="none"
)
model = get_peft_model(model, adapter_config)

# Setup device, optimizer
device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
model.to(device)

optimizer = torch.optim.AdamW(model.parameters(), lr=2e-5)
scheduler = CosineAnnealingLR(optimizer, T_max=numEpochs, eta_min=1e-6)

dataLoader = FCBHDataLoader(dataset, "train", batchSize, numWorkers)

# Training loop
bestLoss = float('inf')
for epoch in range(numEpochs):
    model.train()
    trainLoss = 0

    for audioBatch, labelBatch, texts in dataLoader:
        print(texts)
        inputValues = audioBatch.to(device)
        labels = labelBatch.to(device)

        # process inputs in model
        outputs = model(input_values=inputValues, labels=labels)
        loss = outputs.loss

        optimizer.zero_grad()
        loss.backward()
        optimizer.step()

        trainLoss += loss.item()

    avgTrainLoss = trainLoss / len(dataLoader)
    scheduler.step()

    print(f"Epoch {epoch+1}/{numEpochs}, Train Loss: {avgTrainLoss:.4f}, LR: {scheduler.get_last_lr()[0]:.2e}")

    # Save best model
    if avgTrainLoss < bestLoss:
        bestLoss = avgTrainLoss
        model.save_pretrained(
            os.path.join(MMS_ADAPTERS_DIR, adapterName),
            state_dict=model.state_dict(),
            save_embedding_layers=False
        )

    # Periodically check CER/WER on a few examples
    #if (epoch + 1) % 10 == 0:
    if True:
        model.eval()
        totalWer = 0
        totalCer = 0
        numSamples = min(5, len(dataset))

        with torch.no_grad():
            for i in range(numSamples):
                inputValues, labelValues, text = dataset[i]
                inputValues = inputValues.unsqueeze(0).to(device)

                outputs = model(input_values=inputValues)
                predictedIds = torch.argmax(outputs.logits, dim=-1)
                transcription = processor.batch_decode(predictedIds)[0]

                # Calculate WER (Word Error Rate)
                wer = jiwer.wer(text, transcription)
                print("wer", i, wer)

                # Calculate CER (Character Error Rate)
                cer = jiwer.cer(text, transcription)
                print("cer", i, cer)

                totalWer += wer
                totalCer += cer
                print("total wer", i, totalWer)
                print("total cer", i, totalCer)

                print(f"Example {i+1}:")
                print(f"  Reference: {text}")
                print(f"  Predicted: {transcription}")
                print(f"  WER: {wer:.4f}")
                print(f"  CER: {cer:.4f}")

        # Calculate average WER and CER
        avgWer = totalWer / numSamples
        avgCer = totalCer / numSamples

        print("\nEpoch", epoch)
        print(f"Average WER: {avgWer:.4f}")
        print(f"Average CER: {avgCer:.4f}")



