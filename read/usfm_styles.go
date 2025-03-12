package read

import (
	"context"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

func include(style string) bool {
	last := style[len(style)-1]
	if last >= '0' && last <= '9' {
		style = style[:len(style)-1]
	}
	answer, ok := usfm[style]
	if !ok {
		log.Warn(context.Background(), "USFM map does not have entry: ", style)
	}
	return answer
}

var usfm = map[string]bool{
	`book.id`:   false,
	`chapter.c`: false,
	`verse.v`:   false,
	// Identification
	`para.ide`:  false, // file encoding
	`para.h`:    true,  // heading
	`para.toc`:  false, // table of contents
	`para.toca`: false, // alt language table of contents
	`para.rem`:  false, // comments
	`para.usfm`: false, // usfm markup version
	// Introductions
	`para.imt`:  false, // introduction major title
	`para.is`:   false, // introduction section heading
	`para.ip`:   false, // introduction paragraph
	`para.ipi`:  false, // indented introduction paragraph
	`para.im`:   false, // introduction flush let
	`para.imi`:  false, // indented introduction flush left
	`para.ipq`:  false, // introduction quote from scripture
	`para.imq`:  false, // introduction flush elft quote from scripture
	`para.ipr`:  false, // introduction right aligned
	`para.iq`:   false, // introduction poetic line
	`para.ib`:   false, // introduction blank line
	`para.ili`:  false, // introduction list item
	`para.iot`:  false, // introduction outline title
	`para.io`:   false, // introduction outline entry
	`para.iex`:  false, // introduction explanatory or bridge text
	`para.imte`: false, // introduction major title ending
	`para.ie`:   false, // introduction end
	// Titles & Headings
	`para.mt`:  true,  // main title
	`para.mte`: false, // main title at introduction ending
	`para.cl`:  false, // chapter label
	`para.cd`:  false, // chapter description
	`para.ms`:  false, // major section heading
	`para.mr`:  false, // major section reference range
	`para.s`:   false, // section heading
	`para.sr`:  false, // section reference
	`para.r`:   false, // parallel passage reference
	`para.d`:   false, // descriptive hebrew title
	`para.sp`:  false, // speaker identification
	`para.sd`:  false, // semantic division
	// Paragraphs
	`para.p`:       true,  // normal paragraph
	`para.m`:       true,  // margin paragraph
	`para.po`:      true,  // opening of an epistle
	`para.pr`:      true,  // right aligned paragraph
	`para.cls`:     true,  // closure of an epistle
	`para.pmo`:     true,  // embedded text opening
	`para.pm`:      true,  // embedded text paragraph
	`para.pmc`:     true,  // embedded text closing
	`para.pmr`:     true,  // embedded text refrain
	`para.pi`:      true,  // indented paragraph
	`para.mi`:      true,  // indented flush left paragraph
	`para.pc`:      true,  // centered paragraph
	`para.ph`:      true,  // indented paragraph with hanging indent, deprecated
	`para.lit`:     true,  // liturgical note / comment
	`para.nb`:      true,  // paragraph with no break from prior paragraph
	`para.pb`:      true,  // page break
	`para.cp`:      false, // published chapter number
	`para.restore`: false, // comment about restored text
	// Poetry
	`para.q`:  true,  // poetic line
	`para.qr`: true,  // right aligned poetic line
	`para.qc`: true,  // centered poetic line
	`para.qa`: false, // acrostic heading
	`para.qm`: true,  // embedded text poetic line
	`para.qd`: false, // Hebrew note - false?
	`para.b`:  false, // blank line
	// Lists
	`para.lh`:   true, // list header
	`para.li`:   true, // list entry
	`para.lf`:   true, // list footer
	`para.lim`:  true, // embedded list entry
	`para.litl`: true, // list entry total
	// Table
	`row.tr`:   true,
	`cell.th`:  true,
	`cell.thr`: true,
	`cell.tc`:  true,
	`cell.tcr`: true,
	// Identification
	`char.va`: false, // second/alternate verse number
	`char.vp`: false, // published verse marker
	`char.ca`: false, // second/alternate chapter number
	// Special Text
	`char.add`:   false, // translator’s addition
	`char.addnp`: false, // chinese to be dotted, underlined, deprecated
	`char.bk`:    true,  // quoted book title
	`char.dc`:    false, // deuterocanonical additions
	`char.ior`:   false, // introduction outline reference range
	`char.iqt`:   false, // introduction quoted text
	`char.k`:     true,  // keyword
	`char.litl`:  true,  // list entry total
	`char.nd`:    true,  // name of God
	`char.ord`:   true,  // ordinal number ending
	`char.pn`:    true,  // proper name
	`char.png`:   true,  // geographic proper name
	`char.qac`:   true,  // used to indicate acrostic letter in poetic line
	`char.qs`:    true,  // selah
	`char.qt`:    true,  // quoted text
	`char.rq`:    false, // inline quotation reference
	`char.sig`:   true,  // signature of the author of an epistle
	`char.sls`:   true,  // passage of text based on secondary language or alt source
	`char.tl`:    true,  // transliterated words
	`char.wj`:    true,  // words of Jesus
	// Character Styling
	`char.em`:   true, // emphasis
	`char.bd`:   true, // bold
	`char.bdit`: true, // bold + italic
	`char.it`:   true, // italic
	`char.no`:   true, // normal
	`char.sc`:   true, // small cap
	`char.sup`:  true, // subscript
	// Special Features
	`char.rb`:  true,  // Ruby glossing
	`char.pro`: true,  // pronunciation information
	`char.w`:   true,  // wordlist/glossary
	`char.wg`:  true,  // Greek word
	`char.wh`:  true,  // Hebrew word
	`char.wa`:  true,  // Aramaic word
	`char.fig`: false, // DNALAI
	// Structured List Entries
	`char.lik`: true, // list entry key content
	`char.liv`: true, // list entry value content
	// Linking
	`char.jmp`: true, // available for linking
	// Note
	`note.f`:      false, // footnote
	`note.fe`:     false, // end note
	`note.ef`:     false, // extended study note
	`note.x`:      false, // cross reference
	`note.ex`:     false, // extended cross reference
	`sidebar.esb`: false,
	`figure.fig`:  false,
	// optbreak no style
	// MS Section
	`ms.qt`: false, // quotation speaker
	`ms.ts`: false, // translators section
	// this might also be para.ts
}
