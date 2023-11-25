package fieldpath

import (
	"fmt"

	"github.com/justinsb/kweb/templates/scopes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type LiteralExpression struct {
	Value string
}

func (e *LiteralExpression) Eval(o interface{}) (interface{}, bool) {
	return e.Value, true
}

func (e *LiteralExpression) String() string {
	s := fmt.Sprintf("%q", e.Value)
	return s
}

type IdentifierExpression struct {
	Key string
}

func (e *IdentifierExpression) Eval(o interface{}) (interface{}, bool) {
	if o == nil {
		return nil, false
	}

	switch o := o.(type) {
	case *scopes.Scope:
		return o.Eval(e.Key)

	default:
		klog.Warningf("unhandled type in IdentifierExpression %v: %T", e, o)
		return nil, false
	}
}

func (e *IdentifierExpression) String() string {
	s := fmt.Sprintf("%s", e.Key)
	return s
}

type Expression interface {
	fmt.Stringer
	Eval(m interface{}) (interface{}, bool)
}

type IndexExpression struct {
	Base  Expression
	Key   string
	Style string
}

func (e *IndexExpression) Eval(o interface{}) (interface{}, bool) {
	if e.Base != nil {
		v, ok := e.Base.Eval(o)
		if !ok {
			return nil, false
		}
		o = v
	}

	if o == nil {
		return nil, false
	}

	switch o := o.(type) {
	case *scopes.Scope:
		// TODO: Only at top level?
		return o.Eval(e.Key)

	case *unstructured.Unstructured:
		v, ok := o.Object[e.Key]
		if !ok {
			return nil, false
		}
		return v, true

	case unstructured.Unstructured:
		v, ok := o.Object[e.Key]
		if !ok {
			return nil, false
		}
		return v, true

	case map[string]interface{}:
		v, ok := o[e.Key]
		if !ok {
			klog.Infof("key %q not found in map %v", e.Key, o)
			return nil, false
		}
		return v, true

	case proto.Message:
		if o == nil {
			return nil, false
		}
		msg := o.ProtoReflect()
		field := msg.Descriptor().Fields().ByName(protoreflect.Name(e.Key))
		if field == nil {
			klog.Warningf("field %q not found in proto message %v", e.Key, msg.Descriptor().FullName())
			return nil, false
		}
		v := msg.Get(field)
		switch field.Kind() {
		case protoreflect.MessageKind:
			return v.Message().Interface(), true
		default:
			return v, true
		}

	default:
		klog.Warningf("unhandled type in IndexExpression %v: %T", e, o)
		return nil, false
	}
}

func (e *IndexExpression) String() string {
	var s string
	if e.Base != nil {
		s = e.Base.String()
	}
	if e.Style == "." {
		s += "."
		s += e.Key
	} else {
		s += "["
		s += e.Key
		s += "]"
	}

	return s
}
