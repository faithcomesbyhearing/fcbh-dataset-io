from flashlight.lib.text.dictionary import Dictionary, load_words, create_word_dict
from flashlight.lib.text.dictionary import pack_replabels
from flashlight.lib.text.decoder import KenLM
from flashlight.lib.text.decoder import Trie, SmearingMode
from sqlite_utility import *


def create_tokens_dict(words):
    char_set = set()
    for word in words:
        for ch in word[0]:
            char_set.add(ch.lower())
    char_set = sorted(char_set)
    with open("data/tokens.txt", mode='w', encoding='utf-8') as file:
        _ = file.write("|\n")
        for ch in char_set:
            _ = file.write(ch + "\n")
        file.flush()
    tokens_dict = Dictionary(file.name)
    return tokens_dict

def create_lexicon_word_dict(words):
    word_set = set()
    for word in words:
        word_set.add(word[0].lower())
    word_set = sorted(word_set)
    with open("data/lexicon.txt", mode='w', encoding='utf-8') as file:
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
    lexicon = load_words(file.name)
    word_dict = create_word_dict(lexicon)
    return lexicon, word_dict

dbPath = os.getenv('FCBH_DATASET_DB')+ '/GaryNTest/N2CUL_MNT.db'
database = SqliteUtility(dbPath)
words = database.select("SELECT word FROM words WHERE ttype = 'W'",())
database.close()

tokens_dict = create_tokens_dict(words)

word_dict, lexicon = create_lexicon_word_dict(words)

==== right here

lm = KenLM("path/lm.arpa", word_dict) # or "path/lm.bin"

sil_idx = tokens_dict.get_index("|")

trie = Trie(token_dict.index_size(), sil_idx)
start_state = lm.start(False)

def tkn_to_idx(spelling: list, token_dict : Dictionary, maxReps : int = 0):
    result = []
    for token in spelling:
        result.append(token_dict.get_index(token))
    return pack_replabels(result, token_dict, maxReps)


for word, spellings in lexicon.items():
    usr_idx = word_dict.get_index(word)
    _, score = lm.score(start_state, usr_idx)
    for spelling in spellings:
        # convert spelling string into vector of indices
        spelling_idxs = tkn_to_idx(spelling, token_dict, 1)
        trie.insert(spelling_idxs, usr_idx, score)

    trie.smear(SmearingMode.MAX) # propagate word score to each spelling node to have some lm proxy score in each node.




options = LexiconDecoderOptions(
    beam_size=500,         # range 25-500, default 50-100, 200-500 high accuracy
    token_beam_size=50,    # default 25-50, large token sets 50-100, restrict num tokens at each step
    beam_threshold=25,     # default 15-25, aggressive pruning 5-10, 25 common
    lm_weight=2.69,        # LLM influence, default 1-2, typical 0.5-3, 2.69 common
    word_score=2.8,        # Pos encourages word insertion, default 0 to -1, typical -3 to 3, 2.8 common
    unk_score=-5.0,        # -Inf(no unknown words) to -5, less restrictive -2 to -3
    sil_score=0.0,         # Silence tokens, default 0, typical -0.5 to 0.5
    log_add=False,         # default false, Use max instead of log-add
    criterion_type=CriterionType.CTC  # For CTC-based models
)

import numpy
from flashlight.lib.text.decoder import LexiconDecoder


# To use KenLM
from flashlight.lib.text.decoder import KenLM
lm = KenLM("path/lm.arpa", word_dict) # or "path/lm.bin"
sil_idx = token_dict.get_index("|")
blank_idx = token_dict.get_index("#") # for CTC
unk_idx = word_dict.get_index("<unk>")
transitions = numpy.zeros((token_dict.index_size(), token_dict.index_size()) # for ASG fill up with correct values
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
# emissions is numpy.array of emitting model predictions with shape [T, N], where T is time, N is number of tokens
results = decoder.decode(emissions.ctypes.data, T, N)
# results[i].tokens contains tokens sequence (with length T)
# results[i].score contains score of the hypothesis
# results is sorted array with the best hypothesis stored with index=0.