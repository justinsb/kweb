package scopes

import "context"

type Scope struct {
	Values map[string]Value
}

type Value struct {
	Value    interface{}
	Function func() interface{}
}

func NewScope() *Scope {
	return &Scope{Values: make(map[string]Value)}
}

func (s *Scope) Eval(ctx context.Context, name string) (interface{}, bool) {
	v, ok := s.Values[name]
	if !ok {
		return nil, false
	}
	if v.Function != nil {
		fnVal := v.Function()
		return fnVal, true
	} else {
		return v.Value, true
	}
}
