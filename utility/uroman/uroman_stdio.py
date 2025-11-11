import sys
import os
import uroman as ur
sys.path.insert(0, os.path.abspath(os.path.join(os.environ['GOPROJ'], 'logger')))
from error_handler import setup_error_handler

# Uroman
# https://aclanthology.org/P18-4003.pdf
# Any Numeric Digits are converted to ascii 0-9
# Punctuation is preserved, and changed to ascii or other roman
# Case is not changes
# Diacritical marks by themselves are ignored
# I observed cases where a diacritical mark in NFD for was ignored,
# but one in NFC format was used

setup_error_handler()
uroman = ur.Uroman()
for line in sys.stdin:
    output = uroman.romanize_string(line)
    sys.stdout.write(output)
    sys.stdout.flush()

# conda activate mms_fa
# python3 uroman.py
# hello world

