package sessions

import (
	"sync"

	"google.golang.org/protobuf/proto"
	"k8s.io/klog/v2"
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
	Data []byte
}

func (s *Session) Clear(v proto.Message) {
	key := string(v.ProtoReflect().Descriptor().FullName())
	_, found := s.values[key]
	if !found {
		return
	}
	delete(s.values, key)
	s.dirty = true
}

func (s *Session) Set(msg proto.Message) {
	key := string(msg.ProtoReflect().Descriptor().FullName())
	if s.values == nil {
		s.values = make(map[string]*sessionValue)
	}
	b, err := proto.Marshal(msg)
	if err != nil {
		klog.Fatalf("encoding message %v", err)
	}
	s.values[key] = &sessionValue{
		Data: b,
	}
	s.dirty = true
}

func (s *Session) Get(dest proto.Message) bool {
	key := string(dest.ProtoReflect().Descriptor().FullName())
	v, found := s.values[key]
	if !found {
		return false
	}
	if err := proto.Unmarshal(v.Data, dest); err != nil {
		klog.Warningf("error parsing session data %q: %v", key, err)
		return false
	}
	return true
}
