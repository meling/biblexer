package biblexer

import (
	"fmt"
	"strings"
)

//TODO: Add support for ignoring comments later
//TODO: Add support @preamble, @comment

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

//go:generate stringer -type=itemType
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
	itemEqual                // delimiter separating name and content (=)
	itemTagContent           // the content for the tag
	itemTagDelim             // delimiter separating name-content pairs or tags (,)
	itemTagContentStartDelim // content start delimiter ({)
	itemTagContentStopDelim  // content stop delimiter (})
	itemQuoteDelim           // content start/stop delimiter (")
	itemConcat               // the concatination symbol (#)
	itemStringKey            // string macro key
)

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
		case l.isUnbrokenAlphaNumericToken(r):
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
		case r == '=': // @string macro support
			l.backup()
			l.emit(itemStringKey)
			l.emit1(itemEqual) // absorb '='
			return lexTagContentStartDelim
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
		case l.isUnbrokenAlphaNumericToken(r):
			// absorb and emit when delimiter is found
		case r == '=':
			l.backup()
			l.emit(itemTagName)
			l.emit1(itemEqual) // absorb '='
			return lexTagContentStartDelim
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

// lexTagContentStartDelim scans the name-content start delimiter.
func lexTagContentStartDelim(l *lexer) stateFn {
	l.ignoreSpaces()
	for {
		switch r := l.next(); {
		case l.isUnbrokenAlphaNumericToken(r):
			// absorb and emit when delimiter is found
		case r == '"':
			l.emit(itemQuoteDelim)
			return lexTagContent
		case r == '{':
			l.emit(itemTagContentStartDelim)
			return lexTagContent
		case r == '}':
			// handle @string key as the last element of a tag
			l.backup()
			l.emit(itemStringKey)
			return lexTagName
		case r == '#': // Concatination support for @string macros
			l.backup()
			l.emit(itemStringKey)
			l.emit1(itemConcat) // absorb '#'
			return lexTagContentStartDelim
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

// lexTagContent scans the elements inside the content.
func lexTagContent(l *lexer) stateFn {
	l.ignoreSpaces()
	braces := 0
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r) || isSpace(r):
			// absorb and emit when delimiter is found
		case r == '{':
			braces++
			// absorb internal brace
		case r == '}' && braces > 0:
			braces--
			// absorb internal brace
		case r == '"':
			l.backup()
			l.emit(itemTagContent)
			l.emit1(itemQuoteDelim) // absorb '"'
			return lexTagDelim
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
		case r == '}':
			// handle last name-content pair without ',' delimiter in lexTagName
			l.backup()
			return lexTagName
		case r == '#': // Concatination support for content strings
			l.backup()
			// l.emit(itemTagContent) //TODO: This was a bug in one case, check others
			l.emit1(itemConcat) // absorb '#'
			return lexTagContentStartDelim
		case r == eof:
			return l.errorf("unexpected eof at line %d", l.lineNumber())
		default:
			return l.errorf("unexpected character %#U at line %d", r, l.lineNumber())
		}
	}
}
