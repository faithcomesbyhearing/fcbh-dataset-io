import os
import sys
import struct
import io
from datasets import Dataset, Audio
from transformers import Wav2Vec2ForCTC
from transformers import Wav2Vec2Processor
from transformers import AutoProcessor
import torch


def ensureMinimumTensorSize(tensor, minTensorLength, padValue):
    currentLength = tensor.shape[-1]
    if currentLength < minTensorLength:
        paddingNeeded = minTensorLength - currentLength
        tensor = torch.nn.functional.pad(tensor, (0, paddingNeeded), mode='constant', value=padValue)
    return tensor


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
    buffer = io.BytesIO(blob_data)
    inputTensor = torch.load(buffer)
    inputTensor = ensureMinimumTensorSize(inputTensor, minTensorLength, 0)
    inputTensor = inputTensor.to(device)
    inputs = {"input_values": inputTensor.unsqueeze(0)}
    with torch.no_grad():
        outputs = model(**inputs).logits
    ids = torch.argmax(outputs, dim=-1)[0]
    transcription = processor.decode(ids)
    print(transcription, flush=True)
    #sys.stdout.write(transcription)
    #sys.stdout.write("\n")
    #sys.stdout.flush()
    print("\t", transcription, file=sys.stderr)


