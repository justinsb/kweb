package blobstoreclient

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"io"
	"strings"

	pb "github.com/justinsb/kweb/services/blobstore/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/klog/v2"
)

type BlobStoreClient struct {
	client pb.BlobStoreClient
}

type BlobInfo struct {
	SHA256 []byte
	Length int64
}

type Client interface {
	Upload(ctx context.Context, r io.Reader) (*BlobInfo, error)
}

type ClientOptions struct {
	Host      string
	TLSConfig *tls.Config
}

func NewClient(ctx context.Context, options ClientOptions) (Client, error) {
	log := klog.FromContext(ctx)
	connectTo := options.Host
	if connectTo == "" {
		return nil, fmt.Errorf("must specify host")
	}

	var opts []grpc.DialOption

	// opts = append(opts, grpc.WithBlock())
	if options.TLSConfig != nil {
		// cert, err := tls.LoadX509KeyPair(options.ClientCertPath, options.ClientKeyPath)
		// if err != nil {
		// 	log.Fatalf("failed to load client cert: %v", err)
		// }

		// ca := x509.NewCertPool()
		// caFilePath := options.ServerCAPath
		// caBytes, err := os.ReadFile(caFilePath)
		// if err != nil {
		// 	return nil, fmt.Errorf("reading %q: %w", caFilePath, err)
		// }
		// if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		// 	return nil, fmt.Errorf("unable to parse certificates from %q", caFilePath)
		// }

		// tlsConfig := &tls.Config{
		// 	ServerName:   options.Name,
		// 	Certificates: []tls.Certificate{cert},
		// 	RootCAs:      ca,
		// }
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(options.TLSConfig)))

		// opts = append(opts, grpc.WithContextDialer(func(ctx context.Context, name string) (net.Conn, error) {
		// 	dialer := &net.Dialer{}
		// 	return dialer.DialContext(ctx, "tcp", options.Host+":8443")
		// }))
		if !strings.Contains(connectTo, ":") {
			connectTo += ":443"
		}
	} else {
		klog.Warningf("using insecure connection")
		if !strings.Contains(connectTo, ":") {
			connectTo += ":80"
		}
		opts = append(opts, grpc.WithInsecure())
	}

	log.Info("connecting to blobstore server", "host", connectTo)
	conn, err := grpc.Dial(connectTo, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to service on %s: %w", connectTo, err)
	}

	blobStoreClient := pb.NewBlobStoreClient(conn)
	return &BlobStoreClient{client: blobStoreClient}, nil
}

func (c *BlobStoreClient) Upload(ctx context.Context, r io.Reader) (*BlobInfo, error) {
	log := klog.FromContext(ctx)

	stream, err := c.client.CreateBlob(ctx)
	if err != nil {
		return nil, fmt.Errorf("error from CreateBlob call: %w", err)
	}
	defer func() {
		err := stream.CloseSend()
		if err != nil {
			log.Error(err, "closing stream")
		}
	}()

	store := "todo"

	hasher := sha256.New()

	buf := make([]byte, 65536)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			req := &pb.CreateBlobRequest{
				Store: store,
				Data:  buf[:n],
			}

			if err := stream.Send(req); err != nil {
				return nil, fmt.Errorf("sending CreateBlob message: %w", err)
			}

			if _, err := hasher.Write(req.Data); err != nil {
				return nil, fmt.Errorf("hashing data: %w", err)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("reading data: %w", err)
		}
	}

	sha256 := hasher.Sum(nil)
	{
		req := &pb.CreateBlobRequest{
			Store:  store,
			Sha256: sha256,
			Done:   true,
		}

		log.Info("sending req", "req", req)
		if err := stream.Send(req); err != nil {
			return nil, fmt.Errorf("sending CreateBlob message: %w", err)
		}
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("closing stream: %w", err)
	}

	blobInfo := &BlobInfo{
		SHA256: sha256,
		Length: response.Length,
	}
	return blobInfo, nil
}
