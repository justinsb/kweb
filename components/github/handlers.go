package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v45/github"
	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/kube/kubeclient"
	"github.com/justinsb/kweb/components/login/providers/loginwithgithub"
	"github.com/justinsb/kweb/components/users"
	"github.com/justinsb/kweb/components/users/pb"
	"k8s.io/klog/v2"
)

func (p *Component) doEntryPoint(ctx context.Context, req *components.Request) (components.Response, error) {
	err := req.ParseForm()
	if err != nil {
		return components.ErrorResponse(http.StatusBadRequest), err
	}

	code := req.FormValue("code")
	installationID := req.FormValue("installation_id")
	setupAction := req.FormValue("setup_action")

	klog.Infof("code = %q", code)
	klog.Infof("installation_id = %q", installationID)
	klog.Infof("setup_action = %q", setupAction)

	return components.RedirectResponse("/"), nil
}

// DebugInfo is a simple endpoint for debugging, while we can't do much more
// TODO: Remove me!
func (c *Component) DebugInfo(ctx context.Context, req *components.Request) (components.Response, error) {
	user := users.GetUser(ctx)
	var html string
	if user == nil {
		html = "not logged in"
	} else {
		html = "logged in as " + user.UserInfo.GetSpec().GetEmail()
	}

	html += "<br/>"

	var options github.InstallationTokenOptions
	options.Permissions = &github.InstallationPermissions{
		Issues:   github.String("read"),
		Metadata: github.String("read"),
	}
	client, err := GetClient(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}

	githubUserID := ""
	{
		// Note: need to all namespaces
		// TODO: Fix this
		typedClient := kubeclient.TypedClient(c.kube, &pb.UserAuth{})
		auths, err := typedClient.List(ctx, "")
		if err != nil {
			return nil, err
		}
		for _, auth := range auths {
			if auth.GetSpec().GetUserID() != user.UserInfo.GetMetadata().GetName() {
				continue
			}

			if auth.GetSpec().GetProviderID() == loginwithgithub.ProviderID {
				githubUserID = auth.GetSpec().GetProviderUserName()
			}
		}
	}
	if client != nil && githubUserID != "" {
		githubUser, _, err := client.Users.Get(ctx, githubUserID)
		if err != nil {
			html += fmt.Sprintf("error listing user: %v", err)
		} else {
			html += fmt.Sprintf("%s<br/>", githubUser.GetURL())
		}
	}

	if client != nil && githubUserID != "" {
		var opts github.RepositoryListOptions
		respositories, _, err := client.Repositories.List(ctx, githubUserID, &opts)
		if err != nil {
			html += fmt.Sprintf("error listing repos: %v", err)
		}
		for _, repository := range respositories {
			html += fmt.Sprintf("%s<br/>", repository.GetURL())
		}
	}

	if client != nil {
		var opts github.IssueListByRepoOptions
		issues, _, err := client.Issues.ListByRepo(ctx, "kubernetes", "kubernetes", &opts)
		if err != nil {
			html += fmt.Sprintf("error listing issues: %v", err)
		}
		for _, issue := range issues {
			html += fmt.Sprintf("%s<br/>", issue.GetURL())
		}
	}
	response := components.SimpleResponse{
		Body: []byte(html),
	}
	return response, nil
}
