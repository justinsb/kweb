package fieldpath

import (
	"fmt"
)

func ParseCondition(s string) (Condition, error) {
	p := Parser{}
	p.Init(s)

	expr, err := p.ParseCondition()
	if err != nil {
		return nil, fmt.Errorf("error during parsing of %q: %w", s, err)
	}
	if err := p.Complete(); err != nil {
		return nil, fmt.Errorf("error parsing condition %q: %w", s, err)
	}
	return expr, nil
}

func ParseExpression(s string) (Expression, error) {
	p := Parser{}
	p.Init(s)

	expr, err := p.ParseExpression()
	if err != nil {
		return nil, fmt.Errorf("error during parsing of %q: %w", s, err)
	}
	if err := p.Complete(); err != nil {
		return nil, fmt.Errorf("error parsing expression %q: %w", s, err)
	}
	return expr, nil
}
