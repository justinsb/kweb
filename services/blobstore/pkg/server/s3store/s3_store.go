package s3store

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/justinsb/kweb/services/blobstore/pkg/server"
	"github.com/justinsb/packages/kinspire/client"
	"k8s.io/klog/v2"
)

type S3Store struct {
	bucket    string
	keyPrefix string

	s3Client *s3.Client
}

var _ server.Store = &S3Store{}

func NewS3Store(ctx context.Context, bucket string, keyPrefix string) (*S3Store, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	credentials, err := client.GetAWSCredentials()
	if err != nil {
		return nil, fmt.Errorf("getting AWS credentials: %w", err)
	}
	awsConfig.Credentials = credentials
	awsConfig.Region = "us-east-2" // TODO: Should come from config?
	s3Client := s3.NewFromConfig(awsConfig)

	// Issue a bucket-location query, at least to check that everything is ok
	bucketLocation, err := s3Client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: aws.String(bucket)})
	if err != nil {
		return nil, fmt.Errorf("getting location for S3 bucket %q: %w", bucket, err)
	}
	// TODO: Use direct-to-location client?
	klog.Infof("bucket location is %q", bucketLocation.LocationConstraint)

	return &S3Store{
		s3Client:  s3Client,
		bucket:    bucket,
		keyPrefix: keyPrefix,
	}, nil
}

func (s *S3Store) CreateBlob(ctx context.Context, sha256 []byte, content io.Reader) (*server.BlobInfo, error) {
	log := klog.FromContext(ctx)

	// Create an uploader with the session and default options
	uploader := manager.NewUploader(s.s3Client)

	key := s.keyPrefix + hex.EncodeToString(sha256)

	// Upload the file to S3.
	result, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   content,
		// TODO: Does the multipart uploader actually verify checksums?
		// ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
		// ChecksumSHA256:    aws.String(base64.StdEncoding.EncodeToString(sha256)),
	})
	if err != nil {
		return nil, fmt.Errorf("uploading file to S3: %w", err)
	}
	log.Info("uploaded file to s3", "path", result.Location)

	blobInfo := &server.BlobInfo{
		//Length: n,
	}
	return blobInfo, nil
}

func (s *S3Store) Open(ctx context.Context, sha256 []byte) (io.ReadCloser, error) {
	log := klog.FromContext(ctx)

	key := s.keyPrefix + hex.EncodeToString(sha256)

	// TODO: Cache objects?

	log.Info("reading object from s3", "path", "s3://"+s.bucket+"/"+key)
	response, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("reading s3 object: %w", err)
	}
	return response.Body, nil
}
