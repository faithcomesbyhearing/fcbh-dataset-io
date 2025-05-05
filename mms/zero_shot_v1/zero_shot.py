# Import additional libraries
import kenlm
from utils.lm import create_unigram_lm, maybe_generate_pseudo_bigram_arpa

MODEL_ID = "mms-meta/mms-zeroshot-300m"

processor = AutoProcessor.from_pretrained(MODEL_ID)
model = Wav2Vec2ForCTC.from_pretrained(MODEL_ID)

# Modified transcribe_audio function with KenLM support
def transcribe_audio(audio_file_path, words_file_path, use_lm=True):
    # Load audio
    audio_samples, _ = librosa.load(audio_file_path, sr=16000, mono=True)

    # Process audio with model
    inputs = processor(
        audio_samples, sampling_rate=16000, return_tensors="pt"
    )

    # Get model output
    with torch.no_grad():
        outputs = model(**inputs).logits

    # Load words and create word counts
    word_counts = {}
    num_sentences = 0
    with open(words_file_path, 'r') as f:
        lines = f.readlines()
        num_sentences = len(lines)
        for line in lines:
            for word in line.strip().split():
                word_counts.setdefault(word, 0)
                word_counts[word] += 1

    # Create lexicon
    lexicon = uromanize(list(word_counts.keys()))

    # Create language model if requested
    lm_path = None
    wscore = None
    lmscore = None

    if use_lm:
        tmp_file = tempfile.NamedTemporaryFile()
        lm_path = tmp_file.name
        create_unigram_lm(word_counts, num_sentences, lm_path)
        maybe_generate_pseudo_bigram_arpa(lm_path)

        # Default scores when using LM
        wscore = -0.18  # WORD_SCORE_DEFAULT_IF_LM
        lmscore = 1.48  # LM_SCORE_DEFAULT
    else:
        # Default score when not using LM
        wscore = -3.5  # WORD_SCORE_DEFAULT_IF_NOLM
        lmscore = 0

    # Create the CTC decoder with language model
    with tempfile.NamedTemporaryFile() as lexicon_file:
        with open(lexicon_file.name, "w") as f:
            for word, spelling in lexicon.items():
                f.write(word + " " + spelling + "\n")

        # Create beam search decoder
        beam_search_decoder = ctc_decoder(
            lexicon=lexicon_file.name,
            tokens=token_file,
            lm=lm_path,
            beam_size=500,
            beam_size_token=50,
            lm_weight=lmscore,
            word_score=wscore,
            blank_token="<s>",
        )

        # Decode the output
        beam_search_result = beam_search_decoder(outputs)
        transcription = " ".join(beam_search_result[0][0].words).strip()

    return transcription