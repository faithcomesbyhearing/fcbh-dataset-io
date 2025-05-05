import sys
import json
import torch
import torchaudio
from torchaudio.pipelines import MMS_FA as bundle

# This was an attempt to reduce GPU memory by batching the text.
# While it works, it was not used, because it passes the entire audio 
# mission[0] into the align method with each call.  This means that if
# some phrases are repeated, the aligner could match to a first ocurrence
# rather than the correct one.  The could result in overlapping timestamps.

batch_size = 50

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
    num_frames = emission.size(1)
    ratio = waveform.size(1) / num_frames / sample_rate

    words = inp["words"]
    all_token_spans = []
    for i in range(0, len(words), batch_size):
        word_batch = words[i:i+batch_size]

        # Process this batch of words
        with torch.inference_mode():
            token_spans_batch = aligner(emission[0], tokenizer(word_batch))
        all_token_spans.extend(token_spans_batch)

    chapter = []
    for spans in all_token_spans:
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


