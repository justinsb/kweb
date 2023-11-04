package github

import (
	"context"
	"net/http"

	"github.com/justinsb/kweb/components"
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
