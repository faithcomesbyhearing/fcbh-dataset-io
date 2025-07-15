import torch
import torch.nn as nn
from torch.utils.data import DataLoader
from transformers import get_linear_schedule_with_warmup
import logging
# from train_adapter
import os
import sys
from transformers import Wav2Vec2FeatureExtractor
from transformers import Wav2Vec2Processor
from transformers import Wav2Vec2ForCTC
from tokenizer import createTokenizer
from sqlite_utility import *
from data_preparation import *
from dataset import *
#from evaluate import load
from safetensors.torch import save_file as safe_save_file
from transformers.models.wav2vec2.modeling_wav2vec2 import WAV2VEC2_ADAPTER_SAFE_FILE
#from memory_callback import *
#from torch.utils.data import DataLoader
import psutil

#
# This is sample code provided by Claude 7/12/25
#

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def train_mms_adapter(model, dataset, num_epochs=3, lr=5e-5,
                     warmup_steps=100, max_grad_norm=1.0, log_steps=50):
    """
    Train MMS adapter with raw PyTorch

    Args:
        model: Your MMS model with adapter
        dataset: Your custom dataset that returns padded batches as dicts
        num_epochs: Number of training epochs
        lr: Learning rate
        warmup_steps: Warmup steps for lr scheduler
        max_grad_norm: Max gradient norm for clipping
        log_steps: Steps between logging
    """
#dataset = MyDataset(sampleDB)
#num_epochs = 1
#lr = 5e-5
#warmup_steps = 20
#max_grad_norm = 1.0
#log_steps = 1

    dataloader = DataLoader(dataset, batch_size=None, shuffle=True)

    optimizer = torch.optim.AdamW(model.parameters(), lr=lr)
    total_steps = len(dataloader) * num_epochs
    scheduler = get_linear_schedule_with_warmup(
        optimizer, num_warmup_steps=warmup_steps, num_training_steps=total_steps
    )

    model.train()
    global_step = 0
    total_loss = 0

    for epoch in range(num_epochs):
        logger.info(f"Starting epoch {epoch + 1}/{num_epochs}")
        epoch_loss = 0

        #for step, batch in enumerate(tqdm(dataloader, desc=f"Epoch {epoch + 1}")):
        for step, batch in enumerate(dataloader):
        #dataloader_iter = iter(enumerate(dataloader))
        #step, batch = next(dataloader_iter):
            batch = {k: v.to(model.device) for k, v in batch.items()}
            outputs = model(**batch)
            loss = outputs.loss
            loss.backward()
            torch.nn.utils.clip_grad_norm_(model.parameters(), max_grad_norm)
            optimizer.step()
            scheduler.step()
            optimizer.zero_grad()
            epoch_loss += loss.item()
            total_loss += loss.item()
            global_step += 1
            if global_step % log_steps == 0:
                avg_loss = total_loss / global_step
                current_lr = scheduler.get_last_lr()[0]
                logger.info(
                    f"Step {epoch}/{step}: Loss = {loss.item():.4f}, "
                    f"Avg Loss = {avg_loss:.4f}, LR = {current_lr:.2e}"
                )
                if torch.cuda.is_available():
                    allocated = torch.cuda.memory_allocated() / 1024**3
                    reserved = torch.cuda.memory_reserved() / 1024**3
                    free = reserved - allocated
                    fragmentation = free / reserved if reserved > 0 else 0
                    gpuMax = torch.cuda.max_memory_allocated() / 1024**3
                    logger.info(f"Step: {epoch}/{step}  Allocated: {allocated:.2f}MB, Reserved: {reserved:.2f}GB, "
                             f"Free: {free:.2f}GB, Fragmentation: {fragmentation:.4f}, GPU Max {gpuMax}")
                    # Display model size
                    total_params = sum(p.numel() for p in model.parameters())
                    trainable_params = sum(p.numel() for p in model.parameters() if p.requires_grad)
                    print(f"Total parameters: {total_params / 1e9:.2f}GB")
                    print(f"Trainable parameters: {trainable_params / 1e6:.2f}MB")
                    print(f"Frozen parameters: {(total_params - trainable_params) / 1e9:.2f}GB")
                    if fragmentation > 0.3:  # 30% fragmentation
                        torch.cuda.empty_cache()
                        logger.warn("Cleared CUDA cache due to fragmentation")
                cpu_mem = psutil.virtual_memory().percent
                logger.info(f"Step {epoch}/{step}: CPU memory {cpu_mem:.1f}%")

        avg_epoch_loss = epoch_loss / len(dataloader)
        logger.info(f"Epoch {epoch + 1} completed. Average loss: {avg_epoch_loss:.4f}")

    logger.info("Training completed!")
    return model


if len(sys.argv) < 6:
    print("Usage: python train_adapter.py {iso639-3} {databasePath} {audioDirectory} {batchMB} {numEpochs}", file=sys.stderr)
    sys.exit(1)
targetLang = sys.argv[1]
databasePath = sys.argv[2]
audioDirectory = sys.argv[3]
batchSizeMB = int(sys.argv[4])
numEpochs = int(sys.argv[5])

print(targetLang, "BatchSizeMB", batchSizeMB, "NumEpochs", numEpochs)

database = SqliteUtility(databasePath)
tokenizer = createTokenizer(database, targetLang)

featureExtractor = Wav2Vec2FeatureExtractor(
    feature_size=1,
    sampling_rate=16000,
    padding_value=0.0,
    do_normalize=True,
    return_attention_mask=True
)

processor = Wav2Vec2Processor(
    feature_extractor=featureExtractor,
    tokenizer=tokenizer
)

sampleDB = dataPreparation(database, databasePath, audioDirectory, processor, 128, batchSizeMB)
#sampleDB = SqliteUtility(os.path.join(os.getenv('FCBH_DATASET_TMP'), 'N2KEUWB4.db'))
database.close()

model = Wav2Vec2ForCTC.from_pretrained(
    "facebook/mms-1b-all",
    ctc_loss_reduction="mean",
    pad_token_id=processor.tokenizer.pad_token_id,
    vocab_size=len(processor.tokenizer),
    ignore_mismatched_sizes=True,   # accept tokenizer of different size (required)
)
model.init_adapter_layers()
model.freeze_base_model()

adapter_weights = model._get_adapters()
for param in adapter_weights.values():
    param.requires_grad = True

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
model.to(device)

dataset = MyDataset(sampleDB)
warmupSteps = int(len(dataset) * numEpochs * 0.05)
trainedModel = train_mms_adapter(
        model,
        dataset,
        num_epochs = numEpochs,
        lr = 5e-5,
        warmup_steps = warmupSteps,
        max_grad_norm = 0.5,
        log_steps = 1
)
sampleDB.close()

outputDir = os.path.join(os.getenv('FCBH_DATASET_DB'), 'mms_adapters', targetLang)
adapterFile = WAV2VEC2_ADAPTER_SAFE_FILE.format(targetLang)
adapterFile = os.path.join(outputDir, adapterFile)
print("OutputDir for weights", adapterFile)
safe_save_file(trainedModel._get_adapters(), adapterFile, metadata={"format": "pt"})
processorDir = os.path.join(outputDir, "processor_" + targetLang)
print("OutputDir for processor", processorDir)
processor.save_pretrained(processorDir)
