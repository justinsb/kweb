package mustache

import "github.com/justinsb/kweb/templates/lexparse"

type lexer struct {
	lexparse.BaseLexer
}

func (l *lexer) Init(s string) {
	l.BaseLexer.Init(s)
}

type token = lexparse.Token

const (
	tokenTypeLeftMustache  lexparse.TokenType = '{'
	tokenTypeRightMustache                    = '}'
	tokenTypeOther                            = '.'
	tokenTypeEOF                              = lexparse.TokenTypeEOF
	tokenTypeError                            = lexparse.TokenTypeError
)

func (l *lexer) lexOther(first rune) (token, error) {
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

		case '{', '}':
			l.Unread(r)
			break runeLoop
		default:
			s = append(s, r)
		}
	}
	return token{TokenType: tokenTypeOther, Value: string(s)}, nil
}

func (l *lexer) Next() (token, error) {
	r := l.Read()

	switch r {
	case lexparse.LexerRuneEOF:
		return token{TokenType: tokenTypeEOF, Value: ""}, nil

	case lexparse.LexerRuneError:
		return token{}, l.Err()

	case '{':
		r2 := l.Peek()
		if r2 == '{' {
			l.Read()
			return token{TokenType: tokenTypeLeftMustache, Value: "{{"}, nil
		}

	case '}':
		r2 := l.Peek()
		if r2 == '}' {
			l.Read()
			return token{TokenType: tokenTypeRightMustache, Value: "}}"}, nil
		}
	}

	return l.lexOther(r)
}
