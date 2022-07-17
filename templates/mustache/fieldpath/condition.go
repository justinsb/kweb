package fieldpath

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"k8s.io/klog/v2"
)

type Condition interface {
	fmt.Stringer
	EvalCondition(m interface{}) bool
}

type BinaryCondition struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (e *BinaryCondition) EvalCondition(o interface{}) bool {
	lv, ok := e.Left.Eval(o)
	if !ok {
		return false
	}
	rv, ok := e.Right.Eval(o)
	if !ok {
		return false
	}

	if e.Operator == "==" {
		equal := reflect.DeepEqual(lv, rv)
		return equal
	}
	if e.Operator == "!=" {
		equal := reflect.DeepEqual(lv, rv)
		return !equal
	}
	klog.Fatalf("unhandled operator %q", e.Operator)
	return false
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

func (e *TruthyCondition) EvalCondition(o interface{}) bool {
	v, ok := e.Expr.Eval(o)
	if !ok {
		return false
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
		klog.Warningf("unhandled type in TruthyCondition %v: %T", e, v)
		return false
	}

	return truthy
}

func (e *TruthyCondition) String() string {
	var s string
	s += e.Expr.String()
	return s
}

type NegateCondition struct {
	Inner Condition
}

func (e *NegateCondition) EvalCondition(o interface{}) bool {
	inner := e.Inner.EvalCondition(o)
	return !inner
}

func (e *NegateCondition) String() string {
	var s string
	s += "!"
	s += e.Inner.String()
	return s
}
