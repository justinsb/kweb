package gcsstore

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/justinsb/kweb/services/blobstore/pkg/server"
	"k8s.io/klog/v2"
)

type GCSStore struct {
	bucket    string
	keyPrefix string

	gcsClient *storage.Client
}

var _ server.Store = &GCSStore{}

func NewGCSStore(ctx context.Context, bucket string, keyPrefix string) (*GCSStore, error) {
	log := klog.FromContext(ctx)

	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating Google Cloud Storage client: %w", err)
	}

	store := &GCSStore{
		gcsClient: gcsClient,
		bucket:    bucket,
		keyPrefix: keyPrefix,
	}

	// b, _ := os.ReadFile("/var/run/secrets/cloud.google.com/token")
	// log.Info("read token", "token", string(b))

	// Find out and print our current token identity
	// {
	// 	var opts []option.ClientOption

	// 	// Prepend default options to avoid overriding options passed by the user.
	// 	opts = append(opts, option.WithScopes(storage.ScopeFullControl, "https://www.googleapis.com/auth/cloud-platform"))

	// 	creds, err := transport.Creds(ctx, opts...)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("getting google credentials: %w", err)
	// 	}
	// 	log.Info("got creds", "universe", creds.UniverseDomain())

	// 	token, err := creds.TokenSource.Token()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("getting token: %w", err)
	// 	}
	// 	u := "https://www.googleapis.com/oauth2/v1/tokeninfo?id_token=" + token.AccessToken
	// 	resp, err := http.Get(u)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("getting tokeninfo: %w", err)
	// 	}
	// 	defer resp.Body.Close()
	// 	b, err := io.ReadAll(resp.Body)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("reading tokeninfo response: %w", err)
	// 	}
	// 	log.Info("got tokeninfo", "info", string(b))
	// }

	// Issue a bucket-location query, at least to check that everything is ok
	attrs, err := store.getBucketAttributes(ctx)
	if err != nil {
		return nil, err
	}
	log.Info("verified blob storage bucket", "bucket", store.bucket, "location", attrs.Location)

	return store, nil
}

func (s *GCSStore) getBucketAttributes(ctx context.Context) (*storage.BucketAttrs, error) {
	// log := klog.FromContext(ctx)

	gcsBucket := s.gcsClient.Bucket(s.bucket)
	gcsURL := "gs://" + s.bucket

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	attrs, err := gcsBucket.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting bucket attributes for %s: %w", gcsURL, err)
	}

	return attrs, nil
}

func (s *GCSStore) CreateBlob(ctx context.Context, sha256 []byte, content io.Reader) (*server.BlobInfo, error) {
	log := klog.FromContext(ctx)

	key := s.keyPrefix + hex.EncodeToString(sha256)
	gcsURL := "gs://" + s.bucket + "/" + key

	gcsBucket := s.gcsClient.Bucket(s.bucket)
	gcsObject := gcsBucket.Object(key)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	w := gcsObject.NewWriter(ctx)
	if _, err := io.Copy(w, content); err != nil {
		return nil, fmt.Errorf("writing to GCS %s: %w", gcsURL, err)
	}

	// Close, just like writing a file.
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("finalizing GCS file %s: %w", gcsURL, err)
	}

	log.Info("uploaded file to gcs", "path", gcsURL)

	blobInfo := &server.BlobInfo{
		//Length: n,
	}
	return blobInfo, nil
}

func (s *GCSStore) Open(ctx context.Context, sha256 []byte) (io.ReadCloser, error) {
	// log := klog.FromContext(ctx)

	key := s.keyPrefix + hex.EncodeToString(sha256)
	gcsURL := "gs://" + s.bucket + "/" + key

	// TODO: Cache objects?

	gcsBucket := s.gcsClient.Bucket(s.bucket)
	gcsObject := gcsBucket.Object(key)

	r, err := gcsObject.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading object %s: %w", gcsURL, err)
	}

	return r, nil
}
