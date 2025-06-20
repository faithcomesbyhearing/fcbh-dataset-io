import os
import sys
import torch
from transformers import Wav2Vec2FeatureExtractor
from transformers import Wav2Vec2Processor
from transformers import Trainer
from transformers import TrainingArguments
from transformers import Wav2Vec2ForCTC
from tokenizer import createTokenizer
from sqlite_utility import *
from data_pruner import dataPruner
from dataset import *
from mms_collator import *
from evaluate import load
from safetensors.torch import save_file as safe_save_file
from transformers.models.wav2vec2.modeling_wav2vec2 import WAV2VEC2_ADAPTER_SAFE_FILE

#
# https://huggingface.co/blog/mms_adapters
#

wer_metric = load("wer")

def compute_metrics(pred):
    pred_logits = pred.predictions
    pred_ids = np.argmax(pred_logits, axis=-1)
    pred.label_ids[pred.label_ids == -100] = processor.tokenizer.pad_token_id
    pred_str = processor.batch_decode(pred_ids)
    # we do not want to group tokens when computing the metrics
    label_str = processor.batch_decode(pred.label_ids, group_tokens=False)
    wer = wer_metric.compute(predictions=pred_str, references=label_str)
    return {"wer": wer}


if len(sys.argv) < 6:
    print("Usage: python train_adapter.py {iso639-3} {databasePath} {audioDirectory} {batchSize} {numEpochs}")
    sys.exit(1)
targetLang = sys.argv[1]
databasePath = sys.argv[2]
audioDirectory = sys.argv[3]
batchSize = int(sys.argv[4])
numEpochs = int(sys.argv[5])

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

dataPruner(database) # remove lines with likely errors
dataset = MyDataset(database, audioDirectory, processor)
database.close()

#data_collator = DataCollatorCTCWithPadding(
#    processor=processor,
#    padding=True,
#    return_attention_mask=True,
#)

dataCollator = SimpleMMSCollator(processor)

model = Wav2Vec2ForCTC.from_pretrained(
    "facebook/mms-1b-all",
    attention_dropout=0.0,          # No attention droppout
    hidden_dropout=0.0,             # No layer? attention
    feat_proj_dropout=0.0,          # No feature projection dropout
    layerdrop=0.0,                  # No layer dropout
    ctc_loss_reduction="mean",
    pad_token_id=processor.tokenizer.pad_token_id,
    vocab_size=len(processor.tokenizer),
    ignore_mismatched_sizes=True,   # accept tokenizer of different size (required I think)
)

# Claude comment: You've set most dropout parameters to 0.0, which means no regularization through
# dropout. This might be appropriate if your dataset is large enough or if you're using other
# regularization techniques, but adding some dropout can help prevent overfitting on smaller
# datasets.

model.init_adapter_layers()
model.freeze_base_model()

adapter_weights = model._get_adapters()
for param in adapter_weights.values():
    param.requires_grad = True

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
useGPU = (device == "gpu")

trainingArgs = TrainingArguments(
  output_dir = os.path.join(os.getenv('FCBH_DATASET_DB'), 'mms_adapters'),
  group_by_length = True,
  per_device_train_batch_size = batchSize,
  #eval_strategy = "epoch",
  save_strategy = "epoch",          # Save checkpoints every epoch
  logging_strategy = "steps",       # Log results every epoch
  num_train_epochs = numEpochs,
  use_cpu = not useGPU,
  gradient_checkpointing = True,  # True reduces memory use at cost of performance
  fp16 = useGPU, # could speed up GPU
  #save_steps=200,
  #eval_steps=100,
  logging_steps=10,
  learning_rate = 1e-3,
  warmup_steps = 100,
  save_total_limit = 2,
  push_to_hub = False,
)

trainer = Trainer(
    model = model,
    data_collator = dataCollator,
    args = trainingArgs,
    compute_metrics = compute_metrics,
    train_dataset = dataset,
    #eval_dataset = dataset, Avoid doing eval, until eval dataset is developed
    #tokenizer=processor.feature_extractor,
    processing_class = processor.feature_extractor,
)

trainer.train()

adapterFile = WAV2VEC2_ADAPTER_SAFE_FILE.format(targetLang)
adapterFile = os.path.join(trainingArgs.output_dir, adapterFile)
safe_save_file(model._get_adapters(), adapterFile, metadata={"format": "pt"})
processorDir = os.path.join(trainingArgs.output_dir, "processor_" + targetLang)
processor.save_pretrained(processorDir)

""" Loading Adapter
model = Wav2Vec2ForCTC.from_pretrained("facebook/mms-1b-all")
model.init_adapter_layers()

# Load your trained adapter
from safetensors.torch import load_file as safe_load_file
adapter_weights = safe_load_file(adapter_file)
model.load_adapter(adapter_weights, target_lang)
processorDir = os.path.join(training_args.output_dir, "processor_" + targetLang)
processor = Wav2Vec2Processor.from_pretrained(processorDir)
"""

""" Alternative Save Model
# modelPath = os.path.join(os.getenv('FCBH_DATASET_DB'), 'models', targetLang + '_adapter')
# trainer.save_model(modelPath)
# processor.save_pretrained(modelPath)

# Then load for inference
# model = Wav2Vec2ForCTC.from_pretrained(modelPath)
# processor = Wav2Vec2Processor.from_pretrained(modelPath)
"""






