package server

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/cookies"
	"github.com/justinsb/kweb/components/github"
	"github.com/justinsb/kweb/components/healthcheck"
	"github.com/justinsb/kweb/components/kube/kubeclient"
	"github.com/justinsb/kweb/components/pages"

	// "github.com/justinsb/kweb/components/login/providers"
	"github.com/justinsb/kweb/components/login/providers/loginwithgithub"
	"github.com/justinsb/kweb/components/login/providers/loginwithgoogle"
	"github.com/justinsb/kweb/components/sessions"
	"github.com/justinsb/kweb/components/users"

	"k8s.io/klog/v2"
)

type Server struct {
	components.Server

	mutex sync.Mutex
	mux   *http.ServeMux
}

type Options struct {
	UserNamespaceStrategy users.NamespaceMapper
	Pages                 pages.Options
}

func (o *Options) InitDefaults(appName string) {
	o.UserNamespaceStrategy = users.NewSingleNamespaceMapper(appName)
	o.Pages.InitDefaults(appName)
}

func New(opt Options) (*Server, error) {
	restConfig, err := GetRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes configuration: %w", err)
	}

	s := &Server{}

	healthcheckComponent := healthcheck.NewHealthcheckComponent()
	s.Components = append(s.Components, healthcheckComponent)

	cookiesComponent := cookies.NewCookiesComponent()
	s.Components = append(s.Components, cookiesComponent)

	sessionComponent := sessions.NewSessionComponent()
	s.Components = append(s.Components, sessionComponent)

	pagesComponent := pages.New(opt.Pages)
	s.Components = append(s.Components, pagesComponent)

	kubeClient, err := kubeclient.New(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes controller client: %w", err)
	}

	userComponent, err := users.NewUserComponent(kubeClient, opt.UserNamespaceStrategy)
	if err != nil {
		return nil, fmt.Errorf("error building user component: %w", err)
	}
	s.Components = append(s.Components, userComponent)

	githubAppID := os.Getenv("GITHUB_APP_ID")
	if githubAppID != "" {
		// TODO: Get from kube secret or file?
		if os.Getenv("GITHUB_APP_KEY") == "" {
			return nil, fmt.Errorf("expected GITHUB_APP_KEY to be set")
		}
		rsaPrivateKey, err := parsePrivateKey(os.Getenv("GITHUB_APP_KEY"))
		if err != nil {
			return nil, err
		}
		githubApp, err := github.New(kubeClient, githubAppID, rsaPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("error building github component: %w", err)
		}
		s.Components = append(s.Components, githubApp)

		// TODO: Cron-type tasks
		if err := githubApp.SyncInstallations(context.Background()); err != nil {
			klog.Warningf("error syncing github installations: %v", err)
		}
	}

	clientID := os.Getenv("OAUTH2_CLIENT_ID")
	if clientID != "" {
		clientSecret := os.Getenv("OAUTH2_CLIENT_SECRET")

		isGoogle := false
		if isGoogle {
			googleProvider, err := loginwithgoogle.NewGoogleProvider("google", clientID, clientSecret)
			if err != nil {
				return nil, fmt.Errorf("error building google provider: %w", err)
			}

			// TODO: Clean this up ... we should have a shared login component (e.g. that implements logout?)
			loginComponent, err := loginwithgoogle.NewComponent(userComponent, googleProvider)
			if err != nil {
				return nil, fmt.Errorf("error building login component: %w", err)
			}
			s.Components = append(s.Components, loginComponent)
		} else {
			githubAuth, err := loginwithgithub.NewGithubProvider(clientID, clientSecret)
			if err != nil {
				return nil, fmt.Errorf("error building github auth provider: %w", err)
			}

			// TODO: Clean this up ... we should have a shared login component (e.g. that implements logout?)
			loginComponent, err := loginwithgithub.NewComponent(userComponent, githubAuth)
			if err != nil {
				return nil, fmt.Errorf("error building login component: %w", err)
			}
			s.Components = append(s.Components, loginComponent)
		}
	}

	return s, nil
}

func (s *Server) ListenAndServe(ctx context.Context, listen string, listening chan<- net.Addr) error {
	defer func() {
		if listening != nil {
			close(listening)
		}
	}()

	if err := s.ensureMux(); err != nil {
		return err
	}

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

func parsePrivateKey(p string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("error reading file %q: %w", p, err)
	}

	pemBlock, rest := pem.Decode(b)
	if pemBlock == nil {
		return nil, fmt.Errorf("cannot decode file %q as pem", p)
	}

	if rest != nil {
		rest = bytes.TrimSpace(rest)
	}
	if len(rest) != 0 {
		return nil, fmt.Errorf("unexpected additional data in file %q", p)
	}

	if pemBlock.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("unexpected type for private key: %q", pemBlock.Type)
	}

	parsed, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key from %q: %w", p, err)
	}
	return parsed, nil
}
