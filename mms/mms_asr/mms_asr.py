import os
import json
import sys
from transformers import Wav2Vec2ForCTC
from transformers import AutoProcessor
import torch
import torchaudio
from transformers.models.wav2vec2.modeling_wav2vec2 import WAV2VEC2_ADAPTER_SAFE_FILE
from safetensors.torch import load_file
sys.path.insert(0, os.path.abspath(os.path.join(os.environ['GOPROJ'], 'logger')))
from error_handler import setup_error_handler


## Documentation used to write this program
## https://huggingface.co/docs/transformers/main/en/model_doc/mms
## This program is NOT reentrant because of torch.cuda.empty_cache()

setup_error_handler()

def isSupportedLanguage(modelId:str, lang:str):
    processor = AutoProcessor.from_pretrained(modelId)
    dict = processor.tokenizer.vocab.keys()
    for l in dict:
        if l == lang:
            return True
    return False

def ensureMinimumTensorSize(batch, minTensorLength, padValue):
    tensor = batch["input_values"]
    batch_size = tensor.shape[0]
    originalLen = tensor.shape[-1]
    if originalLen < minTensorLength:
        paddingNeeded = minTensorLength - originalLen
        tensor = torch.nn.functional.pad(tensor, (0, paddingNeeded), mode='constant', value=padValue)
        batch['input_values'] = tensor
        mask = torch.zeros((batch_size, minTensorLength), dtype=torch.long)
        mask[:, :originalLen] = 1
        batch['attention_mask'] = mask
    return batch

if len(sys.argv) < 2:
    print("Usage: mms_asr.py  {iso639-3}  adapter(optional)", file=sys.stderr)
    sys.exit(1)
lang = sys.argv[1]
adapter = len(sys.argv) > 2 and sys.argv[2].lower() == "adapter"
if torch.cuda.is_available():
    device = 'cuda'
else:
    device = 'cpu'
modelId = "facebook/mms-1b-all"
if adapter:
    outputDir = os.path.join(os.getenv('FCBH_DATASET_DB'), 'mms_adapters', lang)
    processorDir = os.path.join(outputDir, f"processor_{lang}")
    processor = AutoProcessor.from_pretrained(processorDir)
    model = Wav2Vec2ForCTC.from_pretrained(modelId)
    vocabSize = len(processor.tokenizer)
    hiddenSize = model.lm_head.weight.shape[1]
    model.lm_head = torch.nn.Linear(hiddenSize, vocabSize)
    adapterFile = WAV2VEC2_ADAPTER_SAFE_FILE.format(lang)
    adapterFile = os.path.join(outputDir, adapterFile)
    adapter_weights = load_file(adapterFile)
    model.load_state_dict(adapter_weights, strict=False)
else:
    if not isSupportedLanguage(modelId, lang):
        print(lang, "is not supported by", modelId, file=sys.stderr)
    processor = AutoProcessor.from_pretrained(modelId, target_lang=lang)
    model = Wav2Vec2ForCTC.from_pretrained(modelId, target_lang=lang, ignore_mismatched_sizes=True)

model = model.to(device)
for line in sys.stdin:
    torch.cuda.empty_cache()
    audio = json.loads(line)
    info = torchaudio.info(audio["path"], format="wav")
    if info.sample_rate != 16000:
        print("Audio sample rate must be 16000", file=sys.stderr, flush=True)
        sys.exit(1)
    if audio.get("end_ts") == 0:
        speech, sample_rate = torchaudio.load(audio["path"])
    else:
        speech, sample_rate = torchaudio.load(
            audio["path"],
            frame_offset=int(audio["begin_ts"] * 16000),
            num_frames=int((audio["end_ts"] - audio["begin_ts"]) * 16000)
        )
    inputTensor = processor(
        speech.squeeze(),
        sampling_rate=16000,
        return_tensors="pt",
        padding=False
    ).input_values
    inputs = ensureMinimumTensorSize({"input_values": inputTensor}, 3200, 0)
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
## {"path":"/Users/gary/FCBH2024/download/ENGWEB/ENGWEBN2DA-mp3-64/B02___01_Mark________ENGWEBN2DA.wav"}

# {"path": "/Users/gary/FCBH2024/download/ART_12231842/ART Line VOX/ART_888_AMO_009_00451_VOX.wav"}



