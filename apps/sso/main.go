package main

import (
	"context"
	"embed"
	"os"

	"github.com/justinsb/kweb"
	"github.com/justinsb/kweb/apps/sso/components/jwtissuer"
	"github.com/justinsb/kweb/apps/sso/components/oidclogin"
	"github.com/justinsb/kweb/components/keystore"
	"github.com/justinsb/kweb/components/keystore/pb"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
)

//go:embed pages
var pages embed.FS

func main() {
	opt := kweb.NewOptions("kweb-sso-system")
	opt.Server.Pages.Base = pages

	app, err := kweb.NewApp(opt)
	if err != nil {
		klog.Fatalf("error building app: %v", err)
	}

	restConfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Fatalf("error getting kubernetes config: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.Fatalf("error building kubernetes client: %v", err)
	}
	keystore, err := keystore.NewKubernetesKeyStore(kubeClient, "kweb-sso-system", "oidc-keys")
	if err != nil {
		klog.Fatalf("error building kubernetes keystore: %v", err)
	}
	ctx := context.Background()
	keys, err := keystore.KeySet(ctx, "oidc-keys", pb.KeyType_KEYTYPE_RSA)
	if err != nil {
		klog.Fatalf("error building kubernetes keys: %v", err)
	}

	issuer := os.Getenv("OIDC_ISSUER")
	audience := os.Getenv("OIDC_AUDIENCE")

	app.AddComponent(&jwtissuer.JWTIssuerComponent{
		Keys:     keys,
		Issuer:   issuer,
		Audience: audience,
	})
	userComponent := app.Users()
	oidcLogin := oidclogin.NewOIDCLoginComponent(ctx, oidclogin.Options{
		Issuer:   issuer,
		Audience: audience,
	}, userComponent)
	app.AddComponent(oidcLogin)

	app.RunFromMain()
}
