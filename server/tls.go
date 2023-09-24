package server

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"k8s.io/klog/v2"
)

func (s *Server) buildTLSConfig(ctx context.Context) (*tls.Config, error) {
	// Worload API socket path
	const socketPath = "unix:///run/spire/sockets/agent.sock"

	klog.Infof("creating x509 source with %q", socketPath)
	// Create a `workloadapi.X509Source`, it will connect to Workload API using provided socket.
	// If socket path is not defined using `workloadapi.SourceOption`, value from environment variable `SPIFFE_ENDPOINT_SOCKET` is used.
	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(socketPath)))
	if err != nil {
		return nil, fmt.Errorf("unable to create X509Source: %w", err)
	}
	// defer source.Close()

	svid, err := source.GetX509SVID()
	if err != nil {
		return nil, err
	}
	klog.Infof("my x509 is %v", svid)

	// Allowed SPIFFE ID
	clientID := spiffeid.RequireFromString("spiffe://example.org/ns/default/sa/gateway-instance")

	klog.Infof("creating httpserver, requires %v", clientID)
	// Create a `tls.Config` to allow mTLS connections, and verify that presented certificate has the specified SPIFFE ID
	tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeID(clientID))

	return tlsConfig, nil
}
