package tests

import (
	"strings"
	"testing"
)

const gordonFilter = `is_new: no
dataset_name: {dataset}_audio
bible_id: {bibleId}
username: GaryTest_15e
testament:
  nt: yes
compare:
  html_report: yes
  base_dataset: {dataset}
  gordon_filter: {threshold}
  compare_settings: 
    lower_case: y
    remove_prompt_chars: y
    remove_punctuation: y
    double_quotes: 
      remove: y
    apostrophe: 
      remove: y
    hyphen:
      remove: y
    diacritical_marks:
      normalize_nfd: y
`

func TestGordonFilter(t *testing.T) {
	var yaml = gordonFilter
	yaml = strings.Replace(yaml, "{dataset}", "N2CHF_TBL", 2)
	yaml = strings.Replace(yaml, "{bibleId}", "CHFTBL", 1)
	yaml = strings.Replace(yaml, "{threshold}", "1", 1)
	DirectSqlTest(yaml, []SqliteTest{}, t)
	yaml = gordonFilter
	yaml = strings.Replace(yaml, "{dataset}", "N2CUL_MNT", 2)
	yaml = strings.Replace(yaml, "{bibleId}", "CULMNT", 1)
	yaml = strings.Replace(yaml, "{threshold}", "1", 1)
	DirectSqlTest(yaml, []SqliteTest{}, t)
	yaml = gordonFilter
	yaml = strings.Replace(yaml, "{dataset}", "N2MDW_BSC", 2)
	yaml = strings.Replace(yaml, "{bibleId}", "MDWBSC", 1)
	yaml = strings.Replace(yaml, "{threshold}", "1", 1)
	DirectSqlTest(yaml, []SqliteTest{}, t)
	yaml = gordonFilter
	yaml = strings.Replace(yaml, "{dataset}", "N2IKP_MLT", 2)
	yaml = strings.Replace(yaml, "{bibleId}", "IKPMLT", 1)
	yaml = strings.Replace(yaml, "{threshold}", "1", 1)
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
