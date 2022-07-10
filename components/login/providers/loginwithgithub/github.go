package loginwithgithub

import (
	"context"
	"fmt"
	"strconv"

	githubapi "github.com/google/go-github/v45/github"
	"github.com/justinsb/kweb/components"
	userapi "github.com/justinsb/kweb/components/users/pb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GithubProvider struct {
	conf *oauth2.Config
}

const ProviderID = "github"

func NewGithubProvider(clientID, clientSecret string) (*GithubProvider, error) {
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		// Scopes: not needed with Github App auth
		Endpoint: github.Endpoint,
	}
	return &GithubProvider{
		conf: conf,
	}, nil
}

func (p *GithubProvider) ProviderID() string {
	return ProviderID
}

func (p *GithubProvider) GetLoginURL(ctx context.Context, redirectURL string, state string) string {
	conf := *p.conf
	conf.RedirectURL = redirectURL

	url := conf.AuthCodeURL(state)
	return url
}

func (p *GithubProvider) Redeem(ctx context.Context, redirectURL string, code string) (*components.AuthenticationInfo, *oauth2.Token, error) {
	conf := *p.conf
	conf.RedirectURL = redirectURL

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to redeem token: %w", err)
	}

	userInfo, err := p.userInfoForToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	info := &components.AuthenticationInfo{
		Provider:         p,
		ProviderUserID:   strconv.FormatInt(userInfo.GetID(), 10),
		ProviderUserName: userInfo.GetLogin(),
	}

	return info, token, nil
}

func (p *GithubProvider) PopulateUserData(ctx context.Context, token *oauth2.Token, authInfo *components.AuthenticationInfo) (*userapi.UserSpec, error) {
	userInfo, err := p.userInfoForToken(ctx, token)
	if err != nil {
		return nil, err
	}
	// TODO: Are github emails verified?

	spec := &userapi.UserSpec{}
	spec.Email = userInfo.GetEmail()
	return spec, nil
}

func (p *GithubProvider) userInfoForToken(ctx context.Context, token *oauth2.Token) (*githubapi.User, error) {

	// TODO: Cache / micro-cache this information?

	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(ctx, ts)

	githubClient := githubapi.NewClient(tc)

	userInfo, _, err := githubClient.Users.Get(ctx, "") // passing the empty string gets the current user
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	return userInfo, nil
}
