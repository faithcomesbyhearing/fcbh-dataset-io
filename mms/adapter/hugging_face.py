# Requirements
# !pip install datasets[audio]
#  !pip install evaluate
#  !pip install git+https://github.com/huggingface/transformers.git
#  !pip install jiwer
#  !pip install accelerate

#
# Create Wav2Vec2CTCTokenizer
#

from datasets import load_dataset, load_metric, Audio

common_voice_train = load_dataset("mozilla-foundation/common_voice_6_1", "tr", split="train+validation", use_auth_token=True)
common_voice_test = load_dataset("mozilla-foundation/common_voice_6_1", "tr", split="test", use_auth_token=True)

common_voice_train = common_voice_train.remove_columns(["accent", "age", "client_id", "down_votes", "gender", "locale", "segment", "up_votes"])
common_voice_test = common_voice_test.remove_columns(["accent", "age", "client_id", "down_votes", "gender", "locale", "segment", "up_votes"])

#from datasets import ClassLabel
#import random
#import pandas as pd
#from IPython.display import display, HTML

#def show_random_elements(dataset, num_examples=10):
#    assert num_examples <= len(dataset), "Can't pick more elements than there are in the dataset."
#    picks = []
#    for _ in range(num_examples):
#        pick = random.randint(0, len(dataset)-1)
#        while pick in picks:
#            pick = random.randint(0, len(dataset)-1)
#        picks.append(pick)
#
#    df = pd.DataFrame(dataset[picks])
#    display(HTML(df.to_html()))

#show_random_elements(common_voice_train.remove_columns(["path", "audio"]), num_examples=10)

import re
chars_to_remove_regex = '[\,\?\.\!\-\;\:\"\“\%\‘\”\�\']'

def remove_special_characters(batch):
    batch["sentence"] = re.sub(chars_to_remove_regex, '', batch["sentence"]).lower()
    return batch

common_voice_train = common_voice_train.map(remove_special_characters)
common_voice_test = common_voice_test.map(remove_special_characters)

#def replace_hatted_characters(batch):
#     batch["sentence"] = re.sub('[â]', 'a', batch["sentence"])
#     batch["sentence"] = re.sub('[î]', 'i', batch["sentence"])
#     batch["sentence"] = re.sub('[ô]', 'o', batch["sentence"])
#     batch["sentence"] = re.sub('[û]', 'u', batch["sentence"])
#     return batch

#common_voice_train = common_voice_train.map(replace_hatted_characters)
#common_voice_test = common_voice_test.map(replace_hatted_characters)

def extract_all_chars(batch):
  all_text = " ".join(batch["sentence"])
  vocab = list(set(all_text))
  return {"vocab": [vocab], "all_text": [all_text]}

vocab_train = common_voice_train.map(extract_all_chars, batched=True, batch_size=-1, keep_in_memory=True, remove_columns=common_voice_train.column_names)
vocab_test = common_voice_test.map(extract_all_chars, batched=True, batch_size=-1, keep_in_memory=True, remove_columns=common_voice_test.column_names)

vocab_list = list(set(vocab_train["vocab"][0]) | set(vocab_test["vocab"][0]))

vocab_dict = {v: k for k, v in enumerate(sorted(vocab_list))}

vocab_dict["|"] = vocab_dict[" "]
del vocab_dict[" "]

vocab_dict["[UNK]"] = len(vocab_dict)
vocab_dict["[PAD]"] = len(vocab_dict)
len(vocab_dict)

target_lang = "tur"

new_vocab_dict = {target_lang: vocab_dict}

# Note: In case you want to use this notebook to add a new adapter layer to an existing model repo
# use make sure to not create an empty, new vocab dict, but instead re-use one that already exists.
# To do so you should uncomment the following cells and replace mms_adapter_repo with a
# model repo id to which you want to add your adapter weights.

# from transformers import Wav2Vec2CTCTokenizer
# mms_adapter_repo = "patrickvonplaten/wav2vec2-large-mms-1b-turkish-colab"  # make sure to replace this path with a repo to which you want to add your new adapter weights
# tokenizer = Wav2Vec2CTCTokenizer.from_pretrained(mms_adapter_repo)
# new_vocab = tokenizer.vocab
# new_vocab[target_lang] = vocab_dict

import json
with open('vocab.json', 'w') as vocab_file:
    json.dump(new_vocab_dict, vocab_file)

from transformers import Wav2Vec2CTCTokenizer

tokenizer = Wav2Vec2CTCTokenizer.from_pretrained("./", unk_token="[UNK]", pad_token="[PAD]", word_delimiter_token="|", target_lang=target_lang)

repo_name = "wav2vec2-large-mms-1b-turkish-colab"

tokenizer.push_to_hub(repo_name)

#
# Create Wav2Vec2FeatureExtractor
#

from transformers import Wav2Vec2FeatureExtractor

feature_extractor = Wav2Vec2FeatureExtractor(feature_size=1, sampling_rate=16000, padding_value=0.0, do_normalize=True, return_attention_mask=True)

from transformers import Wav2Vec2Processor

processor = Wav2Vec2Processor(feature_extractor=feature_extractor, tokenizer=tokenizer)

#
# PreProcess Data
#

common_voice_train[0]["path"]

common_voice_train[0]["audio"]

common_voice_train = common_voice_train.cast_column("audio", Audio(sampling_rate=16_000))
common_voice_test = common_voice_test.cast_column("audio", Audio(sampling_rate=16_000))

