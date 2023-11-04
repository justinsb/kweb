package oidc

import (
	"context"

	userapi "github.com/justinsb/kweb/components/users/pb"
	"k8s.io/klog/v2"
)

func (a *Authenticator) UserFromJWT(ctx context.Context, token string) (*userapi.User, error) {
	tokenInfo, ok := a.tryAuthenticateJWT(ctx, token)
	if !ok {
		return nil, nil
	}

	var rawClaims map[string]interface{}
	tokenInfo.token.Claims(&rawClaims)
	// klog.Infof("rawClaims is %#v", rawClaims)

	userID := tokenInfo.token.Subject
	if userID != "" {
		user, err := a.users.LoadUser(ctx, userID)
		if err != nil {
			klog.Warningf("failed to load user: %v", err)
			return nil, err
		}
		if user == nil {
			klog.Warningf("user %q was in JWT but was not found", userID)
			return nil, nil
		}
		return user, nil
	}
	return nil, nil
}
