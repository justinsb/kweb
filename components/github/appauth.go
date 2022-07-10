package github

import (
	"context"
	"crypto/rsa"
	"fmt"
	"sync"
	"time"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
	"k8s.io/klog/v2"
)

// Github apps authenticate using a private key to create a JWT token
// https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-a-github-app
//
// They can then use that app token to generate a user-scoped token (technically, an installation scoped token)
// https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-an-installation

func (c *Component) authForInstallation(ctx context.Context, installationID int64, options github.InstallationTokenOptions) (*Client, error) {
	// TODO: Cache userTokenSource

	ts := &userTokenSource{
		installationID: installationID,
		component:      c,
		options:        options,
	}
	tc := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(tc)
	return &Client{
		Client: ghClient,
	}, nil
}

type userTokenSource struct {
	component      *Component
	installationID int64
	options        github.InstallationTokenOptions

	mutex sync.Mutex
	token *oauth2.Token
}

var _ oauth2.TokenSource = &userTokenSource{}

// Token returns a token or an error.
// Token must be safe for concurrent use by multiple goroutines.
// The returned Token must not be modified.
func (t *userTokenSource) Token() (*oauth2.Token, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.token != nil {
		now := time.Now()
		if t.token.Expiry.Before(now.Add(time.Minute)) {
			klog.Infof("github token has less than one minute remaining (%v vs %v), will treat as expired", t.token.Expiry, now)
		} else {
			return t.token, nil
		}
	}

	ctx := context.TODO()
	appClient, err := t.component.appClient(ctx)
	if err != nil {
		return nil, err
	}

	ghToken, _, err := appClient.Apps.CreateInstallationToken(ctx, t.installationID, &t.options)
	if err != nil {
		return nil, fmt.Errorf("failed to create github installation-scoped token for %v: %w", t.installationID, err)
	}

	token := &oauth2.Token{
		AccessToken: ghToken.GetToken(),
		Expiry:      ghToken.GetExpiresAt(),
	}
	t.token = token
	// TODO: Expose Permissions and Repositories?
	// Permissions  *InstallationPermissions `json:"permissions,omitempty"`
	// Repositories []*Repository            `json:"repositories,omitempty"`

	return token, nil
}

func (c *Component) appClient(ctx context.Context) (*github.Client, error) {
	// TODO: Caching?
	ts := &appTokenSource{
		appPrivateKey: c.appPrivateKey,
		githubAppID:   c.githubAppID,
	}
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return client, nil
}

type appTokenSource struct {
	appPrivateKey *rsa.PrivateKey
	githubAppID   string

	mutex sync.Mutex
	token *oauth2.Token
}

var _ oauth2.TokenSource = &appTokenSource{}

// Token returns a token or an error.
// Token must be safe for concurrent use by multiple goroutines.
// The returned Token must not be modified.
func (t *appTokenSource) Token() (*oauth2.Token, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.token != nil {
		now := time.Now()
		if t.token.Expiry.Before(now.Add(time.Minute)) {
			klog.Infof("github app token has less than one minute remaining (%v vs %v), will treat as expired", t.token.Expiry, now)
		} else {
			return t.token, nil
		}
	}

	token, err := t.buildJWT()
	if err != nil {
		return nil, err
	}

	t.token = token

	return token, nil
}

func (t *appTokenSource) buildJWT() (*oauth2.Token, error) {
	now := time.Now()

	expiry := now.Add(time.Minute * 10) // max expiry
	claimSet := &jws.ClaimSet{
		Iss: t.githubAppID,
		Iat: now.Add(-time.Minute).Unix(), // allow for clock drift
		Exp: expiry.Unix(),                // max expiry
	}
	hdr := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
		//KeyID: ???
	}
	msg, err := jws.Encode(hdr, claimSet, t.appPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error encoding JWT: %w", err)
	}
	return &oauth2.Token{AccessToken: msg, TokenType: "Bearer", Expiry: expiry}, nil
}
