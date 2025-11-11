import os
import sys
import struct
import io
import torch
import numpy as np
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
from common import ensureMinimumTensorSize, decompress
from datasets import Dataset, Audio
from transformers import Wav2Vec2ForCTC
from transformers import Wav2Vec2Processor
from transformers import AutoProcessor


if len(sys.argv) < 2:
    print("Usage: python -u wav2vec2_asr.py  {iso639-3}", file=sys.stderr)
    sys.exit(1)
lang = sys.argv[1]
print("lang", lang, file=sys.stderr)
if torch.cuda.is_available():
    device = 'cuda'
else:
    device = 'cpu'
modelDir = os.path.join(os.getenv('FCBH_DATASET_DB'), 'wav2vec2_models', lang)
processor = Wav2Vec2Processor.from_pretrained(modelDir)
model = Wav2Vec2ForCTC.from_pretrained(modelDir)
model = model.to(device)
model.eval()
minTensorLength = 8000 # 0.5 sec

while True:
    length_bytes = sys.stdin.buffer.read(4)
    if len(length_bytes) < 4:
        break
    length = struct.unpack('>I', length_bytes)[0]
    blob_data = sys.stdin.buffer.read(length)
    if len(blob_data) < length:
        break  # Incomplete read
    inputTensor = decompress(blob_data, np.float32)
    originalLen = len(inputTensor)
    inputTensor, attentionMask = ensureMinimumTensorSize(inputTensor, minTensorLength, 0)
    inputTensor = inputTensor.to(device)
    attentionMask = attentionMask.to(device)
    inputs = {
        "input_values": inputTensor.unsqueeze(0),
        "attention_mask": attentionMask.unsqueeze(0)
    }
    with torch.no_grad():
        outputs = model(**inputs).logits
    ids = torch.argmax(outputs, dim=-1)[0]
    transcription = processor.decode(ids)
    print(transcription, flush=True)
    #print("\t", transcription, file=sys.stderr)


