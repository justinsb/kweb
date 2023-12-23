package oauthsessions

import (
	"context"
	"net/http"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/kube/kubeclient"
	"github.com/justinsb/kweb/components/oauthsessions/api"
	"github.com/justinsb/kweb/components/users"
	userapi "github.com/justinsb/kweb/components/users/pb"
	"github.com/justinsb/kweb/templates/scopes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OAuthSessionsComponent struct {
	kube *kubeclient.Client
}

func NewOAuthSessionsComponent(kube *kubeclient.Client) (*OAuthSessionsComponent, error) {
	c := &OAuthSessionsComponent{
		kube: kube,
	}

	return c, nil
}

func GetComponent(ctx context.Context) *OAuthSessionsComponent {
	var component *OAuthSessionsComponent
	components.GetComponent(ctx, &component)
	return component
}

func (c *OAuthSessionsComponent) AddToScope(ctx context.Context, scope *scopes.Scope) {

}

func (c *OAuthSessionsComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	return nil
}

// var contextKeyScopeInfo = &scopeInfo{}

type scopeInfo struct {
	// parent   *OAuthSessionsComponent
	sessions []*api.OauthSession
}

func GetAllOauthSessions(ctx context.Context) ([]*api.OauthSession, error) {
	// info := ctx.Value(contextKeyScopeInfo).(*scopeInfo)
	info := &scopeInfo{} // TODO: Caching
	return info.getAllOauthSessions(ctx)
}

func GetOauthSession(ctx context.Context) (*api.OauthSession, error) {
	sessions, err := GetAllOauthSessions(ctx)
	if err != nil {
		return nil, err
	}
	var best *api.OauthSession
	for _, session := range sessions {
		if best == nil {
			best = session
			continue
		}
		if best.Spec.ExpiresAt < session.Spec.ExpiresAt {
			best = session
		}
	}
	return best, nil
}

func (i *scopeInfo) getAllOauthSessions(ctx context.Context) ([]*api.OauthSession, error) {
	if i.sessions == nil {
		user := users.GetUser(ctx)
		if user != nil {
			sessions, err := GetComponent(ctx).LoadOauthSessions(ctx, user)
			if err != nil {
				return nil, err
			}
			i.sessions = sessions
		}
	}
	return i.sessions, nil
}

func (c *OAuthSessionsComponent) LoadOauthSessions(ctx context.Context, user *userapi.User) ([]*api.OauthSession, error) {
	if user == nil {
		return nil, nil
	}
	userID := user.GetMetadata().GetName()

	// TODO: We really need an index!
	allSessions := &api.OauthSessionList{}
	ns := user.GetMetadata().GetNamespace()
	if err := c.kube.Uncached().List(ctx, allSessions, client.InNamespace(ns)); err != nil {
		return nil, err
	}

	var matches []*api.OauthSession
	for i := range allSessions.Items {
		session := &allSessions.Items[i]
		if session.Spec.User == userID {
			matches = append(matches, session)
		}
	}

	return matches, nil
}
