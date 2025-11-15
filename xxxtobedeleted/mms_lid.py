import os
from datasets import load_dataset, Dataset, Audio
from transformers import Wav2Vec2ForSequenceClassification, AutoFeatureExtractor
import torch

# This code comes from the following page on huggingface
# https://huggingface.co/facebook/mms-lid-4017

class MMSLanguageIdentifier:

    ## create an instance of a language identifier to be used any number of times
    def __init__(self):
        model_id = "facebook/mms-lid-4017"
        self.processor = AutoFeatureExtractor.from_pretrained(model_id)
        self.model = Wav2Vec2ForSequenceClassification.from_pretrained(model_id)


    # Function to identify the language of an audio file
    def identify(self, audioFile: str) -> str:
        fromDict = Dataset.from_dict({"audio": [audioFile]})
        streamData = fromDict.cast_column("audio", Audio())
        sample = next(iter(streamData))["audio"]["array"]
        inputs = self.processor(sample, sampling_rate=16_000, return_tensors="pt")
        with torch.no_grad():
            outputs = self.model(**inputs).logits
        lang_id = torch.argmax(outputs, dim=-1)[0].item()
        detected_lang = self.model.config.id2label[lang_id]
        return detected_lang


if __name__ == "__main__":
    audioPath = os.environ.get("FCBH_DATASET_FILES") + "/NPIDPI/NPIDPIN1DA/B02___01_Mark________NPIDPIN1DA.wav"
    ident = MMSLanguageIdentifier()
    language = ident.identify(audioPath)
    print("language", language)




