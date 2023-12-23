package sessions

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/cookies"
	"github.com/justinsb/kweb/debug"
	"github.com/justinsb/kweb/templates/scopes"
	"k8s.io/klog/v2"
)

const cookieSessionID = "session"

// SessionComponent is the component that implements Sessions
type SessionComponent struct {
	storage Storage
}

func NewSessionComponent(storage Storage) *SessionComponent {
	return &SessionComponent{
		storage: storage,
	}
}

func (c *SessionComponent) beforeRequest(ctx context.Context, req *components.Request) (*Session, error) {
	sessionID := ""
	cookie, err := req.Cookie(cookieSessionID)
	if err != nil && err != http.ErrNoCookie {
		return nil, fmt.Errorf("error reading cookie: %w", err)
	}
	if cookie != nil {
		sessionID = cookie.Value
	}

	var session *Session
	if sessionID != "" {
		s, err := c.storage.LookupSession(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		session = s
		if session != nil {
			klog.Infof("using session %q: %#v", sessionID, session)
			return session, nil
		}
		klog.Infof("session %q was in cookie but not found", sessionID)
		sessionID = ""
	}

	session = &Session{
		newSession: true,
		component:  c,
	}
	return session, nil
}

var _ components.RequestFilter = &SessionComponent{}

func (c *SessionComponent) ProcessRequest(ctx context.Context, req *components.Request, next components.RequestFilterChain) (components.Response, error) {
	session, err := c.beforeRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	klog.Infof("SessionComponent::ProcessRequest")
	req.Session = session

	response, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	if session.dirty {
		err := c.storage.WriteSession(ctx, session)
		if err != nil {
			return nil, err
		}

		if session.newSession {
			sessionCookie := http.Cookie{
				Name:     cookieSessionID,
				Value:    session.ID,
				Expires:  time.Now().Add(time.Hour * 24 * 365),
				HttpOnly: true,
				Secure:   true,
				Path:     "/", // Otherwise cookie is filtered
			}

			if !req.BrowserUsingHTTPS() {
				if req.IsLocalhost() {
					klog.Warningf("setting cookie to _not_ be secure, because running on localhost")
					sessionCookie.Secure = false
				} else {
					klog.Warningf("session invoked but running without TLS (and not on localhost); likely won't work")
				}
			}

			cookies.SetCookie(ctx, sessionCookie)

			session.newSession = false
		}
		klog.Infof("session %q => %v", session.ID, debug.JSON(session.values))
		session.dirty = false
	}

	return response, nil
}

func (c *SessionComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	return nil
}

func (c *SessionComponent) AddToScope(ctx context.Context, scope *scopes.Scope) {
}
