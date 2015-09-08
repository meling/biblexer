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
		// case i.typ > itemKeyword:
		// 	return fmt.Sprintf("<%s>", i.val)
		// case len(i.val) > 10:
		// 	return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemComment          // comment (%)
	itemEntryType        // entry type
	itemEntryTypeDelim   // entry type delimiter (@)
	itemCiteKey          // cite key
	itemTagDelim         // tag delimiter (,)
	itemTagKeyValueDelim // the delimiter separating key and value (=)
	itemLeftEntryDelim   // left entry delimiter ({)
	itemRightEntryDelim  // right entry delimiter (})
	itemValueLeftDelim   // value left delimiter ({)
	itemValueRightDelim  // value right delimiter (})
	itemConcat           // the concatination symbol (#)
	//TODO: should be itemTagKey and itemTagValue
	itemKey   // the key (on left of =)
	itemValue // quoted string (includes quotes)
)

var key = map[string]itemType{
	"citekey": itemCiteKey,
	"@":       itemEntryTypeDelim,
	"{":       itemLeftEntryDelim,
	"}":       itemRightEntryDelim,
	"#":       itemConcat,
	",":       itemTagDelim,
	"=":       itemTagKeyValueDelim,
}

// state functions

const (
	// TODO: Add support for ignoring comments later
	commentDelim = "%"
)

// lexText scans until an opening action delimiter, "@".
func lexText(l *lexer) stateFn {
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
	l.pos += 1 // absorb '@'
	l.emit(itemEntryTypeDelim)
	return lexEntryType
}

// lexEntryType scans the entry type.
func lexEntryType(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || isSpace(r):
			return l.errorf("unclosed action")
		case isAlphaNumeric(r):
			// absorb.
		case r == '{':
			l.backup()
			l.emit(itemEntryType)
			return lexLeftEntryDelim
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexLeftEntryDelim scans the left entry delimiter, which is known to be present.
func lexLeftEntryDelim(l *lexer) stateFn {
	l.pos += 1 // absorb '{'
	l.emit(itemLeftEntryDelim)
	return lexCiteKey
}

// lexCiteKey scans the cite key.
func lexCiteKey(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || isSpace(r):
			return l.errorf("unclosed action")
		case isAlphaNumeric(r) || r == '/':
			// absorb.
		case r == ',':
			l.backup()
			l.emit(itemCiteKey)
			return lexTagDelim
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexTagDelim scans the tag delimiter, which is known to be present.
func lexTagDelim(l *lexer) stateFn {
	l.pos += 1 // absorb ','
	l.emit(itemTagDelim)
	return lexTagKey
}

// lexRightEntryDelim scans the right entry delimiter, which is known to be present.
func lexRightEntryDelim(l *lexer) stateFn {
	l.pos += 1 // absorb '}'
	l.emit(itemRightEntryDelim)
	return lexText
}

// lexTag scans the tag key, which can be any non-spaced string.
func lexTagKey(l *lexer) stateFn {
	keyNotFound := true
	spaces := 0
	for {
		if strings.HasPrefix(l.input[l.pos:], "}") {
			return lexRightEntryDelim
		}
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb tag key; we will back up and emit below
			keyNotFound = false
		case isSpace(r) && keyNotFound:
			// ignore spaces until tag key is found
			l.ignore()
		case r == '=':
			l.backupN(spaces)
			l.emit(itemKey)
			l.forwardN(spaces)
			return lexTagKeyValueDelim
		case isSpace(r):
			// count spaces on right of tag key
			spaces++
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexTagKeyValueDelim scans the tag key-value delimiter, which is known to be present.
func lexTagKeyValueDelim(l *lexer) stateFn {
	l.pos += 1 // absorb '='
	l.emit(itemTagKeyValueDelim)
	return lexTagValueLeftDelim
}

// lexTagValue scans the elements inside the main bib entry.
func lexTagValueLeftDelim(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isSpace(r):
			l.ignore()
		case r == '{':
			l.emit(itemValueLeftDelim)
			return lexTagValue
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexTagValue scans the elements inside a value.
func lexTagValue(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r) || isSpace(r):
			// absorb
		case r == '}':
			l.backup()
			l.emit(itemValue)
			return lexValueRightDelim
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexKeyValueDelim scans the tag key-value delimiter, which is known to be present.
func lexValueRightDelim(l *lexer) stateFn {
	l.pos += 1 // absorb '}'
	l.emit(itemValueRightDelim)
	return lexTagDone
}

// lexTagDone scans the elements inside the main bib entry.
func lexTagDone(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == ',':
			l.emit(itemTagDelim)
			return lexTagKey
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}

// lexTagValue scans the elements inside the main bib entry.
func lexTagValueOLD(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r) || isSpace(r):
			// absorb
		case r == ',':
			//TODO: This code seems to be dead (have no impact)??
			l.backup()
			l.emit(itemValue)
			l.pos += 1 // absorb ','
			l.emit(itemTagDelim)
			return lexTagKey
		case isSpace(r):
			l.ignore()
		case r == '{':
			l.emit(itemValueLeftDelim)
			// l.ignore()
		case r == '"':
			e := l.peek()
			if e == ',' || e == '}' {
				l.backup()
				l.emit(itemValue)
				l.pos += 2 // absorb '",' or '"}'
				// l.emit(itemValueRightDelim)
				l.ignore()
				return lexTagKey
			}
			l.ignore()
		case r == '}':
			e := l.peek()
			if e == ',' || e == '}' {
				l.backup()
				l.emit(itemValue)
				l.pos += 2 // absorb '},' or '}}'
				l.ignore()
				return lexTagKey
			}
			l.ignore()
		case r == eof:
			return l.errorf("unclosed action")
		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}
}
