package blobstoreclient

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	pb "github.com/justinsb/kweb/services/blobstore/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/klog/v2"
)

type BlobStoreClient struct {
	client pb.BlobStoreClient
}

type BlobFile struct {
	path string
}

func (f *BlobFile) Close() error {
	// TODO: Maybe cache these locally?
	if err := os.Remove(f.path); err != nil {
		return fmt.Errorf("removing temp file %q: %w", f.path, err)
	}
	return nil
}

func (f *BlobFile) Path() string {
	return f.path
}

func (f *BlobFile) Open() io.ReadCloser {
	return &blobFileReader{blobFile: f}
}

type blobFileReader struct {
	f        *os.File
	blobFile *BlobFile
}

var _ io.ReadCloser = &blobFileReader{}

func (r *blobFileReader) Close() error {
	var errs []error
	if err := r.f.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.blobFile.Close(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (r *blobFileReader) Read(p []byte) (int, error) {
	if r.f == nil {
		f, err := os.Open(r.blobFile.path)
		if err != nil {
			return 0, err
		}
		r.f = f
	}

	return r.f.Read(p)
}

type BlobInfo struct {
	SHA256 []byte
	Length int64
}

type Client interface {
	Upload(ctx context.Context, r io.Reader) (*BlobInfo, error)
	Download(ctx context.Context, sha256 []byte) (*BlobFile, error)
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

func (c *BlobStoreClient) Download(ctx context.Context, blobHash []byte) (*BlobFile, error) {
	log := klog.FromContext(ctx)

	store := "todo"

	req := &pb.GetBlobRequest{
		Store:  store,
		Sha256: blobHash,
	}
	stream, err := c.client.GetBlob(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error from GetBlob call: %w", err)
	}
	defer func() {
		err := stream.CloseSend()
		if err != nil {
			log.Error(err, "closing stream")
		}
	}()

	hasher := sha256.New()

	tmpFile, err := WriteTempFile(ctx, func(w io.Writer) error {
		mw := io.MultiWriter(w, hasher)
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("reading stream: %w", err)
			}

			if len(msg.Data) != 0 {
				_, err := mw.Write(msg.Data)
				if err != nil {
					return fmt.Errorf("writing to temp file: %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	blobFile := &BlobFile{path: tmpFile}

	actualHash := hasher.Sum(nil)
	if !bytes.Equal(actualHash, blobHash) {
		blobFile.Close()
		return nil, fmt.Errorf("downloaded blob hash did not match expected hash")
	}

	return blobFile, nil
}

func WriteTempFile(ctx context.Context, writer func(w io.Writer) error) (string, error) {
	log := klog.FromContext(ctx)

	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}

	tempFilePath := tempFile.Name()

	shouldDelete := true
	defer func() {
		if shouldDelete {
			if err := os.Remove(tempFilePath); err != nil {
				log.Error(err, "removing temp file", "path", tempFilePath)
			}
		}
	}()

	shouldClose := true
	defer func() {
		if shouldClose {
			if err := tempFile.Close(); err != nil {
				log.Error(err, "closing temp file", "path", tempFilePath)
			}
		}
	}()

	if err := writer(tempFile); err != nil {
		return "", fmt.Errorf("writing to temp file %q: %w", tempFilePath, err)
	}

	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("closing temp file %q: %w", tempFilePath, err)
	}
	shouldClose = false

	shouldDelete = false

	return tempFilePath, nil
}
