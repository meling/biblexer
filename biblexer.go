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
	itemComment             // delimiter for comments (%)
	itemEntryTypeDelim      // entry type delimiter (@)
	itemEntryType           // the entry type
	itemEntryStartDelim     // entry start delimiter ({)
	itemEntryStopDelim      // entry stop delimiter (})
	itemCiteKey             // the cite key
	itemTagName             // the tag name (on left of =)
	itemTagNameContentDelim // delimiter separating name and content (=)
	itemTagContent          // the content for the tag
	itemTagDelim            // delimiter separating name-content pairs or tags (,)
	itemContentStartDelim   // content start delimiter ({)
	itemContentStopDelim    // content stop delimiter (})
	itemContentQuoteDelim   // content start/stop delimiter (")
	itemConcat              // the concatination symbol (#)
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

//TODO: Rename lex-functions to use the new bibtex terminology
//TODO: Add support for ignoring comments later
//TODO: Add support for @string, @preamble, @comment
//TODO: Add support for concatination #
//TODO: Add support for quote-based ("") content
//TODO: Add support for braces in content

// state functions

const (
	commentDelim = "%"
)

// lex scans the input until an entry type delimiter, "@".
func lex(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], "@") {
			if l.pos > l.start {
				// ignore anything that comes before the @ delimiter.
				l.ignore()
			}
			return lexEntryTypeDelim
		}
		if l.next() == eof {
			l.emit(itemEOF)
			return nil
		}
	}
}

// lexEntryTypeDelim scans the entry type delimiter, which is known to be present.
func lexEntryTypeDelim(l *lexer) stateFn {
	l.emit1(itemEntryTypeDelim) // absorb '@'
	return lexEntryType
}

// lexEntryType scans the entry type.
func lexEntryType(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		case r == '{':
			l.backup()
			l.emit(itemEntryType)
			return lexEntryStartDelim
		case isSpace(r):
			// discard spaces after entry type (to avoid emitting with spaces)
			l.discard()
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexEntryStartDelim scans the entry delimiter, which is known to be present.
func lexEntryStartDelim(l *lexer) stateFn {
	l.emit1(itemEntryStartDelim) // absorb '{'
	return lexCiteKey
}

// lexCiteKey scans the cite key.
func lexCiteKey(l *lexer) stateFn {
	l.ignoreSpaces()
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		case r == ',':
			l.backup()
			l.emit(itemCiteKey)
			return lexTagDelim
		case isSpace(r):
			// discard spaces after cite key (to avoid emitting with spaces)
			l.discard()
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexTagDelim scans the tag delimiter, which is known to be present.
func lexTagDelim(l *lexer) stateFn {
	l.emit1(itemTagDelim) // absorb ','
	return lexTagName
}

// lexEntryStopDelim scans the entry stop delimiter, which is known to be present.
func lexEntryStopDelim(l *lexer) stateFn {
	l.emit1(itemEntryStopDelim) // absorb '}'
	// start over, searching for the next bib entry
	return lex
}

// lexTagName scans the tag name, which can be any non-spaced string.
func lexTagName(l *lexer) stateFn {
	// ignore spaces before tag name
	l.ignoreSpaces()
	for {
		if strings.HasPrefix(l.input[l.pos:], "}") {
			return lexEntryStopDelim
		}
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb tag name; we will back up and emit below
		case r == '=':
			l.backup()
			l.emit(itemTagName)
			return lexTagNameContentDelim
		case isSpace(r):
			// discard spaces after tag name (to avoid emitting with spaces)
			l.discard()
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexTagNameContentDelim scans the name-content delimiter, which is known to be present.
func lexTagNameContentDelim(l *lexer) stateFn {
	l.emit1(itemTagNameContentDelim) // absorb '='
	return lexContentStartDelim
}

// lexContentStartDelim scans the name-content start delimiter.
func lexContentStartDelim(l *lexer) stateFn {
	l.ignoreSpaces()
	for {
		switch r := l.next(); {
		// case isSpace(r):
		// 	l.ignore()
		case r == '{':
			l.emit(itemContentStartDelim)
			return lexTagContent
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexTagContent scans the elements inside the content.
func lexTagContent(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r) || isSpace(r):
			// absorb
		case r == '}':
			l.backup()
			l.emit(itemTagContent)
			return lexContentStopDelim
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexContentStopDelim scans the name-content stop delimiter, which is known to be present.
func lexContentStopDelim(l *lexer) stateFn {
	l.emit1(itemContentStopDelim) // absorb '}'
	return lexTagDone
}

//TODO: this should be reusing lexTagDelim

// lexTagDone scans the elements inside the main bib entry.
func lexTagDone(l *lexer) stateFn {
	l.ignoreSpaces()
	for {
		switch r := l.next(); {
		case r == ',':
			l.emit(itemTagDelim)
			return lexTagName
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}
