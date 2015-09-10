package bibtex

import (
	"fmt"
	"testing"
)

// passSet1 contains entries that all should produce
// the same sequence of tokens, given by expectedSet1.
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
	`@article{mycitekey1972,
	author={   Hein Meling},
	title={The wonderful paper},}`,
	`@article{mycitekey1972,
	author={Hein Meling   },
	title={The wonderful paper},}`,
	`@article{mycitekey1972,
	author={   Hein Meling   },
	title={The wonderful paper},}`,
	`@article{mycitekey1972,
		author = {Hein Meling},
		title = {{The wonderful paper}},
}`,
	`@article{mycitekey1972,
	author = {Hein Meling},
	title = {{{Another} {wonderful} paper}},
}`,
}

// expectedSet1 is the sequence of tokens expected for each entry in passSet1.
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

// passSet2 contains entries that all should produce
// the same sequence of tokens, given by expectedSet2.
var passSet2 = []string{
	`@article{mycitekey1972,
	author={Hein Meling},
	title={The wonderful paper}}`,
	`@article{mycitekey1972,
	author={Hein Meling},
	title={The wonderful \TeX paper}
	}`,
	`@article{mycitekey1972,
	author={Hein Meling},
	title={The wonderful paper}

	}`,
}

// expectedSet2 is the sequence of tokens expected for each entry in passSet2.
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
	itemEntryStopDelim, // entry stop delimiter '}' without preceeding ','
	itemEOF,
}

var passSet3 = []string{
	`@article{citekey72,
		author = "Hein Meling",
		title = "The Greatest Gopher Paper",
	}`,
	`@article{citekey72,
		author = "Hein Meling",
		title = "The Greatest {Gopher} Paper",
	}`,
	`@article{citekey72,
		author = "Hein Meling",
		title = "The Greatest Paper"  ,
	}`,
	`@article{citekey72,
		author = "Hein Meling",
		title = "The {Greatest} Paper"  ,
	}`,
	`@article{citekey72,
		author = "Hein Meling",
		title = "{Another} Even {Greater} Paper"  ,
	}`,
	`@article{citekey72,
		author = "Hein Meling",
		title = "{{Another} Even {Greater} Than The {Greatest Paper}}"  ,
	}`,
}

// expectedSet3 is the sequence of tokens expected for each entry in passSet3.
var expectedSet3 = []itemType{
	itemEntryTypeDelim,
	itemEntryType,
	itemEntryStartDelim,
	itemCiteKey,
	itemTagDelim,
	itemTagName,
	itemTagNameContentDelim,
	itemTagContentQuoteDelim,
	itemTagContent,
	itemTagContentQuoteDelim,
	itemTagDelim,
	itemTagName,
	itemTagNameContentDelim,
	itemTagContentQuoteDelim,
	itemTagContent,
	itemTagContentQuoteDelim,
	itemTagDelim,
	itemEntryStopDelim,
	itemEOF,
}

var passSet4 = []string{
	`@article{c72, author = "Mrs. Gopher" # " and Mr. Pike"}`,
	`@string{ gopher = "Mrs. Gopher"	}
	 @article{c72, author = gopher #" and Mr. Pike"}`,
	`@string{ gopher = "Mrs. Gopher" }
	 @string{ pike = "Mr. Gopher"	}
	 @article{c72, author = gopher # pike # " and Mr. Micky"}`,
	// `@string{ gopher = "Mrs. Gopher" }
	//  @string{ pike = "Mr. Gopher"	}
	//  @article{c72, author = gopher # pike }`, // KNOWN ISSUE: CANNOT END WITH STRING MACRO
}

var failSet = [...]string{
	`@article{mycitekey1972,
	aut  hor = {Hein Meling},
	title = {The wonderful paper},
}`,
	`@article{mycitekey with spaces 1972,
	author = {Hein Meling},
	title = {The wonderful paper},
}`,
	`@art  icle{mycitekey1972,
  author = {Hein Meling},
  title = {The wonderful paper},
}`,
}

func ExampleLexer() {
	l := newLexer("bib", "@article{meling72, author = {Hein},}")
	for it := l.nextItem(); it.typ != itemEOF; it = l.nextItem() {
		fmt.Print(it, " ")
	}
	// Output: "@" "article" "{" "meling72" "," "author" "=" "{" "Hein" "}" "," "}"
}

func ExampleFailingLexer() {
	l := newLexer("bib", failSet[1])
	for it := l.nextItem(); it.typ != itemEOF; it = l.nextItem() {
		fmt.Print(it, " ")
	}
	// Output: "@" "article" "{" unexpected character U+0077 'w' at line 1
}

func TestLexer(t *testing.T) {
	doTest(t, passSet1, expectedSet1)
	doTest(t, passSet2, expectedSet2)
	doTest(t, passSet3, expectedSet3)
}

func TestLexerExperimental(t *testing.T) {
	for i := 0; i < len(passSet4); i++ {
		l := newLexer("bib", passSet4[i])
		for it := l.nextItem(); it.typ != itemEOF; it = l.nextItem() {
			fmt.Println(it)
		}
	}
}

func TestFailingLexer(t *testing.T) {
	for i := 0; i < len(failSet); i++ {
		l := newLexer("bib", failSet[i])
		it := l.nextItem()
		for it.typ != itemEOF && it.typ != itemError {
			it = l.nextItem()
		}
		if it.typ != itemError {
			t.Errorf("Got %s, expected %s", it.typ, itemError)
		}
	}
}

func doTest(t *testing.T, passSet []string, expectedSet []itemType) {
	for i := 0; i < len(passSet); i++ {
		l := newLexer("bib", passSet[i])
		for j := 0; j < len(expectedSet); j++ {
			it := l.nextItem()
			// fmt.Println(it)
			if it.typ != expectedSet[j] {
				t.Errorf("Got %s, expected %s", it, expectedSet[j])
			}
		}
	}
}
