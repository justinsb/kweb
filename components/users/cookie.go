package users

import (
	"context"
	"fmt"

	"github.com/justinsb/kweb/components"
	userapi "github.com/justinsb/kweb/components/users/pb"
)

const sessionKeyUserID = "userid"

func (c *UserComponent) userFromSession(ctx context.Context) (*userapi.User, error) {
	request := components.GetRequest(ctx)

	userID := request.Session.GetString(sessionKeyUserID)
	if userID == "" {
		return nil, nil
	}
	user, err := c.LoadUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (c *UserComponent) LoadUser(ctx context.Context, userID string) (*userapi.User, error) {
	if userID == "" {
		return nil, nil
	}
	user := &userapi.User{}
	key := c.buildUserKey(userID)
	if err := c.kube.Get(ctx, key, user); err != nil {
		// apierrors.IsNotFound would be unexpected here; the userid is set in the session
		return nil, fmt.Errorf("error fetching user %v: %w", key, err)
	}

	return user, nil
}

func Logout(ctx context.Context) {
	SetUser(ctx, nil)
}
