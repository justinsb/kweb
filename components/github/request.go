package github

import (
	"context"

	"github.com/google/go-github/v45/github"
	"github.com/justinsb/kweb/components/github/pb"
	"github.com/justinsb/kweb/components/kube/kubeclient"
	"github.com/justinsb/kweb/components/users"
	"k8s.io/klog/v2"
)

func GetClient(ctx context.Context, options github.InstallationTokenOptions) (*Client, error) {
	requestInfo := getRequestInfo(ctx)
	if requestInfo == nil {
		return nil, nil
	}
	return requestInfo.GetClient(ctx, options)
}

var contextKeyRequest = &requestInfo{}

func getRequestInfo(ctx context.Context) *requestInfo {
	v := ctx.Value(contextKeyRequest)
	if v == nil {
		return nil
	}
	return v.(*requestInfo)
}

type requestInfo struct {
	component *Component
}

func (r *requestInfo) GetClient(ctx context.Context, options github.InstallationTokenOptions) (*Client, error) {
	user := users.GetUser(ctx)
	if user == nil {
		return nil, nil
	}

	ns := user.UserInfo.GetMetadata().GetNamespace()
	typed := kubeclient.TypedClient(r.component.kube, &pb.AppInstallation{})
	installations, err := typed.List(ctx, ns)
	if err != nil {
		return nil, err
	}

	if len(installations) == 0 {
		return nil, nil
	}

	if len(installations) > 1 {
		klog.Infof("using first installation, but multiple installations found for user %q", ns)
	}
	installation := installations[0]

	client, err := r.component.authForInstallation(ctx, installation.GetSpec().GetId(), options)
	if err != nil {
		return nil, err
	}
	return client, nil
}
