import sys
import json
import torch
import torchaudio
from torchaudio.pipelines import MMS_FA as bundle
#import multiprocessing


# mms_forced_aligner was developed with the following documentation
# https://pytorch.org/audio/main/tutorials/forced_alignment_for_multilingual_data_tutorial.html
# https://pytorch.org/audio/main/tutorials/ctc_forced_alignment_api_tutorial.html
# A conda environment was created
# conda create -n mms_fa python=3.11
# conda activate mms_fa
# cd mms_fa
# conda install pytorch::pytorch torchvision torchaudio -c pytorch
# pip install uroman # conda does not have it
# pip install soundfile # needed for loading audio files
# This program is NOT reentrant because of torch.cuda.empty_cache()


#if len(sys.argv) < 2:
#    sys.stderr.write("Usage: mms_align.py  {iso639-3}\n")
#    sys.exit(1)
#lang = sys.argv[1]
device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
model = bundle.get_model(with_star=False)
model.to(device)
tokenizer = bundle.get_tokenizer()
aligner = bundle.get_aligner()
for line in sys.stdin:
    torch.cuda.empty_cache() # This will not be OK for concurrent processes
    inp = json.loads(line)
    waveform, sample_rate = torchaudio.load(inp["audio_file"])
    assert sample_rate == bundle.sample_rate
    with torch.inference_mode():
        emission, _ = model(waveform.to(device))
        token_spans = aligner(emission[0], tokenizer(inp["words"]))
    num_frames = emission.size(1)
    ratio = waveform.size(1) / num_frames / sample_rate
    chapter = []
    for spans in token_spans:
        word = []
        for token in spans:
            char = [ token.token, token.start, token.end, token.score]
            word.append(char)
        chapter.append(word)
        word = []
    result = {
       'ratio': ratio,
       'dictionary': bundle.get_dict(),
       'tokens': chapter
    }
    output = json.dumps(result)
    sys.stdout.write(output)
    sys.stdout.write("\n")
    sys.stdout.flush()

# Testing
# conda activate mms_fa
# time python mms_align.py eng < engweb_fa_inp.json > engweb_fa_out.json

# small sample
# time python mms_align.py deu < german.json


