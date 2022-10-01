package fieldpath

import (
	"fmt"

	"github.com/justinsb/kweb/templates/lexparse"
)

type lexer struct {
	lexparse.BaseLexer
}

func (l *lexer) Init(s string) {
	l.BaseLexer.Init(s)
}

type token = lexparse.Token

const (
	tokenTypeIdentifier         lexparse.TokenType = 'I'
	tokenTypeQuotedString                          = '"'
	tokenTypeDot                                   = '.'
	tokenTypeLeftSquareBracket                     = '['
	tokenTypeRightSquareBracket                    = ']'
	tokenTypeNot                                   = '!'
	tokenTypeEquals                                = '='
	tokenTypeEOF                                   = lexparse.TokenTypeEOF
	tokenTypeError                                 = lexparse.TokenTypeError
)

func (l *lexer) lexQuotedString(quote rune) (token, error) {
	var s []rune
runeLoop:
	for {
		r := l.Read()

		switch r {
		case lexparse.LexerRuneEOF:
			return token{}, fmt.Errorf("expected closing quote for string")

		case lexparse.LexerRuneError:
			return token{}, l.Err()

		case quote:
			break runeLoop
		}

		s = append(s, r)
	}
	return token{TokenType: tokenTypeQuotedString, Value: string(s)}, nil
}

func (l *lexer) lexIdentifier(first rune) (token, error) {
	var s []rune
	s = append(s, first)
runeLoop:
	for {
		r := l.Read()
		switch r {
		case lexparse.LexerRuneEOF:
			break runeLoop

		case lexparse.LexerRuneError:
			return token{}, l.Err()

		case '_':
			s = append(s, r)

		default:
			if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') {
				s = append(s, r)
			} else {
				l.Unread(r)
				break runeLoop
			}
		}
	}
	return token{TokenType: tokenTypeIdentifier, Value: string(s)}, nil
}

func (l *lexer) Next() (token, error) {
top:
	r := l.Read()

	switch r {
	case lexparse.LexerRuneEOF:
		return token{TokenType: tokenTypeEOF, Value: ""}, nil

	case lexparse.LexerRuneError:
		return token{}, l.Err()

	case ' ':
		goto top

	case '.':
		return token{TokenType: tokenTypeDot, Value: "."}, nil
	case '[':
		return token{TokenType: tokenTypeLeftSquareBracket, Value: "["}, nil
	case ']':
		return token{TokenType: tokenTypeRightSquareBracket, Value: "]"}, nil
	case '=':
		return token{TokenType: tokenTypeEquals, Value: "="}, nil
	case '!':
		return token{TokenType: tokenTypeNot, Value: "!"}, nil
	case '"':
		return l.lexQuotedString(r)
	case '\'':
		return l.lexQuotedString(r)
	default:
		return l.lexIdentifier(r)
	}
}
