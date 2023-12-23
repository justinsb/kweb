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
	"github.com/justinsb/kweb/templates/scopes"
	"golang.org/x/oauth2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type UserComponent struct {
	kube            *kubeclient.Client
	namespaceMapper NamespaceMapper
}

type NamespaceMapper interface {
	NamespaceForUserID(userID string) string
}

type SingleNamespaceMapper struct {
	namespace string
}

func NewSingleNamespaceMapper(ns string) *SingleNamespaceMapper {
	return &SingleNamespaceMapper{namespace: ns}
}

func (m *SingleNamespaceMapper) NamespaceForUserID(userID string) string {
	return m.namespace
}

type NamespacePerUser struct {
	prefix string
}

func (m *NamespacePerUser) NamespaceForUserID(userID string) string {
	return m.prefix + userID
}

func NewUserComponent(kube *kubeclient.Client, namespaceMapper NamespaceMapper) (*UserComponent, error) {
	c := &UserComponent{
		kube:            kube,
		namespaceMapper: namespaceMapper,
	}

	return c, nil
}

var _ components.RequestFilter = &UserComponent{}

func (c *UserComponent) ProcessRequest(ctx context.Context, req *components.Request, next components.RequestFilterChain) (components.Response, error) {
	user, err := c.userFromSession(ctx)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, contextKeyUser, &scopeInfo{currentUser: user})

	return next(ctx, req)
}

func (c *UserComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	return nil
}

func (c *UserComponent) AddToScope(ctx context.Context, scope *scopes.Scope) {
	scope.Values["user"] = scopes.Value{
		Function: func() interface{} {
			return GetUser(ctx)
		},
	}
}

var contextKeyUser = &scopeInfo{}

type scopeInfo struct {
	currentUser *userapi.User
}

func SetUser(ctx context.Context, user *userapi.User) {
	info := ctx.Value(contextKeyUser)
	if info == nil {
		klog.Fatalf("user component not configured (key not in context)")
	}
	info.(*scopeInfo).currentUser = user

	req := components.GetRequest(ctx)
	if user != nil {
		userID := user.Metadata.Name
		userInfo := &userapi.UserSessionInfo{
			UserId: userID,
		}
		req.Session.Set(userInfo)
	} else {
		req.Session.Clear(&userapi.UserSessionInfo{})
	}
}

func GetUser(ctx context.Context) *userapi.User {
	info := ctx.Value(contextKeyUser).(*scopeInfo)
	return info.currentUser
}

func (c *UserComponent) buildUserKey(userID string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: c.namespaceMapper.NamespaceForUserID(userID),
		Name:      userID,
	}
}

func (c *UserComponent) MapToUser(ctx context.Context, token *oauth2.Token, info *components.AuthenticationInfo) (*userapi.User, error) {
	// TODO: When namespace == name, should we make it cluster scoped and shard them differently?
	// Although then we are expressing that we don't normally read all these objects consistently, when we split by namespace

	providerID := info.Provider.ProviderID()

	// TODO: We really need an index!
	usersClient := kubeclient.TypedClient(c.kube, &userapi.User{})
	// A bit of a hack!
	namespace := ""
	switch nsStrategy := c.namespaceMapper.(type) {
	case *SingleNamespaceMapper:
		namespace = nsStrategy.namespace
	case *NamespacePerUser:
		namespace = ""
		klog.Warningf("doing very inefficient all-namespace scan")
	default:
		return nil, fmt.Errorf("unknown namespace strategy %T", nsStrategy)
	}
	allUsers, err := usersClient.List(ctx, namespace)
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
	userSpec, err := info.PopulateUserData(ctx, token, info)
	if err != nil {
		return nil, fmt.Errorf("failed to build user info: %w", err)
	}

	userSpec.LinkedAccounts = append(userSpec.LinkedAccounts, &userapi.LinkedAccount{
		ProviderID:       providerID,
		ProviderUserID:   info.ProviderUserID,
		ProviderUserName: info.ProviderUserName,
	})

	userKey := c.buildUserKey(userID)
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
