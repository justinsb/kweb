package login

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/login/pb"
	"github.com/justinsb/kweb/components/users"
	"k8s.io/klog/v2"
)

const CookieNameJWT = "auth-token"

// const stateCookieName = "_oauth2_state"
const sessionOauth2State = "_oauth2_state"

func randomID(bytes int) string {
	b := make([]byte, bytes)
	if _, err := cryptorand.Read(b); err != nil {
		klog.Fatalf("building random id: %v", err)
	}
	sessionID := base64.RawURLEncoding.EncodeToString(b)
	return sessionID
}

func (p *Component) StartOAuth2Login(ctx context.Context, req *components.Request, provider components.AuthenticationProvider) (components.Response, error) {
	providerID := provider.ProviderID()

	err := req.ParseForm()
	if err != nil {
		return components.ErrorResponse(http.StatusBadRequest), err
	}

	redirect := req.FormValue("redirect")
	if redirect == "" {
		redirect = "/"
	}

	state := &pb.StateData{}
	state.ProviderId = providerID
	state.Redirect = redirect
	state.Nonce = randomID(32)

	stateString := encodeState(state)

	req.Session.Set(sessionOauth2State, state)

	redirectURI := p.getRedirectURI(req, providerID)

	return components.RedirectResponse(provider.GetLoginURL(ctx, redirectURI, stateString)), nil
}

func (p *Component) Logout(ctx context.Context, req *components.Request) (components.Response, error) {
	users.Logout(ctx)

	return components.RedirectResponse("/"), nil
}

func (p *Component) getRedirectURI(req *components.Request, providerID string) string {
	var u url.URL
	u.Scheme = req.URL.Scheme
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	u.Host = req.Host

	if req.BrowserUsingHTTPS() {
		u.Scheme = "https"
	}

	u.Path = "/_login/oauth2-callback/" + providerID
	return u.String()
}

func (p *Component) OAuthCallback(ctx context.Context, req *components.Request) (components.Response, error) {
	// finish the oauth cycle
	err := req.ParseForm()
	if err != nil {
		return nil, err
	}

	stateObj := req.Session.Get(sessionOauth2State)
	var state *pb.StateData
	if stateObj != nil {
		state = stateObj.(*pb.StateData)
	}
	stateString := ""
	if state != nil {
		stateString = encodeState(state)
	}

	stateParameter := req.URL.Query().Get("state")
	if stateParameter != stateString {
		klog.Warningf("state in session does not match state in request")
		return nil, fmt.Errorf("state mismatch got=%q vs want=%q", stateParameter, state)
	}

	req.Session.Clear(sessionOauth2State)

	errorString := req.Form.Get("error")
	if errorString != "" {
		return components.ErrorResponse(http.StatusForbidden), fmt.Errorf("permission denied: %v", errorString)
	}

	redirect := state.Redirect
	if !strings.HasPrefix(redirect, "/") {
		redirect = "/"
	}

	code := req.Form.Get("code")
	if code == "" {
		return components.ErrorResponse(http.StatusBadRequest), errors.New("missing code")
	}

	redirectURI := p.getRedirectURI(req, state.ProviderId)
	provider := p.providers[state.ProviderId]
	if provider == nil {
		return nil, fmt.Errorf("unknown provider %q", state.ProviderId)
	}

	if err := provider.Redeem(ctx, redirectURI, code); err != nil {
		return nil, err
	}

	return components.RedirectResponse(redirect), nil
}
