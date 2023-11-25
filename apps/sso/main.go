package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/justinsb/kweb"
	"github.com/justinsb/kweb/apps/sso/components/jwtissuer"
	"github.com/justinsb/kweb/apps/sso/components/oidclogin"
	"github.com/justinsb/kweb/apps/sso/pkg/oidc"
	"github.com/justinsb/kweb/components/keystore"
	"github.com/justinsb/kweb/components/keystore/pb"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

//go:embed all:pages
var pages embed.FS

func main() {
	ctx := context.Background()

	log := klog.FromContext(ctx)

	opt := kweb.NewOptions("kweb-sso-system")
	opt.Server.Pages.Base = pages
	opt.Server.UseSPIFFE = true

	jwtIssuerOptions := jwtissuer.Options{}
	flag.StringVar(&jwtIssuerOptions.CookieDomain, "jwtIssuer.cookieDomain", jwtIssuerOptions.CookieDomain, "")

	oidcOptions := oidc.Options{}
	flag.StringVar(&oidcOptions.Issuer, "oidcLogin.issuer", oidcOptions.Issuer, "")
	flag.StringVar(&oidcOptions.Audience, "oidcLogin.audience", oidcOptions.Audience, "")

	var errors []error
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		name := f.Name
		envVar := name
		// envVar := strings.ReplaceAll(envVar, ".", "_")
		// envVar = strings.ToUpper(envVar)
		v := os.Getenv(envVar)
		if v != "" {
			if err := f.Value.Set(v); err != nil {
				errors = append(errors, fmt.Errorf("error setting flag %q to env var %q value %q: %w", name, envVar, v, err))
			}
		}
		log.Info("flag/env", "flag", f.Name, "env", envVar, "value", v)
	})
	if len(errors) != 0 {
		klog.Fatalf("error mapping env vars to flags: %v", errors[0])
	}

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
	keys, err := keystore.KeySet(ctx, "oidc-keys", pb.KeyType_KEYTYPE_RSA)
	if err != nil {
		klog.Fatalf("error building kubernetes keys: %v", err)
	}

	userComponent := app.Users()
	oidcAuthentiator := oidc.NewAuthenticator(userComponent, oidcOptions)

	oidcLoginComponent := oidclogin.NewOIDCLoginComponent(ctx, oidcAuthentiator, userComponent)
	app.AddComponent(oidcLoginComponent)

	jwtIssuer := jwtissuer.NewJWTIssuerComponent(keys, oidcAuthentiator, jwtIssuerOptions)
	app.AddComponent(jwtIssuer)

	app.RunFromMain()
}
