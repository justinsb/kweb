package loginwithgithub

import (
	"context"
	"fmt"
	"strconv"

	githubapi "github.com/google/go-github/v45/github"
	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/users"
	userapi "github.com/justinsb/kweb/components/users/pb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"k8s.io/klog/v2"
)

type GithubProvider struct {
	conf       *oauth2.Config
	userMapper components.UserMapper
}

const ProviderID = "github"

func NewGithubProvider(clientID, clientSecret string, userMapper components.UserMapper) (*GithubProvider, error) {
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		// Scopes: not needed with Github App auth
		Endpoint: github.Endpoint,
	}
	return &GithubProvider{
		conf:       conf,
		userMapper: userMapper,
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

func (p *GithubProvider) Redeem(ctx context.Context, redirectURL string, code string) error {
	conf := *p.conf
	conf.RedirectURL = redirectURL

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to redeem token: %w", err)
	}

	userInfo, err := p.userInfoForToken(ctx, token)
	if err != nil {
		return err
	}

	info := &components.AuthenticationInfo{
		Provider:         p,
		ProviderUserID:   strconv.FormatInt(userInfo.GetID(), 10),
		ProviderUserName: userInfo.GetLogin(),
		PopulateUserData: p.PopulateUserData,
	}

	// set cookie, or deny
	user, err := p.userMapper.MapToUser(ctx, token, info)
	if err != nil {
		klog.Infof("error mapping to user: %v", err)
		return err
	}

	klog.Infof("authentication complete %v", info)

	users.SetUser(ctx, user)

	return nil
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
