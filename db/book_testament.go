package db

import (
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
)

var BookOT = []string{`GEN`, `EXO`, `LEV`, `NUM`, `DEU`, `JOS`, `JDG`, `RUT`, `1SA`, `2SA`, `1KI`, `2KI`,
	`1CH`, `2CH`, `EZR`, `NEH`, `EST`, `JOB`, `PSA`, `PRO`, `ECC`, `SNG`, `ISA`, `JER`, `LAM`, `EZK`, `DAN`,
	`HOS`, `JOL`, `AMO`, `OBA`, `JON`, `MIC`, `NAM`, `HAB`, `ZEP`, `HAG`, `ZEC`, `MAL`, `COV`}
var BookNT = []string{`MAT`, `MRK`, `LUK`, `JHN`, `ACT`, `ROM`, `1CO`, `2CO`, `GAL`, `EPH`, `PHP`, `COL`,
	`1TH`, `2TH`, `1TI`, `2TI`, `TIT`, `PHM`, `HEB`, `JAS`, `1PE`, `2PE`, `1JN`, `2JN`, `3JN`, `JUD`, `REV`}

func RequestedBooks(testament request.Testament) []string {
	var results []string
	if testament.OT {
		results = append(results, BookOT...)
	} else if len(testament.OTBooks) > 0 {
		results = append(results, testament.OTBooks...)
	}
	if testament.NT {
		results = append(results, BookNT...)
	} else if len(testament.NTBooks) > 0 {
		results = append(results, testament.NTBooks...)
	}
	return results
}
