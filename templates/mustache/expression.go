package mustache

import (
	"context"
	"fmt"
	"strings"

	"github.com/justinsb/kweb/templates/mustache/fieldpath"
	"github.com/justinsb/kweb/templates/scopes"
	"k8s.io/klog/v2"
)

type ExpressionList struct {
	Expressions []Expression
}

func (l *ExpressionList) Eval(ctx context.Context, scope *scopes.Scope) (string, error) {
	var values []string
	for _, e := range l.Expressions {
		v, err := e.Eval(ctx, scope)
		if err != nil {
			return "", err
		}
		values = append(values, v)
	}
	return strings.Join(values, ""), nil
}

func (l *ExpressionList) DebugString() string {
	var values []string
	for _, e := range l.Expressions {
		values = append(values, e.DebugString())
	}
	return strings.Join(values, ",")
}

type LiteralExpression struct {
	Literal string
}

func (l *LiteralExpression) Eval(ctx context.Context, scope *scopes.Scope) (string, error) {
	return l.Literal, nil
}

func (l *LiteralExpression) DebugString() string {
	return fmt.Sprintf("%q", l.Literal)
}

type MustacheExpression struct {
	Expression string
}

func (l *MustacheExpression) DebugString() string {
	return l.Expression
}

func (l *MustacheExpression) Eval(ctx context.Context, scope *scopes.Scope) (string, error) {
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

	klog.Infof("parsed expression %q => %q", l.Expression, exprTree.String())
	v, ok := exprTree.Eval(ctx, scope)
	if !ok {
		return "", nil
	}
	return fmt.Sprintf("%v", v), nil
}

type Expression interface {
	Eval(ctx context.Context, scope *scopes.Scope) (string, error)
	DebugString() string
}
