package github

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v45/github"
	"github.com/justinsb/kweb/components/github/pb"
	"github.com/justinsb/kweb/components/kube"
	"github.com/justinsb/kweb/components/kube/kubeclient"
	"github.com/justinsb/kweb/debug"
	"google.golang.org/protobuf/encoding/prototext"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func buildInstallationKey(userID string, installationID int64) types.NamespacedName {
	return types.NamespacedName{
		Namespace: "user-" + userID,
		Name:      strconv.FormatInt(installationID, 10),
	}
}

func (c *Component) SyncInstallations(ctx context.Context) error {
	appClient, err := c.appClient(ctx)
	if err != nil {
		return err
	}

	// TODO: Use since parameter for more efficient fetching

	var opts github.ListOptions
	opts.PerPage = 100
	for {
		installations, response, err := appClient.Apps.ListInstallations(ctx, &opts)
		if err != nil {
			return fmt.Errorf("error listing installations: %w", err)
		}
		for _, installation := range installations {
			klog.Infof("github installation: %v", debug.JSON(installation))
			githubUserID := installation.GetAccount().GetID()

			kubeInstallation := &pb.AppInstallation{}
			kube.InitObject(kubeInstallation, types.NamespacedName{Namespace: fmt.Sprintf("github-%d", githubUserID), Name: strconv.FormatInt(installation.GetID(), 10)})
			kubeInstallation.Spec = &pb.AppInstallationSpec{
				Id: installation.GetID(),
				Account: &pb.GithubAccount{
					Id:    installation.GetAccount().GetID(),
					Login: installation.GetAccount().GetLogin(),
				},
			}
			klog.Infof("installation is %v", prototext.Format(kubeInstallation))

			if err := c.kube.Apply(ctx, kubeInstallation, kubeclient.ApplyOptions{FieldManager: "kweb-github"}); err != nil {
				return fmt.Errorf("error applying installation object: %w", err)
			}
		}
		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return nil
}
