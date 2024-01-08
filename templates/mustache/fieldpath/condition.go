package fieldpath

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
)

type Condition interface {
	fmt.Stringer
	EvalCondition(ctx context.Context, m interface{}) (bool, error)
}

type BinaryCondition struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (e *BinaryCondition) EvalCondition(ctx context.Context, o interface{}) (bool, error) {
	lv, ok, err := e.Left.Eval(ctx, o)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	rv, ok, err := e.Right.Eval(ctx, o)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	if e.Operator == "==" {
		equal := reflect.DeepEqual(lv, rv)
		return equal, nil
	}
	if e.Operator == "!=" {
		equal := reflect.DeepEqual(lv, rv)
		return !equal, nil
	}
	return false, fmt.Errorf("unhandled operator %q", e.Operator)
}

func (e *BinaryCondition) String() string {
	var s string
	s += e.Left.String()
	s += e.Operator
	s += e.Right.String()

	return s
}

type TruthyCondition struct {
	Expr Expression
}

func (e *TruthyCondition) EvalCondition(ctx context.Context, o interface{}) (bool, error) {
	v, ok, err := e.Expr.Eval(ctx, o)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	var truthy bool

	switch v := v.(type) {
	case string:
		truthy = v != ""

	case proto.Message:
		msg := v.ProtoReflect()
		truthy = msg.IsValid()
	// case runtime.Object:
	// 	truthy = v != nil
	default:
		return false, fmt.Errorf("unhandled type in TruthyCondition %v: %T", e, v)
	}

	return truthy, nil
}

func (e *TruthyCondition) String() string {
	var s string
	s += e.Expr.String()
	return s
}

type NegateCondition struct {
	Inner Condition
}

func (e *NegateCondition) EvalCondition(ctx context.Context, o interface{}) (bool, error) {
	inner, err := e.Inner.EvalCondition(ctx, o)
	if err != nil {
		return false, err
	}
	return !inner, nil
}

func (e *NegateCondition) String() string {
	var s string
	s += "!"
	s += e.Inner.String()
	return s
}
