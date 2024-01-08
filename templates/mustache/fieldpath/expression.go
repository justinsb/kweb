package fieldpath

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/justinsb/kweb/templates/scopes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type LiteralExpression struct {
	Value string
}

func (e *LiteralExpression) Eval(ctx context.Context, o interface{}) (interface{}, bool, error) {
	return e.Value, true, nil
}

func (e *LiteralExpression) String() string {
	s := fmt.Sprintf("%q", e.Value)
	return s
}

type IdentifierExpression struct {
	Key string
}

func (e *IdentifierExpression) Eval(ctx context.Context, o interface{}) (interface{}, bool, error) {
	if o == nil {
		return nil, false, nil
	}

	switch o := o.(type) {
	case *scopes.Scope:
		return o.Eval(ctx, e.Key)

	default:
		return nil, false, fmt.Errorf("unhandled type in IdentifierExpression %v: %T", e, o)
	}
}

func (e *IdentifierExpression) String() string {
	s := fmt.Sprintf("%s", e.Key)
	return s
}

type Expression interface {
	fmt.Stringer
	Eval(ctx context.Context, m interface{}) (interface{}, bool, error)
}

type IndexExpression struct {
	Base  Expression
	Key   string
	Style string
}

func (e *IndexExpression) Eval(ctx context.Context, o interface{}) (interface{}, bool, error) {
	if e.Base != nil {
		v, ok, err := e.Base.Eval(ctx, o)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, nil
		}
		o = v
	}

	if o == nil {
		return nil, false, nil
	}

	switch o := o.(type) {
	case *scopes.Scope:
		// TODO: Only at top level?
		return o.Eval(ctx, e.Key)

	case *unstructured.Unstructured:
		v, ok := o.Object[e.Key]
		if !ok {
			return nil, false, nil
		}
		return v, true, nil

	case unstructured.Unstructured:
		v, ok := o.Object[e.Key]
		if !ok {
			return nil, false, nil
		}
		return v, true, nil

	case map[string]interface{}:
		v, ok := o[e.Key]
		if !ok {
			klog.Infof("key %q not found in map %v", e.Key, o)
			return nil, false, nil
		}
		return v, true, nil

	case map[string]string:
		v, ok := o[e.Key]
		if !ok {
			klog.Infof("key %q not found in map %v", e.Key, o)
			return nil, false, nil
		}
		return v, true, nil

	case proto.Message:
		if o == nil {
			return nil, false, nil
		}
		msg := o.ProtoReflect()
		field := msg.Descriptor().Fields().ByName(protoreflect.Name(e.Key))
		if field == nil {
			klog.Warningf("field %q not found in proto message %v", e.Key, msg.Descriptor().FullName())
			return nil, false, nil
		}
		v := msg.Get(field)
		switch field.Kind() {
		case protoreflect.MessageKind:
			return v.Message().Interface(), true, nil
		default:
			return v, true, nil
		}

	default:
		val := reflect.ValueOf(o)

		structVal := val
		if structVal.Kind() == reflect.Ptr {
			structVal = structVal.Elem()
		}
		if structVal.Kind() == reflect.Struct {
			structValType := structVal.Type()
			nFields := structVal.NumField()
			for i := 0; i < nFields; i++ {
				// TODO: Cache json lookup
				field := structValType.Field(i)
				jsonTag := field.Tag.Get("json")
				if jsonTag == "" {
					jsonTag = field.Name
					jsonTag = strings.ToLower(jsonTag[:1]) + jsonTag[1:]
				}
				jsonTag = strings.TrimSuffix(jsonTag, ",omitempty")
				if jsonTag != e.Key {
					continue
				}
				fieldVal := structVal.Field(i)
				return fieldVal.Interface(), true, nil
			}
			klog.Warningf("unhandled type in IndexExpression %v: %T (field %q not known)", e, o, e.Key)
		}

		if val.Kind() == reflect.Ptr {
			valType := val.Type()

			nMethods := valType.NumMethod()
			for i := 0; i < nMethods; i++ {
				// TODO: Cache method lookup
				method := valType.Method(i)
				key := method.Name
				key = strings.ToLower(key[:1]) + key[1:]
				klog.Infof("method is %q", key)
				if key != e.Key {
					continue
				}
				klog.Infof("invoking method %v", method)
				args := []reflect.Value{reflect.ValueOf(ctx)}
				result := val.Method(i).Call(args)
				klog.Infof("result of method call is %v", result)
				if !result[1].IsNil() {
					err := result[1].Interface().(error)
					return result[0].Interface(), true, err
				}
				return result[0].Interface(), true, nil
			}
			klog.Warningf("unhandled type in IndexExpression %v: %T (method %q not known)", e, o, e.Key)

		}

		klog.Warningf("unhandled type in IndexExpression %v: %T (unexpected reflect kind %v)", e, o, val.Kind())
		return nil, false, nil
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
