package lexparse

import (
	"io"
	"strings"
)

type BaseLexer struct {
	reader      *strings.Reader
	unreadRunes []rune
	err         error
	eof         bool
}

const (
	LexerRuneEOF   rune = -1
	LexerRuneError rune = -2
)

func (l *BaseLexer) Init(s string) {
	l.reader = strings.NewReader(s)
	l.eof = false
	l.err = nil
	l.unreadRunes = l.unreadRunes[:0]
}

func (l *BaseLexer) Unread(r rune) {
	if l.err != nil {
		return
	}
	l.unreadRunes = append(l.unreadRunes, r)
}

func (l *BaseLexer) Read() rune {
	if l.err != nil {
		return LexerRuneError
	}

	if n := len(l.unreadRunes); n != 0 {
		r := l.unreadRunes[n-1]
		l.unreadRunes = l.unreadRunes[:n-1]
		return r
	}

	if l.eof {
		return LexerRuneEOF
	}

	r, _, err := l.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			l.eof = true
			return LexerRuneEOF
		} else {
			l.err = err
			return LexerRuneError
		}
	}
	return r
}

func (l *BaseLexer) Peek() rune {
	r := l.Read()
	l.Unread(r)
	return r
}

func (l *BaseLexer) Err() error {
	if l.err == nil && l.eof {
		return io.EOF
	}
	return l.err
}
