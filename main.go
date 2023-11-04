package kweb

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/users"
	"github.com/justinsb/kweb/server"
	"github.com/justinsb/packages/kinspire/client"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"k8s.io/klog/v2"
)

type App struct {
	options Options

	server *server.Server
}

type Options struct {
	Server server.Options
}

func NewApp(opt *Options) (*App, error) {
	a := &App{options: *opt}

	rand.Seed(time.Now().UnixNano())
	klog.InitFlags(nil)

	flag.StringVar(&opt.Server.Listen, "listen", opt.Server.Listen, "endpoint on which to start http server")

	flag.Parse()

	s, err := server.New(a.options.Server)
	if err != nil {
		return nil, err
	}
	a.server = s
	return a, nil
}

func NewOptions(appName string) *Options {
	o := &Options{}
	o.InitDefaults(appName)
	return o
}

func (o *Options) InitDefaults(appName string) {
	o.Server.InitDefaults(appName)
}

func (a *App) RunFromMain() {
	err := a.run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "unexpected error: %v\n", err)
		os.Exit(1)
	}
}

func (a *App) run(ctx context.Context) error {
	if a.options.Server.UseSPIFFE {
		if err := client.SPIFFE.Init(ctx); err != nil {
			return fmt.Errorf("error initializing SPIFFE")
		}

		source := client.SPIFFE.Source()

		svid, err := source.GetX509SVID()
		if err != nil {
			return err
		}
		klog.Infof("my x509 is %v", svid)

		// Allowed SPIFFE ID
		clientID := spiffeid.RequireFromString("spiffe://k8s.local/ns/default/sa/gateway-instance")

		klog.Infof("creating httpserver, requires %v", clientID)
		// Create a `tls.Config` to allow mTLS connections, and verify that presented certificate has the specified SPIFFE ID
		tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeID(clientID))
		a.options.Server.TLSConfig = tlsConfig
	}
	return a.listenAndServe(ctx)
}

func (a *App) listenAndServe(ctx context.Context) error {
	if err := a.server.ListenAndServe(ctx, a.options.Server.Listen, a.options.Server.TLSConfig, nil); err != nil {
		return fmt.Errorf("error running http server: %w", err)
	}
	return nil
}

func (a *App) AddComponent(component components.Component) {
	a.server.Components = append(a.server.Components, component)
}

func (a *App) Users() *users.UserComponent {
	var userComponent *users.UserComponent
	if err := components.GetComponent(&a.server.Server, &userComponent); err != nil {
		klog.Fatalf("error getting user component: %v", err)
	}
	return userComponent
}
