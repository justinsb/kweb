package login

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/login/pb"
	"github.com/justinsb/kweb/components/users"
	userapi "github.com/justinsb/kweb/components/users/pb"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
)

// const stateCookieName = "_oauth2_state"
const sessionOauth2State = "_oauth2_state"

type UserMapper interface {
	MapToUser(ctx context.Context, req *components.Request, token *oauth2.Token, info *components.AuthenticationInfo) (*userapi.User, error)
}

func (p *Component) OAuthStart(ctx context.Context, req *components.Request) (components.Response, error) {
	err := req.ParseForm()
	if err != nil {
		return components.ErrorResponse(http.StatusBadRequest), err
	}

	redirect := req.FormValue("redirect")
	if redirect == "" {
		redirect = "/"
	}

	state := &pb.StateData{}
	// state.ProviderId = providerID
	state.Redirect = redirect
	state.Nonce = strconv.FormatInt(rand.Int63(), 16)

	stateString := encodeState(state)

	req.Session.Set(sessionOauth2State, state)

	redirectURI := p.getRedirectURI(req)

	return components.RedirectResponse(p.Provider.GetLoginURL(ctx, redirectURI, stateString)), nil
}

func (p *Component) Logout(ctx context.Context, req *components.Request) (components.Response, error) {
	users.Logout(ctx)

	return components.RedirectResponse("/"), nil
}

func (p *Component) getRedirectURI(req *components.Request) string {
	var u url.URL
	u.Scheme = req.URL.Scheme
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	u.Host = req.Host

	u.Path = "/_login/oauth2-callback/" + p.Provider.ProviderID()
	return u.String()
}

func (p *Component) OAuthCallback(ctx context.Context, req *components.Request) (components.Response, error) {
	// finish the oauth cycle
	err := req.ParseForm()
	if err != nil {
		return nil, err
	}

	klog.Infof("request is %#v", req)
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
		klog.Warningf("state in cookie does not match state in request")
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

	redirectURI := p.getRedirectURI(req)
	info, token, err := p.Provider.Redeem(ctx, redirectURI, code)
	if err != nil {
		return nil, fmt.Errorf("error redeeming code: %w", err)
	}

	// set cookie, or deny
	userInfo, err := p.UserMapper.MapToUser(ctx, req, token, info)
	if err != nil {
		klog.Infof("error mapping to user: %v", err)
		return components.ErrorResponse(http.StatusInternalServerError), err
	}

	klog.Infof("authentication complete %v", info)

	users.SetCurrentUser(ctx, token, info.Provider.ProviderID(), userInfo)

	// ctx = user.WithUser(ctx, userInfo)

	return components.RedirectResponse(redirect), nil
}

// DebugInfo is a simple endpoint for debugging, while we can't do much more
// TODO: Remove me!
func (p *Component) DebugInfo(ctx context.Context, req *components.Request) (components.Response, error) {
	user := users.GetUser(ctx)
	var html string
	if user == nil {
		html = "not logged in"
	} else {
		html = "logged in as " + user.UserInfo.GetSpec().GetEmail()
	}

	response := components.SimpleResponse{
		Body: []byte(html),
	}
	return response, nil
}
