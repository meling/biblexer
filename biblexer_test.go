package bibtex

import (
	"fmt"
	"testing"
)

//TODO: Make tests that have space in entry type and between @ and entry type

// "@article", "pass",
// "@ article", "fail",
// "@art icle", "fail",

//TODO: Make test case that for some bibtex input
// compares against an expected sequence of itemType values.

var bibtext = `

@article  {

	journals/tdsc/zorfu/AndersonMRVM15  ,
	author  ={Hein Meling}  ,
	title  =  {Some Title},
}

`

// booktitle="hell yeah",
// testtitle="hell {So cool} yeah",
// note= "note from hell",

// @misc{mycitekey,
// 	author={Hein Meling},
// 	title={Some Title},
// 	booktitle="hell yeah",
// 	testtitle="hell {So cool} yeah",
// 	note= "note from hell",
// }

// `

func TestLexer(t *testing.T) {
	l := NewLexer("bib", bibtext)
	for it := l.nextItem(); it.typ != itemEOF; it = l.nextItem() {
		fmt.Println(it)
	}
}
