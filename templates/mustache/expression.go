package mustache

import (
	"fmt"
	"strings"

	"github.com/justinsb/kweb/templates/mustache/fieldpath"
	"github.com/justinsb/kweb/templates/scopes"
)

type ExpressionList struct {
	Expressions []Expression
}

func (l *ExpressionList) Eval(scope *scopes.Scope) (string, error) {
	var values []string
	for _, e := range l.Expressions {
		v, err := e.Eval(scope)
		if err != nil {
			return "", err
		}
		values = append(values, v)
	}
	return strings.Join(values, ""), nil
}

type LiteralExpression struct {
	Literal string
}

func (l *LiteralExpression) Eval(scope *scopes.Scope) (string, error) {
	return l.Literal, nil
}

type MustacheExpression struct {
	Expression string
}

func (l *MustacheExpression) Eval(scope *scopes.Scope) (string, error) {
	// TODO: Pre-parse
	p := fieldpath.Parser{}
	p.Init(l.Expression)

	exprTree, err := p.ParseExpression()
	if err != nil {
		return "", fmt.Errorf("error during parsing of %q: %w", l.Expression, err)
	}
	if err := p.Complete(); err != nil {
		return "", fmt.Errorf("error parsing expression %q: %w", l.Expression, err)
	}

	v, ok := exprTree.Eval(scope)
	if !ok {
		return "", nil
	}
	return fmt.Sprintf("%v", v), nil
}

type Expression interface {
	Eval(scope *scopes.Scope) (string, error)
}
