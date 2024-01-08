package scopes

import "context"

type Scope struct {
	Values map[string]Value
}

type Value struct {
	Value    any
	Function func() (any, error)
}

func NewScope() *Scope {
	return &Scope{Values: make(map[string]Value)}
}

func (s *Scope) Eval(ctx context.Context, name string) (interface{}, bool, error) {
	v, ok := s.Values[name]
	if !ok {
		return nil, false, nil
	}
	if v.Function != nil {
		fnVal, err := v.Function()
		return fnVal, true, err
	} else {
		return v.Value, true, nil
	}
}
