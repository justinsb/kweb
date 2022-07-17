package users

import (
	"context"
	"fmt"

	"github.com/justinsb/kweb/components"
	userapi "github.com/justinsb/kweb/components/users/pb"
	"golang.org/x/oauth2"
)

const sessionKeyUserID = "userid"

func (c *UserComponent) userFromSession(ctx context.Context) (*userapi.User, error) {
	request := components.GetRequest(ctx)

	userID := request.Session.GetString(sessionKeyUserID)
	if userID == "" {
		return nil, nil
	}
	user := &userapi.User{}
	key := buildUserKey(userID)
	if err := c.kube.Get(ctx, key, user); err != nil {
		// apierrors.IsNotFound would be unexpected here; the userid is set in the session
		return nil, fmt.Errorf("error fetching user %v: %w", key, err)
	}

	return user, nil
}

func Logout(ctx context.Context) {
	request := components.GetRequest(ctx)

	request.Session.Clear(sessionKeyUserID)

}

func SetCurrentUser(ctx context.Context, token *oauth2.Token, providerKey string, user *userapi.User) error {

	// TODO: Where should we store the token?

	req := components.GetRequest(ctx)
	userID := user.Metadata.Name
	req.Session.SetString(sessionKeyUserID, userID)

	return nil
}
