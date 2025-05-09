import os
import numpy
import nltk
import uroman as ur
from flashlight.lib.text.dictionary import Dictionary, load_words, create_word_dict
from flashlight.lib.text.dictionary import pack_replabels
from flashlight.lib.text.decoder import LexiconDecoderOptions, LexiconDecoder, CriterionType
from flashlight.lib.text.decoder import KenLM
from flashlight.lib.text.decoder import Trie, SmearingMode
from sqlite_utility import *

"""
Flashlight text package
https://github.com/flashlight/text

Flashlight python text package (with beam search decoder)
https://github.com/flashlight/text/tree/main/bindings/python

Flashlight python data preparation instructions
https://github.com/flashlight/flashlight/tree/main/flashlight/app/asr#data-preparation

Flashlight - readthedocs.io
https://fl.readthedocs.io/en/latest/

Fairseq MMS zero shot
https://github.com/facebookresearch/fairseq/tree/main/examples/mms/zero_shot

beam search decoder tutorial
https://docs.pytorch.org/audio/main/tutorials/asr_inference_with_ctc_decoder_tutorial.html
"""

TOKEN_FILE = "tokens.txt"
LEXICON_FILE = "lexicon.txt"
SCRIPT_FILE = "script.txt"
MODEL_FILE = "model.arpa"
MODEL_BIN = "model.bin"

def clean_words(words, vocab):
    results = []
    urom = ur.Uroman()
    for word in words:
        result_wd = []
        for ch in word[1]:
            if ch.lower() not in vocab:
                ans = urom.romanize_string(ch)
                if ans in vocab:
                    use = ans
                elif ch == '.':
                    use = ''
                elif ch == '\u2212':
                    use = '-'
                elif ch == 'x':
                    use = 'z'
                else:
                    use = 'n'
                    print("no sub found for", ch)
                if use != '':
                    result_wd.append(use)
                    print("found", ch, "change to", use)
            else:
                result_wd.append(ch)
        results.append((word[0], "".join(result_wd)))
    return results

def create_tokens(vocab, directory):
    sorted_vocab = dict(sorted(vocab.items(), key=lambda item: item[1]))
    with open(os.path.join(directory, TOKEN_FILE), mode='w', encoding='utf-8') as file:
        for ch in sorted_vocab.keys():
            _ = file.write(ch + "\n")
        _ = file.write("<1>\n")
        _ = file.write("#\n")
        file.flush()
    return file.name

def create_lexicon(words, directory):
    word_set = set()
    for word in words:
        word_set.add(word[1].lower())
    word_set = sorted(word_set)
    with open(os.path.join(directory, LEXICON_FILE), mode='w', encoding='utf-8') as file:
        for word in word_set:
            _ = file.write(word + ' ')
            for ch in word:
                if ch == '-':
                    _ = file.write('| ')
                else:
                    _ = file.write(ch + ' ')
            _ = file.write('|\n')
        file.flush()
        print(file.name)
    return file.name

def create_script(words, directory):
    first = True
    curr_script_id = words[0][0]
    with open(os.path.join(directory, SCRIPT_FILE), mode='w', encoding='utf-8') as file:
        for (script_id, word) in words:
            if script_id != curr_script_id:
                _ = file.write('\n')
                curr_script_id = script_id
            elif not first:
                _ = file.write(' ')
            _ = file.write(word.lower())
            first = False
        _ = file.write('\n')
        file.flush()
    return file.name

def tkn_to_idx(spelling: list, token_dict : Dictionary, maxReps : int = 0):
    result = []
    for token in spelling:
        result.append(token_dict.get_index(token))
    return pack_replabels(result, token_dict, maxReps)

def create_decoder(db_path, vocab, directory):
    database = SqliteUtility(db_path)
    words = database.select("SELECT script_id, word FROM words WHERE ttype='W'", ())
    database.close()

    words2 = clean_words(words, vocab)
    tokenFile = create_tokens(vocab, directory)
    lexiconFile = create_lexicon(words2, directory)
    scriptFile = create_script(words2, directory)

    token_dict = Dictionary(tokenFile)
    lexicon = load_words(lexiconFile)
    word_dict = create_word_dict(lexicon)

    modelFile = os.path.join(directory, MODEL_FILE)
    modelBin = os.path.join(directory, MODEL_BIN)
    os.system("kenlm/build/bin/lmplz -o 5 < " + scriptFile + " > " + modelFile)
    os.system("kenlm/build/bin/build_binary " + modelFile + " " + modelBin)

    #lm = KenLM(modelBin, word_dict)
    lm = KenLM(modelFile, word_dict)

    sil_idx = token_dict.get_index("|")

    trie = Trie(token_dict.index_size(), sil_idx)
    start_state = lm.start(False)

    for word, spellings in lexicon.items():
        usr_idx = word_dict.get_index(word)
        _, score = lm.score(start_state, usr_idx)
        for spelling in spellings:
            # convert spelling string into vector of indices
            spelling_idxs = tkn_to_idx(spelling, token_dict, 1)
            trie.insert(spelling_idxs, usr_idx, score)
        trie.smear(SmearingMode.MAX) # propagate word score to each spelling node to have some lm proxy score in each node.
    print("Finished building Trie")

    options = LexiconDecoderOptions(
        beam_size=500,         # range 25-500, default 50-100, 200-500 high accuracy
        beam_size_token=50,    # default 25-50, large token sets 50-100, restrict num tokens at each step
        beam_threshold=25.0,   # default 15-25, aggressive pruning 5-10, 25 common
        lm_weight=2.69,        # LLM influence, default 1-2, typical 0.5-3, 2.69 common
        word_score=2.8,        # Pos encourages word insertion, default 0 to -1, typical -3 to 3, 2.8 common
        unk_score=-5.0,        # -Inf(no unknown words) to -5, less restrictive -2 to -3
        sil_score=0.0,         # Silence tokens, default 0, typical -0.5 to 0.5
        log_add=False,         # default false, Use max instead of log-add
        criterion_type=CriterionType.CTC  # For CTC-based models
    )

    blank_idx = token_dict.get_index("#") # for CTC
    unk_idx = word_dict.get_index("<unk>")
    #transitions = numpy.zeros((token_dict.index_size(), token_dict.index_size()),) # for ASG fill up with correct values
    transitions = [0.0] * (token_dict.index_size() * token_dict.index_size())
    is_token_lm = False # we use word-level LM

    decoder = LexiconDecoder(
        options,
        trie,
        lm,
        sil_idx,
        blank_idx,
        unk_idx,
        transitions,
        is_token_lm
    )
    return decoder
    # emissions is numpy.array of emitting model predictions with shape [T, N], where T is time, N is number of tokens
    #results = decoder.decode(emissions.ctypes.data, T, N)
    # results[i].tokens contains tokens sequence (with length T)
    # results[i].score contains score of the hypothesis
    # results is sorted array with the best hypothesis stored with index=0.

if __name__ == "__main__":
    from transformers import AutoProcessor
    dbPath = os.path.join(os.getenv('FCBH_DATASET_DB'), 'GaryNTest', 'N2CUL_MNT.db')
    model_id = "facebook/mms-1b-all"
    processor = AutoProcessor.from_pretrained(model_id)
    lang = 'cul'
    processor.tokenizer.set_target_lang(lang)
    vocab = processor.tokenizer.get_vocab()
    directory = 'data'
    decoder = create_decoder(dbPath, vocab, directory)
    print(decoder)