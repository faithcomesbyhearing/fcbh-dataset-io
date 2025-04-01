## FA_error_analysis

March 31, 2025
Gary Griswold

Analyzing FA errors in N2CUL/MNT in known_error_analysis.go has produced disappointing 
results.  When looking at fa_error by word, the known errors all end up in the 250-350th
if one sequenced the verses by worst fa_error.  

One could produce a better result by aggragating all of the errors in a line to
produce a larger sort score when there were multiple words with bad scores.  But,
this would obscure single word errors.

Possibly the right solution is to remove false positives using an algorithm similar to
Gordon's.

1. First convert each word score into a pattern that could be repeatable.
   2. Convert each score to an error by -log10.
   3. Round the errors to a integer, which should be one digit from 0 to 7
   4. Convert this errors to a string, which should result in a string the same length as the word
   5. Append the error string to the word to create a pattern.
6. Search for the occurrences of the pattern among apparent errors
   7. Iterate over the entire text.
   8. For those words that have a low score, such as >= 1.0,
   8. Converting words to this pattern.
   9. build a map whose key is the pattern, and whose value is a list of the location of each occurrance
   10. the location might be script_id, char_pos, or book_id, chapter, verse, char_pos
11. Identify those patterns that are likely to be false positives because they occur many times
    12. Iterate over the map of errors
    13. For those that appear to be false positives because they occur >= 100 times
    14. Iterate over the list found,
    15. Mark each word false_positive=true.
16. When generating a report of errors, ignore the false positives.
