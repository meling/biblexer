package bibtex

import "testing"

//TODO: Make tests that have space in entry type and between @ and entry type

// "@article", "pass",
// "@ article", "pass", Actually works with bibtex command
// "@art icle", "fail",

//TODO: Make test case that for some bibtex input
// compares against an expected sequence of itemType values.

var bibtext = `

@  article  {

	journals/tdsc/zorfu/AndersonMRVM15  ,
	author  ={Hein Meling}  ,
	title  =  {Some Title},

}

`

var expected = []itemType{
	itemEntryTypeDelim,
	itemEntryType,
	itemEntryStartDelim,
	itemCiteKey,
	itemTagDelim,
	itemTagName,
	itemTagNameContentDelim,
	itemTagContentStartDelim,
	itemTagContent,
	itemTagContentStopDelim,
	itemTagDelim,
	itemTagName,
	itemTagNameContentDelim,
	itemTagContentStartDelim,
	itemTagContent,
	itemTagContentStopDelim,
	itemTagDelim,
	itemEntryStopDelim,
	itemEOF,
}

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
	i := 0
	for it := l.nextItem(); it.typ != itemEOF; it = l.nextItem() {
		if it.typ != expected[i] {
			t.Errorf("Got %s, expected %s", it.String(), expected[i])
		}
		i++
		t.Log(it)
		// fmt.Println(it)
	}
}
