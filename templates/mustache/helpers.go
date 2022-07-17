package mustache

import (
	"fmt"
)

func ParseExpressionList(s string) (*ExpressionList, error) {
	p := Parser{}
	p.Init(s)

	mustacheExpr, err := p.ParseExpressionList()
	if err != nil {
		return nil, fmt.Errorf("error during parsing of %q: %w", s, err)
	}
	if err := p.Complete(); err != nil {
		return nil, fmt.Errorf("error parsing expression list %q: %w", s, err)
	}
	return mustacheExpr, nil
}
