package bibtex

import (
	"fmt"
	"strings"
)

// itemType identifies the type of lex items.
type itemType int

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	}
	return fmt.Sprintf("%q", i.val)
}

const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemComment              // delimiter for comments (%)
	itemEntryTypeDelim       // entry type delimiter (@)
	itemEntryType            // the entry type
	itemEntryStartDelim      // entry start delimiter ({)
	itemEntryStopDelim       // entry stop delimiter (})
	itemCiteKey              // the cite key
	itemTagName              // the tag name (on left of =)
	itemTagNameContentDelim  // delimiter separating name and content (=)
	itemTagContent           // the content for the tag
	itemTagDelim             // delimiter separating name-content pairs or tags (,)
	itemTagContentStartDelim // content start delimiter ({)
	itemTagContentStopDelim  // content stop delimiter (})
	itemTagContentQuoteDelim // content start/stop delimiter (")
	itemConcat               // the concatination symbol (#)
)

var key = map[string]itemType{
	"citekey": itemCiteKey,
	"@":       itemEntryTypeDelim,
	"{":       itemEntryStartDelim,
	"}":       itemEntryStopDelim,
	"#":       itemConcat,
	",":       itemTagDelim,
	"=":       itemTagNameContentDelim,
}

//TODO: Add support for ignoring comments later
//TODO: Add support for @string, @preamble, @comment
//TODO: Add support for concatination #
//TODO: Add support for quote-based ("") content
//TODO: Add support for braces in content

// state functions

const (
	commentDelim = "%"
)

// lexStart scans the input for bibtex entries.
// lexStart scans until an entry type delimiter "@" is found, and
// starts to process the rest of the bibtex entries in the input.
func lexStart(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], "@") {
			if l.pos > l.start {
				// ignore anything that comes before the @ delimiter.
				l.ignore()
			}
			l.emit1(itemEntryTypeDelim) // absorb '@'
			return lexEntryType
		}
		if l.next() == eof {
			l.emit(itemEOF)
			return nil
		}
	}
}

// lexEntryType scans the entry type.
func lexEntryType(l *lexer) stateFn {
	l.ignoreSpaces()
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb and emit when delimiter is found
		case r == '{':
			l.backup()
			l.emit(itemEntryType)
			l.emit1(itemEntryStartDelim) // absorb '{'
			return lexCiteKey
		case isSpace(r):
			// discard spaces after entry type (to avoid emitting with spaces)
			l.discard()
		case r == eof:
			return l.errorf("unexpected eof at line %d", l.lineNumber())
		default:
			return l.errorf("unexpected character %#U at line %d", r, l.lineNumber())
		}
	}
}

// lexCiteKey scans the cite key.
func lexCiteKey(l *lexer) stateFn {
	l.ignoreSpaces()
	for {
		switch r := l.next(); {
		case l.isUnbrokenAlphaNumericToken(r):
			// absorb and emit when delimiter is found
		case r == ',':
			l.backup()
			l.emit(itemCiteKey)
			l.emit1(itemTagDelim) // absorb ','
			return lexTagName
		case isSpace(r):
			// discard spaces after cite key (to avoid emitting with spaces)
			l.discard()
		case r == eof:
			return l.errorf("unexpected eof at line %d", l.lineNumber())
		default:
			return l.errorf("unexpected character %#U at line %d", r, l.lineNumber())
		}
	}
}

// lexTagName scans the tag name.
func lexTagName(l *lexer) stateFn {
	l.ignoreSpaces()
	for {
		if strings.HasPrefix(l.input[l.pos:], "}") {
			l.emit1(itemEntryStopDelim) // absorb '}'
			// search for the next bib entry
			return lexStart
		}
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb and emit when delimiter is found
		case r == '=':
			l.backup()
			l.emit(itemTagName)
			l.emit1(itemTagNameContentDelim) // absorb '='
			return lexContentStartDelim
		case isSpace(r):
			// discard spaces after tag name (to avoid emitting with spaces)
			l.discard()
		case r == eof:
			return l.errorf("unexpected eof at line %d", l.lineNumber())
		default:
			return l.errorf("unexpected character %#U at line %d", r, l.lineNumber())
		}
	}
}

// lexContentStartDelim scans the name-content start delimiter.
func lexContentStartDelim(l *lexer) stateFn {
	l.ignoreSpaces()
	for {
		switch r := l.next(); {
		case r == '{':
			l.emit(itemTagContentStartDelim)
			return lexTagContent
		case r == eof:
			return l.errorf("unexpected eof at line %d", l.lineNumber())
		default:
			return l.errorf("unexpected character %#U at line %d", r, l.lineNumber())
		}
	}
}

// lexTagContent scans the elements inside the content.
func lexTagContent(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r) || isSpace(r):
			// absorb and emit when delimiter is found
		case r == '}':
			l.backup()
			l.emit(itemTagContent)
			l.emit1(itemTagContentStopDelim) // absorb '}'
			return lexTagDelim
		case r == eof:
			return l.errorf("unexpected eof at line %d", l.lineNumber())
		default:
			return l.errorf("unexpected character %#U at line %d", r, l.lineNumber())
		}
	}
}

// lexTagDelim scans the tag delimiter.
func lexTagDelim(l *lexer) stateFn {
	l.ignoreSpaces()
	for {
		switch r := l.next(); {
		case r == ',':
			l.emit(itemTagDelim)
			return lexTagName
		case r == eof:
			return l.errorf("unexpected eof at line %d", l.lineNumber())
		default:
			return l.errorf("unexpected character %#U at line %d", r, l.lineNumber())
		}
	}
}
