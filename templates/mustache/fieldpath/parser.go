package fieldpath

import (
	"fmt"

	"github.com/justinsb/kweb/templates/lexparse"
)

type Parser struct {
	lexparse.BaseParser
}

func (p *Parser) Init(s string) {
	l := &lexer{}
	l.Init(s)
	p.BaseParser.Init(l)
}

func (p *Parser) ParseExpression() (Expression, error) {
	var e Expression

	switch p.PeekTokenType() {
	case tokenTypeIdentifier:
		t := p.Expect(tokenTypeIdentifier)
		e = &IdentifierExpression{
			Key: t.Value,
		}

	case tokenTypeQuotedString:
		t := p.Expect(tokenTypeQuotedString)
		e = &LiteralExpression{
			Value: t.Value,
		}

	default:
		return nil, fmt.Errorf("expected identifier; got %v", p.PeekTokenType())
	}

	for {
		switch p.PeekTokenType() {
		case tokenTypeDot:
			p.Expect(tokenTypeDot)
			id := p.Expect(tokenTypeIdentifier)

			ie := &IndexExpression{Key: id.Value, Style: "."}
			ie.Base = e
			e = ie

		case tokenTypeLeftSquareBracket:
			p.Expect(tokenTypeLeftSquareBracket)
			id := p.Expect(tokenTypeIdentifier)
			p.Expect(tokenTypeRightSquareBracket)

			ie := &IndexExpression{Key: id.Value, Style: "["}
			ie.Base = e
			e = ie

		case tokenTypeEOF:
			p.Expect(tokenTypeEOF)
			return e, nil

		case tokenTypeError:
			return nil, nil

		default:
			return e, nil
		}
	}
}

func (p *Parser) ParseCondition() (Condition, error) {
	var left Expression
	switch p.PeekTokenType() {
	case tokenTypeNot:
		p.Expect(tokenTypeNot)
		inner, err := p.ParseCondition()
		if err != nil {
			return nil, err
		}
		return &NegateCondition{
			Inner: inner,
		}, nil
	case tokenTypeIdentifier:
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		left = expr

	default:
		return nil, fmt.Errorf("unexpected token to start condition: %v", p.PeekTokenType())
	}

	switch p.PeekTokenType() {
	case tokenTypeEOF:
		// A value expression
		return &TruthyCondition{
			Expr: left,
		}, nil
	case tokenTypeEquals:
		p.Expect(tokenTypeEquals)
		p.Expect(tokenTypeEquals)

		right, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}

		return &BinaryCondition{
			Left:     left,
			Right:    right,
			Operator: "==",
		}, nil
	case tokenTypeNot:
		p.Expect(tokenTypeNot)
		p.Expect(tokenTypeEquals)

		right, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}

		return &BinaryCondition{
			Left:     left,
			Right:    right,
			Operator: "!=",
		}, nil

	default:
		return nil, p.Unexpected()
	}
}
