package loginwithgoogle

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/justinsb/kweb/components"
	userapi "github.com/justinsb/kweb/components/users/pb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jws"
	"k8s.io/klog/v2"
)

type GoogleProvider struct {
	providerKey string
	conf        *oauth2.Config
}

func NewGoogleProvider(providerKey string, clientID, clientSecret string) (*GoogleProvider, error) {
	// Discovery document is at https://accounts.google.com/.well-known/openid-configuration

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		//RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
	return &GoogleProvider{
		providerKey: providerKey,
		conf:        conf,
	}, nil
}

func (p *GoogleProvider) ProviderID() string {
	return p.providerKey
}

func (p *GoogleProvider) GetLoginURL(ctx context.Context, redirectURL string, state string) string {
	conf := *p.conf
	conf.RedirectURL = redirectURL

	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL(state)
	return url
}

func (p *GoogleProvider) Redeem(ctx context.Context, redirectURL string, code string) (*components.AuthenticationInfo, *oauth2.Token, error) {
	conf := *p.conf
	conf.RedirectURL = redirectURL

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to redeem token: %w", err)
	}

	// resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")

	// We don't need to call jws.Verify because we got this token direct from google over HTTPS

	idToken := token.Extra("id_token")
	if idToken == nil {
		return nil, nil, fmt.Errorf("id_token was not found")
	}
	idTokenString, ok := idToken.(string)
	if !ok {
		return nil, nil, fmt.Errorf("id_token was not string")
	}
	claimSet, err := jws.Decode(idTokenString)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding id_token: %w", err)
	}

	if claimSet.Iss != "accounts.google.com" && claimSet.Iss != "https://accounts.google.com" {
		return nil, nil, fmt.Errorf("unexpected issuer %q", claimSet.Iss)
	}

	if claimSet.Sub == "" {
		return nil, nil, fmt.Errorf("JWT did not contain 'sub' value")
	}

	info := &components.AuthenticationInfo{
		Provider:       p,
		ProviderUserID: claimSet.Sub,
	}

	return info, token, nil
}

type oidcUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Domain        string `json:"hd"`
}

func fetchUserInfo(ctx context.Context, client *http.Client, token *oauth2.Token) (*oidcUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("error building HTTP request: %w", err)
	}
	req.Header.Add("Authorization", token.TokenType+" "+token.AccessToken)

	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing HTTP request: %w", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading HTTP response: %w", err)
	}

	klog.Infof("full user info response: %v", string(b))

	userInfo := &oidcUserInfo{}
	if err := json.Unmarshal(b, userInfo); err != nil {
		return nil, fmt.Errorf("error parsing] HTTP response: %w", err)
	}

	return userInfo, nil
}

func (p *GoogleProvider) PopulateUserData(ctx context.Context, token *oauth2.Token, authInfo *components.AuthenticationInfo) (*userapi.UserSpec, error) {
	userInfo, err := fetchUserInfo(ctx, http.DefaultClient, token)
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	if !userInfo.EmailVerified {
		return nil, fmt.Errorf("user email is not verified")
	}

	spec := &userapi.UserSpec{}
	spec.Email = userInfo.Email
	return spec, nil
}
