import os
import sys
import numpy as np
import torch
import torchaudio
import json

#
# July 18, 25. This has been written to replace Wav2Vec2FeatureExtractor in both the adapter
# and in the ASR routine when using the adapter.  When not using the adapter the ASR
# routine should continue to use Wav2Vec2
#
# It has not yet been incorporated into the code base.
"""
MyFeatureExtractor
1. Install MyFeatureExtractor in adapter directory
2. Trainer.py line 111-117 comment out.
3. Trainer.py add featureExtractor = MyFeatureExtractor()
4. Trainer.py line 120 is unchanged (set featureExtrctor in processor
"""

class MyFeatureExtractor:
	def __init__(self):


	def from_pretrained(self, directory):
		filepath = os.path.join(directory, "preprocessor_config.json")
		with open(filepath, "r") as f:
			data = json.load(f)
		return self


	def save_pretrained(self, directory):
		os.makedirs(directory, exist_ok=True)  # Ensure directory exists
		data = { "feature_extractor_type": "MyFeatureExtractor" }
		filepath = os.path.join(directory, "preprocessor_config.json")
		with open(filepath, "w") as f:
    		json.dump(data, f)


	def loadFile(self, audioPath):
        speech, sampleRate = torchaudio.load(audioPath)
    	if sampleRate != 16000:
    		print("Audio file", audioPath, "has sample_rate", sampleRate, file=sys.stderr)
    		sys.exit(1)
        return self(speech)


    def loadSegment(self, audioPath, beginTS, endTS):
        speech, sampleRate = torchaudio.load(
	        audioPath,
	        frame_offset = int(beginTS * 16000),
	        num_frames = int((endTS - beginTS) * 16000)
	    )
		if sampleRate != 16000:
    		print("Audio file", audioPath, "has sample_rate", sampleRate, file=sys.stderr)
    		sys.exit(1)
		return self(speech)


    def __call__(self, speech, sampling_rate=16000, return_tensors="pt"):
      	speech = speech.squeeze().numpy()
    	# normalization
    	speech = speech - np.mean(speech)
        speech = np.array(speech)
        inputValuesTensor = torch.tensor(speech, dtype=torch.float).squeeze(0)
        return inputValuesTensor





