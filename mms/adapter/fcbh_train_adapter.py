import os
import sys
import torch
from transformers import Wav2Vec2ForCTC, Wav2Vec2Config, Wav2Vec2Processor
from transformers import Wav2Vec2CTCTokenizer, Wav2Vec2FeatureExtractor
from adapters import AdapterConfig, SeqBnConfig
from torch.optim.lr_scheduler import CosineAnnealingLR
from peft import PeftConfig, LoraConfig, get_peft_model
import jiwer
from fcbh_dataset import *
from fcbh_dataloader import *
from fcbh_vocabulary import *

MMS_ADAPTERS_DIR = os.path.join(os.getenv("FCBH_DATASET_DB"), "mms_adapters")
# batch size is recommended to be 4 to 8, try 8 to see if there is enough memory
# num workers is recommended to be 2 to 4, start with 2
# try 30-50 epochs, for over-learning do 60-100 epochs
# Consider data augmentation techniques (speed perturbation, noise addition)
# see audiomentations or torch-audiomentations

#
# https://docs.adapterhub.ml/training.html
#

if len(sys.argv) < 6:
    print("Usage: fcbh_train_adapter.py {iso639-3} {databasePath} {audioDirectory} {batchSize} {numEpochs}")
    sys.exit(1)
adapterName = sys.argv[1]
databasePath = sys.argv[2]
audioDirectory = sys.argv[3]
batchSize = int(sys.argv[4])
numEpochs = int(sys.argv[5])

# Load MMS model and processor
#modelName = "facebook/mms-1b-all"
modelName = "facebook/mms-1b-fl102"

vocabFile, vocabulary = getFCBHVocabulary(databasePath)
tokenizer = Wav2Vec2CTCTokenizer(vocab_file=vocabFile)

# Feature extractor (same as MMS)
feature_extractor = Wav2Vec2FeatureExtractor(
    feature_size=1, sampling_rate=16000, padding_value=0.0,
    do_normalize=True, return_attention_mask=True
)
processor = Wav2Vec2Processor(feature_extractor=feature_extractor, tokenizer=tokenizer)
model = Wav2Vec2ForCTC.from_pretrained(modelName)

model.lm_head = torch.nn.Linear(model.config.hidden_size, len(vocabulary))
model.config.vocab_size = len(vocabulary)

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

dataset = FCBHDataset(databasePath, audioDirectory, processor)
dataLoader = FCBHDataLoader(dataset, "full", batchSize)

# Training loop
bestLoss = float('inf')
for epoch in range(numEpochs):
    model.train()
    trainLoss = 0

    #for audioBatch, labelBatch, texts in dataLoader:
    for batch_idx, (inputValues, labels, texts) in enumerate(dataLoader):
        print(texts)
        inputValues = inputValues.to(device)
        #attentionMask = attentionMask.to(device)
        labels = labels.to(device)

        # process inputs in model
        #outputs = model(input_values=inputValues, labels=labels)
        outputs = model(
            input_values=inputValues,
            #attention_mask=attentionMask,
            labels=labels
        )

        #outputs = model(input_values=inputValues)
        #print(f"Logits shape: {outputs.logits.shape}")
        #predicted_ids = torch.argmax(outputs.logits, dim=-1)
        #print(f"Predicted IDs: {predicted_ids}")
        #print(f"Unique IDs: {torch.unique(predicted_ids).tolist()}")
        loss = outputs.loss

        optimizer.zero_grad()
        loss.backward()
        optimizer.step()
        scheduler.step()

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



