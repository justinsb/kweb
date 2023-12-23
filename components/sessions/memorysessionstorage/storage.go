package memorysessionstorage

import (
	"context"
	"encoding/base64"
	"sync"

	cryptorand "crypto/rand"

	"github.com/justinsb/kweb/components/sessions"
	"k8s.io/klog/v2"
)

type MemorySessionStorage struct {
	mutex    sync.Mutex
	sessions map[string][]byte
}

var _ sessions.Storage = &MemorySessionStorage{}

func NewMemorySessionStorage() *MemorySessionStorage {
	return &MemorySessionStorage{
		sessions: make(map[string][]byte),
	}
}

func (s *MemorySessionStorage) LookupSession(ctx context.Context, sessionID string) (*sessions.Session, error) {
	if sessionID == "" {
		return nil, nil
	}

	s.mutex.Lock()
	data, found := s.sessions[sessionID]
	s.mutex.Unlock()

	if !found {
		return nil, nil
	}

	return sessions.Decode(sessionID, data)
}

func (s *MemorySessionStorage) WriteSession(ctx context.Context, session *sessions.Session) error {
	sessionID := session.ID
	if sessionID == "" {
		// A new session
		sessionID = GenerateSessionID()
		session.ID = sessionID
	}

	b, err := sessions.Encode(session)
	if err != nil {
		return err
	}

	klog.Infof("storing session %q", sessionID)
	s.mutex.Lock()
	s.sessions[sessionID] = b
	s.mutex.Unlock()

	return nil
}

func GenerateSessionID() string {
	b := make([]byte, 32, 32)
	if _, err := cryptorand.Read(b); err != nil {
		klog.Fatalf("error building session id: %v", err)
	}
	sessionID := base64.RawURLEncoding.EncodeToString(b)
	return sessionID
}
