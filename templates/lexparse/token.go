package lexparse

type TokenType int

type Token struct {
	TokenType TokenType
	Value     string
}

const (
	TokenTypeEOF   TokenType = -1
	TokenTypeError           = -2
)
