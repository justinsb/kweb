package sessions

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
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
	mutex    sync.Mutex
	sessions map[string]*Session
}

func NewSessionComponent() *SessionComponent {
	return &SessionComponent{
		sessions: make(map[string]*Session),
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
		c.mutex.Lock()
		session = c.sessions[sessionID]
		c.mutex.Unlock()

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
		if session.newSession {
			sessionID, err := c.generateSessionID()
			if err != nil {
				return nil, err
			}
			sessionCookie := http.Cookie{
				Name:     cookieSessionID,
				Value:    sessionID,
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

			klog.Infof("storing session %q", sessionID)
			c.mutex.Lock()
			c.sessions[sessionID] = session
			c.mutex.Unlock()
			session.newSession = false
			session.ID = sessionID
		}
		klog.Infof("session %q => %v", session.ID, debug.JSON(session.values))
		session.dirty = false
	}

	return response, nil
}

func (c *SessionComponent) generateSessionID() (string, error) {
	b := make([]byte, 32, 32)
	if _, err := cryptorand.Read(b); err != nil {
		return "", fmt.Errorf("error building session id: %w", err)
	}
	sessionID := base64.RawURLEncoding.EncodeToString(b)
	return sessionID, nil
}

func (c *SessionComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	return nil
}

func (c *SessionComponent) AddToScope(ctx context.Context, scope *scopes.Scope) {
}
