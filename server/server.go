package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	// _ "github.com/lib/pq"
	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/cookies"
	"github.com/justinsb/kweb/components/kube/kubeclient"
	"github.com/justinsb/kweb/components/login/providers"
	"github.com/justinsb/kweb/components/login/providers/loginwithgoogle"
	"github.com/justinsb/kweb/components/sessions"
	"github.com/justinsb/kweb/components/users"

	"k8s.io/klog/v2"
)

type Server struct {
	components.Server
}

func New() (*Server, error) {
	restConfig, err := GetRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes configuration: %w", err)
	}

	s := &Server{}

	cookiesComponent := cookies.NewCookiesComponent()
	s.Components = append(s.Components, cookiesComponent)

	sessionComponent := sessions.NewSessionComponent()
	s.Components = append(s.Components, sessionComponent)

	kubeClient, err := kubeclient.New(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes controller client: %w", err)
	}

	userComponent, err := users.NewUserComponent(kubeClient)
	if err != nil {
		return nil, fmt.Errorf("error building user component: %w", err)
	}
	s.Components = append(s.Components, userComponent)

	clientID := os.Getenv("OAUTH2_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH2_CLIENT_SECRET")
	googleProvider, err := providers.NewGoogleProvider("google", clientID, clientSecret)
	loginComponent, err := loginwithgoogle.NewComponent(userComponent, googleProvider)
	if err != nil {
		return nil, fmt.Errorf("error building login component: %w", err)
	}
	s.Components = append(s.Components, loginComponent)

	return s, nil
}

func (s *Server) ListenAndServe(ctx context.Context, listen string, listening chan<- net.Addr) error {
	defer close(listening)

	klog.Infof("starting server on %q", listen)

	httpServer := &http.Server{
		Addr:           listen,
		Handler:        s,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctxWithCancel.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		if err := httpServer.Shutdown(shutdownContext); err != nil {
			klog.Warningf("error shutting down http server: %v", err)
		}
		if err := httpServer.Close(); err != nil {
			klog.Warningf("error closing http server: %v", err)
		}
	}()

	ln, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("error listening on %q: %w", listen, err)
	}
	if listening != nil {
		listening <- ln.Addr()
	}
	if err := httpServer.Serve(ln); err != nil {
		if ctxWithCancel.Err() != nil {
			// Shutdown through context
			return nil
		}
		return fmt.Errorf("error running http server: %w", err)
	}

	return nil
}
