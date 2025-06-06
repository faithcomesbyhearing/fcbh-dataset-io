import os
import sys
from datasets import Dataset, Audio
from transformers import Wav2Vec2ForCTC
from transformers import AutoProcessor
import torch
import psutil
from beam_search_decoder import create_decoder


## Documentation used to write this program
## https://huggingface.co/docs/transformers/main/en/model_doc/mms
## This program is NOT reentrant because of torch.cuda.empty_cache()

#def isSupportedLanguage(modelId:str, lang:str):
#    processor = AutoProcessor.from_pretrained(modelId)
#    dict = processor.tokenizer.vocab.keys()
#    for l in dict:
#        if l == lang:
#            return True
#    return False


if len(sys.argv) < 4:
    print("Usage: mms_asr.py  {iso639-3}  {database_path} {lexicon_directory}")
    sys.exit(1)
lang = sys.argv[1]
db_path = sys.argv[2]
lex_directory = sys.argv[3]
if torch.cuda.is_available():
    device = 'cuda'
else:
    device = 'cpu'
#modelId = "facebook/mms-1b-all"
modelId = "mms-meta/mms-zeroshot-300m"
#if not isSupportedLanguage(modelId, lang):
#    print(lang, "is not supported by", modelId)
processor = AutoProcessor.from_pretrained(modelId)#, target_lang=lang)
vocab = processor.tokenizer.get_vocab()
decoder = create_decoder(db_path, vocab, lex_directory)
#model = Wav2Vec2ForCTC.from_pretrained(modelId, target_lang=lang, ignore_mismatched_sizes=True)
model = Wav2Vec2ForCTC.from_pretrained(modelId, ignore_mismatched_sizes=True)
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
    log_probs = torch.log_softmax(outputs, dim=-1).cpu().numpy()
    emissions = log_probs.squeeze(0) # remove batch dimension
    # emissions is numpy.array of emitting model predictions with shape [T, N], where T is time, N is number of tokens
    time_dim = emissions.shape[0]
    num_tokens = emissions.shape[1]
    results = decoder.decode(emissions.ctypes.data, time_dim, num_tokens)
    transcription = processor.decode(results[0].tokens)

    ids = torch.argmax(outputs, dim=-1)[0]
    orig_transcription = processor.decode(ids)
    #sys.stdout.write("\nGreedy Decoding\n")
    #print(orig_transcription)
    #print()
    #sys.stdout.write("Beam Search Decoding\n")
    sys.stdout.write(transcription)
    #print()
    sys.stdout.write("\n")
    sys.stdout.flush()


## Testing
## cd Documents/go2/dataset/mms
## conda activate mms_asr
## python mms_asr.py eng data
## /Users/gary/FCBH2024/download/ENGWEB/ENGWEBN2DA-mp3-64/B02___01_Mark________ENGWEBN2DA.wav

## python mms_asr.py cul /Users/gary/FCBH2024/GaryNTest/N2CUL_MNT.db data
## /Users/gary/FCBH2024/download/CULMNT/CULMNTN2DA/MAT_28_24sec.wav



