package sessions

import (
	"sync"

	"google.golang.org/protobuf/proto"
)

type Session struct {
	ID string

	newSession bool
	dirty      bool
	component  *SessionComponent

	// TODO: Need to consider concurrent requests.  Should we lock the session?  A read-write lock?  Snapshot semantics?
	mutex  sync.Mutex
	values map[string]*sessionValue
}

type sessionValue struct {
	Message   proto.Message
	StringVal string
}

func (s *Session) Clear(key string) {
	_, found := s.values[key]
	if !found {
		return
	}
	delete(s.values, key)
	s.dirty = true
}

func (s *Session) Set(key string, msg proto.Message) {
	if s.values == nil {
		s.values = make(map[string]*sessionValue)
	}
	s.values[key] = &sessionValue{
		Message: msg,
	}
	s.dirty = true
}

func (s *Session) SetString(key string, val string) {
	if s.values == nil {
		s.values = make(map[string]*sessionValue)
	}
	s.values[key] = &sessionValue{
		StringVal: val,
	}
	s.dirty = true
}

func (s *Session) Get(key string) proto.Message {
	v, found := s.values[key]
	if !found {
		return nil
	}
	return v.Message
}

func (s *Session) GetString(key string) string {
	v, found := s.values[key]
	if !found {
		return ""
	}
	return v.StringVal
}
