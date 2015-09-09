package bibtex

import (
	"fmt"
	"testing"
)

//TODO: Make tests that have space in entry type and between @ and entry type

// "@article", "pass",
// "@ article", "pass", Actually works with bibtex command
// "@art icle", "fail",

//TODO: Make bib entries  that should fail.

//TODO: Make test case that for some bibtex input
// compares against an expected sequence of itemType values.

var example = `@article{meling72, author = {Hein},}`

var passSet1 = []string{
	`@article{mycitekey1972,
		author = {Hein Meling},
		title = {The wonderful paper},
}`,
	`@article{mycitekey1972,
		author={Hein Meling},
		title={The wonderful paper},
}`,
	`@article{mycitekey1972,
		author={Hein Meling},
		title={The wonderful paper},
}`,
	`@article{mycitekey1972  ,
	author={Hein Meling},
	title={The wonderful paper},
}`,
	`@article  {mycitekey1972,
	author=   {Hein Meling},
	title={The wonderful paper},
}`,
	`@ article{mycitekey1972,
	author={Hein Meling},
	title={The wonderful paper},
}`,
	`@article{mycitekey1972,
	author={Hein Meling},
	title={The wonderful paper},

  }
`,
	`@article{mycitekey1972,
	author={Hein Meling},
	title={The wonderful paper},}`,
}

var passSet2 = []string{
	`@article{mycitekey1972,
	author={Hein Meling},
	title={The wonderful paper}}`,
	`@article{mycitekey1972,
	author={Hein Meling},
	title={The wonderful paper}
	}`,
}

var expectedSet1 = []itemType{
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

var expectedSet2 = []itemType{
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
	itemEntryStopDelim,
	itemEOF,
}

var fail = [...]string{
	`@article{mycitekey1972,
	aut  hor = {Hein Meling},
	title = {The wonderful paper},
}`,
	`@article{mycitekey with spaces 1972,
	author = {Hein Meling},
	title = {The wonderful paper},
}`,
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

func ExampleLexer() {
	l := NewLexer("bib", example)
	for it := l.nextItem(); it.typ != itemEOF; it = l.nextItem() {
		fmt.Print(it, " ")
	}
	// Output: "@" "article" "{" "meling72" "," "author" "=" "{" "Hein" "}" "," "}"
}

func TestLexer(t *testing.T) {
	doTest(t, passSet1, expectedSet1)
	doTest(t, passSet2, expectedSet2)
}

func doTest(t *testing.T, passSet []string, expectedSet []itemType) {
	for i := 0; i < len(passSet); i++ {
		l := NewLexer("bib", passSet[i])
		for j := 0; j < len(expectedSet); j++ {
			it := l.nextItem()
			if it.typ != expectedSet[j] {
				t.Errorf("Got %s, expected %s", it.String(), expectedSet[j])
			}
		}
	}
}
