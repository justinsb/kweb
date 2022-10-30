package kweb

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"k8s.io/klog/v2"

	"github.com/justinsb/kweb/server"
)

type App struct {
	options Options
}

type Options struct {
	Server server.Options
}

func NewApp(opt *Options) *App {
	a := &App{options: *opt}
	return a
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
	rand.Seed(time.Now().UnixNano())
	klog.InitFlags(nil)

	listen := ":8080"
	flag.StringVar(&listen, "listen", listen, "endpoint on which to start http server")

	flag.Parse()

	s, err := server.New(a.options.Server)
	if err != nil {
		return err
	}

	if err := s.ListenAndServe(ctx, listen, nil); err != nil {
		return fmt.Errorf("error running http server: %w", err)
	}
	return nil
}
