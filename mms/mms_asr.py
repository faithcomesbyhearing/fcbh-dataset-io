import os
import sys
from datasets import Dataset, Audio
from transformers import Wav2Vec2ForCTC
from transformers import AutoProcessor
import torch
import psutil


## Documentation used to write this program
## https://huggingface.co/docs/transformers/main/en/model_doc/mms
## This program is NOT reentrant because of torch.cuda.empty_cache()

def isSupportedLanguage(modelId:str, lang:str):
    processor = AutoProcessor.from_pretrained(modelId)
    dict = processor.tokenizer.vocab.keys()
    for l in dict:
        if l == lang:
            return True
    return False


if len(sys.argv) < 2:
    print("Usage: mms_asr.py  {iso639-3}")
    sys.exit(1)
lang = sys.argv[1]
if torch.cuda.is_available():
    device = 'cuda'
else:
    device = 'cpu'
print("Using device:", device, file=sys.stderr)
modelId = "facebook/mms-1b-all"
if not isSupportedLanguage(modelId, lang):
    print(lang, "is not supported by", modelId)
processor = AutoProcessor.from_pretrained(modelId, target_lang=lang)
model = Wav2Vec2ForCTC.from_pretrained(modelId, target_lang=lang, ignore_mismatched_sizes=True)
model = model.to(device)

for line in sys.stdin:
    torch.cuda.empty_cache() # This will not be OK for concurrent processes
    audioFile = line.strip()
    fromDict = Dataset.from_dict({"audio": [audioFile]})
    streamData = fromDict.cast_column("audio", Audio(sampling_rate=16000))
    sample = next(iter(streamData))["audio"]["array"]

    inputs = processor(sample, sampling_rate=16_000, return_tensors="pt")
    inputs = {name: tensor.to(device) for name, tensor in inputs.items()}
    with torch.no_grad():
        outputs = model(**inputs).logits
    ids = torch.argmax(outputs, dim=-1)[0]
    transcription = processor.decode(ids)
    sys.stdout.write(transcription)
    sys.stdout.write("\n")
    sys.stdout.flush()
    memory = psutil.Process().memory_info().rss / 1024 / 1024 / 1024
    print(f"mms_sr.py: {memory:.2f} GB", file=sys.stderr)

#$HOME + "/NPIDPI/NPIDPIN1DA/B02___01_Mark________NPIDPIN1DA.wav"

# /Users/gary/FCBH2024/download/NPIDPI/NPIDPIN1DA/B02___01_Mark________NPIDPIN1DA.wav

## Testing
## cd Documents/go2/dataset/mms
## conda activate mms_asr
## python mms_asr.py eng
## /Users/gary/FCBH2024/download/ENGWEB/ENGWEBN2DA-mp3-64/B02___01_Mark________ENGWEBN2DA.wav



