package oauthsessions

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/justinsb/kweb/components/oauthsessions/api"
	"github.com/justinsb/kweb/components/users"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func (c *OAuthSessionsComponent) StoreSession(ctx context.Context, token *oauth2.Token) error {
	user := users.GetUser(ctx)
	if user == nil {
		return fmt.Errorf("no user found")
	}

	sessionID := randomID()

	sessionKey := types.NamespacedName{
		Namespace: user.Metadata.Namespace,
		Name:      sessionID,
	}
	session := &api.OauthSession{}
	session.Name = sessionKey.Name
	session.Namespace = sessionKey.Namespace

	session.Spec = api.OauthSessionSpec{
		User:         user.GetMetadata().GetName(),
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry.Unix(),
	}

	if err := c.kube.Uncached().Create(ctx, session); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func randomID() string {
	// TODO: Create ids packages
	b := make([]byte, 32)
	if _, err := cryptorand.Read(b); err != nil {
		klog.Fatalf("building random id: %v", err)
	}
	sessionID := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	sessionID = strings.ToLower(sessionID)
	return sessionID
}
