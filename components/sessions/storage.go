package sessions

import "context"

type Storage interface {
	LookupSession(ctx context.Context, sessionID string) (*Session, error)
	WriteSession(ctx context.Context, session *Session) error
}
