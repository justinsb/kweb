package main

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

func main() {
	err := run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "unexpected error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	rand.Seed(time.Now().UnixNano())
	klog.InitFlags(nil)

	listen := ":8081"
	flag.StringVar(&listen, "listen", listen, "endpoint on which to start http server")

	flag.Parse()

	s, err := server.New()
	if err != nil {
		return err
	}

	if err := s.ListenAndServe(ctx, listen, nil); err != nil {
		return fmt.Errorf("error running http server: %w", err)
	}
	return nil
}
