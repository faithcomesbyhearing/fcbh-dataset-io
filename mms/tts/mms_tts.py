import json
import torch
from transformers import VitsTokenizer, VitsModel, set_seed
import scipy

class MMSTextToSpeech:

    def __init__(self, lang: str):
        # Test that lang is supported?
        modelId = "facebook/mms-tts-" + lang
        self.tokenizer = VitsTokenizer.from_pretrained(modelId)
        self.model = VitsModel.from_pretrained(modelId)


    #def isTTSLanguage(self, lang:str):
    #    dict = self.tokenizer.vocab.keys()
    #    for l in dict:
    #        if l == lang:
    #            return True
    #    return False


    def generate(self, text: str, outputPath: str):
        inputs = self.tokenizer(text=text, return_tensors="pt")
        set_seed(555)  # make deterministic
        with torch.no_grad():
            outputs = self.model(**inputs)
        waveform = outputs.waveform[0]
        #scipy.io.wavfile.write(outputPath, rate=self.model.config.sampling_rate, data=waveform)


if __name__ == "__main__":
    with open("mms_npi.json", "r") as file:
        texts = json.load(file)
    print(texts[5])
    tts = MMSTextToSpeech("xxx")
    #ans = tts.isTTSLanguage("eng")
    sample = texts[5]
    #text = sample["script_text"]
    text = "A cat in the house at tom's ice cream"
    outputPath = "npi" + "_" + sample["reference"] + ".wav"
    tts.generate(text, outputPath)



