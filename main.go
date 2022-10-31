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
	"k8s.io/klog"
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
	err := a.listenAndServe(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "unexpected error: %v\n", err)
		os.Exit(1)
	}
}

func (a *App) listenAndServe(ctx context.Context) error {
	if err := a.server.ListenAndServe(ctx, a.options.Server.Listen, nil); err != nil {
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