# import IPython.display as ipd
# import numpy as np
# import random

# rand_int = random.randint(0, len(common_voice_train)-1)

# print(common_voice_train[rand_int]["sentence"])
# ipd.Audio(data=common_voice_train[rand_int]["audio"]["array"], autoplay=True, rate=16000)

rand_int = random.randint(0, len(common_voice_train)-1)

print("Target text:", common_voice_train[rand_int]["sentence"])
print("Input array shape:", common_voice_train[rand_int]["audio"]["array"].shape)
print("Sampling rate:", common_voice_train[rand_int]["audio"]["sampling_rate"])

common_voice_train = common_voice_train.map(prepare_dataset, remove_columns=common_voice_train.column_names)
common_voice_test = common_voice_test.map(prepare_dataset, remove_columns=common_voice_test.column_names)

# Note: datasets automatically takes care of audio loading and resampling. If you wish
# to implement your own customized data loading/sampling, feel free to just make use of
# the "path" column instead and disregard the "audio" column.

#
# Training
#

import torch

from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Union

@dataclass
class DataCollatorCTCWithPadding:
    """
    Data collator that will dynamically pad the inputs received.
    Args:
        processor (:class:`~transformers.Wav2Vec2Processor`)
            The processor used for proccessing the data.
        padding (:obj:`bool`, :obj:`str` or :class:`~transformers.tokenization_utils_base.PaddingStrategy`, `optional`, defaults to :obj:`True`):
            Select a strategy to pad the returned sequences (according to the model's padding side and padding index)
            among:
            * :obj:`True` or :obj:`'longest'`: Pad to the longest sequence in the batch (or no padding if only a single
              sequence if provided).
            * :obj:`'max_length'`: Pad to a maximum length specified with the argument :obj:`max_length` or to the
              maximum acceptable input length for the model if that argument is not provided.
            * :obj:`False` or :obj:`'do_not_pad'` (default): No padding (i.e., can output a batch with sequences of
              different lengths).
    """

    processor: Wav2Vec2Processor
    padding: Union[bool, str] = True

    def __call__(self, features: List[Dict[str, Union[List[int], torch.Tensor]]]) -> Dict[str, torch.Tensor]:
        # split inputs and labels since they have to be of different lenghts and need
        # different padding methods
        input_features = [{"input_values": feature["input_values"]} for feature in features]
        label_features = [{"input_ids": feature["labels"]} for feature in features]

        batch = self.processor.pad(
            input_features,
            padding=self.padding,
            return_tensors="pt",
        )
        labels_batch = self.processor.pad(
            labels=label_features,
            padding=self.padding,
            return_tensors="pt",
        )

        # replace padding with -100 to ignore loss correctly
        labels = labels_batch["input_ids"].masked_fill(labels_batch.attention_mask.ne(1), -100)

        batch["labels"] = labels

        return batch

data_collator = DataCollatorCTCWithPadding(processor=processor, padding=True)

from evaluate import load

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

from transformers import Wav2Vec2ForCTC

model = Wav2Vec2ForCTC.from_pretrained(
    "facebook/mms-1b-all",
    attention_dropout=0.0,
    hidden_dropout=0.0,
    feat_proj_dropout=0.0,
    layerdrop=0.0,
    ctc_loss_reduction="mean",
    pad_token_id=processor.tokenizer.pad_token_id,
    vocab_size=len(processor.tokenizer),
    ignore_mismatched_sizes=True,
)

model.init_adapter_layers()

model.freeze_base_model()

adapter_weights = model._get_adapters()
for param in adapter_weights.values():
    param.requires_grad = True

from transformers import TrainingArguments

training_args = TrainingArguments(
  output_dir=repo_name,
  group_by_length=True,
  per_device_train_batch_size=32,
  evaluation_strategy="steps",
  num_train_epochs=4,
  gradient_checkpointing=True,
  fp16=True,
  save_steps=200,
  eval_steps=100,
  logging_steps=100,
  learning_rate=1e-3,
  warmup_steps=100,
  save_total_limit=2,
  push_to_hub=True,
)

from transformers import Trainer

trainer = Trainer(
    model=model,
    data_collator=data_collator,
    args=training_args,
    compute_metrics=compute_metrics,
    train_dataset=common_voice_train,
    eval_dataset=common_voice_test,
    tokenizer=processor.feature_extractor,
)

trainer.train()

from safetensors.torch import save_file as safe_save_file
from transformers.models.wav2vec2.modeling_wav2vec2 import WAV2VEC2_ADAPTER_SAFE_FILE
import os

adapter_file = WAV2VEC2_ADAPTER_SAFE_FILE.format(target_lang)
adapter_file = os.path.join(training_args.output_dir, adapter_file)

safe_save_file(model._get_adapters(), adapter_file, metadata={"format": "pt"})

trainer.push_to_hub()

model_id = "patrickvonplaten/wav2vec2-large-mms-1b-turkish-colab"

model = Wav2Vec2ForCTC.from_pretrained(model_id, target_lang="tur").to("cuda")
processor = Wav2Vec2Processor.from_pretrained(model_id)

processor.tokenizer.set_target_lang("tur")

#
# Transcript
#

from datasets import Audio

common_voice_test_tr = load_dataset("mozilla-foundation/common_voice_6_1", "tr", data_dir="./cv-corpus-6.1-2020-12-11", split="test", use_auth_token=True)
common_voice_test_tr = common_voice_test_tr.cast_column("audio", Audio(sampling_rate=16_000))