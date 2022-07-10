package users

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/kube"
	"github.com/justinsb/kweb/components/kube/kubeclient"
	userapi "github.com/justinsb/kweb/components/users/pb"
	"golang.org/x/oauth2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type UserComponent struct {
	kube *kubeclient.Client
}

func NewUserComponent(kube *kubeclient.Client) (*UserComponent, error) {
	c := &UserComponent{
		kube: kube,
	}

	return c, nil
}

var _ components.RequestFilter = &UserComponent{}

func (c *UserComponent) ProcessRequest(ctx context.Context, req *components.Request, next components.RequestFilterChain) (components.Response, error) {
	user, err := c.userFromSession(ctx)
	if err != nil {
		return nil, err
	}
	if user != nil {
		ctx = context.WithValue(ctx, contextKeyUser, user)
	}

	return next(ctx, req)
}

func (c *UserComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) {
}

var contextKeyUser = &User{}

func GetUser(ctx context.Context) *User {
	v := ctx.Value(contextKeyUser)
	if v == nil {
		return nil
	}
	return v.(*User)
}

func buildUserKey(userID string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: "user-" + userID,
		Name:      userID,
	}
}
func (c *UserComponent) MapToUser(ctx context.Context, req *components.Request, token *oauth2.Token, info *components.AuthenticationInfo) (*userapi.User, error) {
	// TODO: When namespace == name, should we make it cluster scoped and shard them differently?
	// Although then we are expressing that we don't normally read all these objects consistently, when we split by namespace

	providerID := info.Provider.ProviderID()

	// TODO: We really need an index!
	usersClient := kubeclient.TypedClient(c.kube, &userapi.User{})
	allUsers, err := usersClient.List(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, user := range allUsers {
		match := false
		for _, linkedAccount := range user.GetSpec().GetLinkedAccounts() {
			if linkedAccount.GetProviderID() != providerID {
				continue
			}
			if linkedAccount.GetProviderUserID() == info.ProviderUserID {
				match = true
			}
		}
		if match {
			return user, nil
		}
	}

	userID := generateUserID()
	userSpec, err := info.Provider.PopulateUserData(ctx, token, info)
	if err != nil {
		return nil, fmt.Errorf("failed to build user info: %w", err)
	}

	userSpec.LinkedAccounts = append(userSpec.LinkedAccounts, &userapi.LinkedAccount{
		ProviderID:       providerID,
		ProviderUserID:   info.ProviderUserID,
		ProviderUserName: info.ProviderUserName,
	})

	userKey := buildUserKey(userID)
	user := &userapi.User{}
	kube.InitObject(user, userKey)
	user.Spec = userSpec

	if err := c.ensureNamespace(ctx, user.Metadata.Namespace); err != nil {
		return nil, err
	}

	// TODO: It is possible that we create two users simultaneously here
	// We likely need to support merging users (which we probably need to do anyway if we support login with multiple accounts)
	// TODO: We could use a SHA of the email (assuming we can get it)
	if err := c.kube.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func generateUserID() string {
	b := make([]byte, 16, 16)
	if _, err := cryptorand.Read(b); err != nil {
		klog.Fatalf("error building user id: %v", err)
	}
	sessionID := hex.EncodeToString(b)
	return sessionID
}

func (c *UserComponent) ensureNamespace(ctx context.Context, namespaceName string) error {
	namespaces := c.kube.Dynamic().Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"})
	ns, err := namespaces.Get(ctx, namespaceName, v1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("error reading namespace %q: %w", namespaceName, err)
		}
	} else {
		return nil
	}

	ns = &unstructured.Unstructured{}
	ns.SetAPIVersion("v1")
	ns.SetName(namespaceName)
	if _, err := namespaces.Create(ctx, ns, v1.CreateOptions{}); err != nil {
		// TODO: Deal with concurrent creation?
		return fmt.Errorf("error creating namespace %q: %w", namespaceName, err)
	}

	return nil
}
