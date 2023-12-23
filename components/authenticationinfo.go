package components

import (
	"context"

	userapi "github.com/justinsb/kweb/components/users/pb"
	"golang.org/x/oauth2"
)

type AuthenticationInfo struct {
	Provider         AuthenticationProvider
	ProviderUserID   string
	ProviderUserName string

	PopulateUserData func(ctx context.Context, token *oauth2.Token, info *AuthenticationInfo) (*userapi.UserSpec, error)
}

// AuthenticationProvider is implemented by authentication services.
type AuthenticationProvider interface {
	ProviderID() string
	// Redeem is called from the oauth2 callback request; it normally exchanges a code for a token for the logged-in user
	Redeem(ctx context.Context, redirectURI string, code string) error
	GetLoginURL(ctx context.Context, redirectURI, state string) string
}

type UserMapper interface {
	MapToUser(ctx context.Context, token *oauth2.Token, info *AuthenticationInfo) (*userapi.User, error)
}
