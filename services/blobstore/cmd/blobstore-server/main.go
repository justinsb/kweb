package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	api "github.com/justinsb/kweb/services/blobstore/api/v1"
	"github.com/justinsb/kweb/services/blobstore/pkg/server"
	"github.com/justinsb/kweb/services/blobstore/pkg/server/gcsstore"
	"github.com/justinsb/kweb/services/blobstore/pkg/server/s3store"
	kinspire "github.com/justinsb/packages/kinspire/client"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/klog/v2"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	log := klog.FromContext(ctx)

	listen := "0.0.0.0:8443"
	storePath := os.Getenv("BLOBSTORE_SERVER_STORE")
	flag.StringVar(&listen, "listen", listen, "endpoint on which to listen")
	flag.StringVar(&storePath, "store", storePath, "storage location")
	klog.InitFlags(nil)
	flag.Parse()

	storePathURL, err := url.Parse(storePath)
	if err != nil {
		return fmt.Errorf("cannot parse flag --store=%q: %w", storePath, err)
	}

	if err := kinspire.SPIFFE.Init(ctx); err != nil {
		return fmt.Errorf("error initializing SPIFFE: %w", err)
	}

	source := kinspire.SPIFFE.Source()
	svid, err := source.GetX509SVID()
	if err != nil {
		return err
	}
	log.Info("got x509 from kinspire", "svid", svid.ID)

	// Create a `tls.Config` to allow mTLS connections, and verify that presented certificate has the specified SPIFFE ID
	tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeAny())

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listen, err)
	}

	var opts []grpc.ServerOption

	opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))

	grpcServer := grpc.NewServer(opts...)

	var store server.Store
	switch storePathURL.Scheme {
	case "file":
		s := server.NewFilesystemStore(storePath)
		store = s
	case "s3":
		bucket := storePathURL.Host
		keyPrefix := storePathURL.Path
		if !strings.HasSuffix(keyPrefix, "/") {
			keyPrefix += "/"
		}
		keyPrefix = strings.TrimPrefix(keyPrefix, "/")
		s, err := s3store.NewS3Store(ctx, bucket, keyPrefix)
		if err != nil {
			return err
		}
		store = s
	case "gs":
		bucket := storePathURL.Host
		keyPrefix := storePathURL.Path
		if !strings.HasSuffix(keyPrefix, "/") {
			keyPrefix += "/"
		}
		keyPrefix = strings.TrimPrefix(keyPrefix, "/")
		s, err := gcsstore.NewGCSStore(ctx, bucket, keyPrefix)
		if err != nil {
			return err
		}
		store = s
	default:
		return fmt.Errorf("unknown store type %q in --store=%q", storePathURL.Scheme, storePath)
	}
	blobStoreServer := server.NewBlobStoreService(store)

	api.RegisterBlobStoreServer(grpcServer, blobStoreServer)
	log.Info("serving blobstore service", "listen", listen)
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}
