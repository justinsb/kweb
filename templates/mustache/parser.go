package mustache

import "github.com/justinsb/kweb/templates/lexparse"

type Parser struct {
	lexparse.BaseParser
}

func (p *Parser) Init(s string) {
	l := &lexer{}
	l.Init(s)
	p.BaseParser.Init(l)
}

func (p *Parser) ParseMustacheExpression() (*MustacheExpression, error) {
	switch p.PeekTokenType() {
	case tokenTypeLeftMustache:
		p.Expect(tokenTypeLeftMustache)
		t := p.Expect(tokenTypeOther)
		p.Expect(tokenTypeRightMustache)
		return &MustacheExpression{Expression: t.Value}, nil
	default:
		return nil, p.Unexpected()
	}
}

func (p *Parser) ParseLiteralExpression() (*LiteralExpression, error) {
	switch p.PeekTokenType() {
	case tokenTypeOther:
		t := p.Expect(tokenTypeOther)
		return &LiteralExpression{Literal: t.Value}, nil

	default:
		return nil, p.Unexpected()
	}
}

func (p *Parser) ParseExpression() (Expression, error) {
	switch p.PeekTokenType() {
	case tokenTypeOther:
		return p.ParseLiteralExpression()
	case tokenTypeLeftMustache:
		return p.ParseMustacheExpression()

	default:
		return nil, p.Unexpected()
	}
}

func (p *Parser) ParseExpressionList() (*ExpressionList, error) {
	el := &ExpressionList{}

	first, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	el.Expressions = append(el.Expressions, first)

	for {
		switch p.PeekTokenType() {
		case tokenTypeLeftMustache, tokenTypeOther:
			e, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			el.Expressions = append(el.Expressions, e)

		case tokenTypeEOF:
			return el, nil

		default:
			return nil, p.Unexpected()
		}
	}
}
