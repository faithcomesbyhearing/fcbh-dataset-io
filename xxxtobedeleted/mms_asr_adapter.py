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

#***# This is a second solution for Peft
modelName = "facebook/mms-1b-fl102"
base_model = Wav2Vec2ForCTC.from_pretrained(modelName)
# Then load the adapter
from peft import PeftModel, PeftConfig
config = PeftConfig.from_pretrained("path/to/adapter_directory")
model = PeftModel.from_pretrained(base_model, "path/to/adapter_directory")

### Warning not yet tested

if len(sys.argv) < 3:
    print("Usage: mms_asr.py {adapterName}")
    sys.exit(1)

adapterName = sys.argv[1]

if torch.cuda.is_available():
    device = 'cuda'
else:
    device = 'cpu'

modelId = "facebook/mms-1b-all"

# We don't need to specify target_lang when using adapters
processor = AutoProcessor.from_pretrained(modelId)
#model = Wav2Vec2ForCTC.from_pretrained(modelId, ignore_mismatched_sizes=True)
model = Wav2Vec2ForCTC.from_pretrained("path/to/adapter_directory", adapterName=adapterName)

# Load the adapter
#adapter_name = os.path.basename(adapter_path)
#model.load_adapter(adapter_path)
#model.set_active_adapters(adapter_name)

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


  ## Testing
  ## cd Documents/go2/dataset/mms
  ## conda activate mms_asr
  ## python mms_asr.py eng
  ## /Users/gary/FCBH2024/download/ENGWEB/ENGWEBN2DA-mp3-64/B02___01_Mark________ENGWEBN2DA.wav

