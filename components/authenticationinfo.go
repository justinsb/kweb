package components

import (
	"context"

	userapi "github.com/justinsb/kweb/components/users/pb"
	"golang.org/x/oauth2"
)

type AuthenticationInfo struct {
	Provider       AuthenticationProvider
	ProviderUserID string
}

// AuthenticationProvider is implemented by authentication services.
type AuthenticationProvider interface {
	ProviderID() string
	Redeem(ctx context.Context, redirectURI string, code string) (*AuthenticationInfo, *oauth2.Token, error)
	GetLoginURL(ctx context.Context, redirectURI, state string) string

	PopulateUserData(ctx context.Context, token *oauth2.Token, info *AuthenticationInfo) (*userapi.UserSpec, error)
}
