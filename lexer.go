package biblexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name  string    // the name of the input; used only for error reports.
	input string    // the string being scanned.
	state stateFn   // the next lexing function to enter
	pos   int       // current position in the input.
	start int       // start position of this item.
	skip  int       // number of rune's to skip (usually spaces)
	width int       // width of last rune read from input.
	items chan item // channel of scanned items.
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// discard skips the current rune that may appear after an item.
// Typically this will be to discard spaces after an item.
func (l *lexer) discard() {
	_, w := utf8.DecodeRuneInString(l.input[l.pos:])
	// add the width of the current rune to the skip count
	l.skip += w
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// ignoreSpaces skips over the remaining seqeunce of spaces.
func (l *lexer) ignoreSpaces() {
	for {
		switch r := l.next(); {
		case isSpace(r):
			l.ignore()
		default:
			l.backup()
			return
		}
	}
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	// backup pos if there are runes to skip
	pos := l.pos - l.skip
	l.items <- item{t, l.input[l.start:pos]}
	l.start = l.pos
	// reset the skip counter
	l.skip = 0
}

// emit passes an item back to the client.
func (l *lexer) emit1(t itemType) {
	l.pos++
	l.emit(t)
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.pos], "\n")
}

// error returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '.' || r == '/' || r == '\\' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isUnbrokenAlphaNumericToken reports whether r is part of an unbroken
// sequence of alphanumeric characters. Any call to discard prior to calling
// isUnbrokenAlphaNumericToken will break the sequence; it will return false.
func (l *lexer) isUnbrokenAlphaNumericToken(r rune) bool {
	return isAlphaNumeric(r) && l.skip == 0
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			if l.state == nil {
				return item{itemEOF, ""}
			}
			l.state = l.state(l)
		}
	}
}

// NewLexer creates a new scanner for the input string.
func newLexer(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		state: lexStart,
		items: make(chan item, 2), // Two items sufficient.
	}
	return l
}
