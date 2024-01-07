package server

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"

	pb "github.com/justinsb/kweb/services/blobstore/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

func NewBlobStoreService(store *FilesystemStore) *blobStoreService {
	return &blobStoreService{store: store}
}

// blobStoreService implements the Blob Store GRPC Service.
type blobStoreService struct {
	pb.UnimplementedBlobStoreServer

	store *FilesystemStore
}

func (s *blobStoreService) GetBlob(req *pb.GetBlobRequest, stream pb.BlobStore_GetBlobServer) error {
	// TODO: Verify req.GetStore()

	ctx := stream.Context()

	hash := req.GetSha256()
	r, err := s.store.Open(ctx, hash)
	if err != nil {
		return err
	}
	defer r.Close()

	// TODO: Pool buffers? (But this isn't the most efficient transmission approach anyway)
	buf := make([]byte, 65536)
	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		resp := &pb.GetBlobReply{
			Data: buf[:n],
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
		if err == io.EOF {
			break
		}
	}

	return nil
}

func (s *blobStoreService) CreateBlob(stream pb.BlobStore_CreateBlobServer) error {
	ctx := stream.Context()
	log := klog.FromContext(ctx)

	tmpFile, err := os.CreateTemp("", "createblob-*")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Error(err, "removing temp file", "path", tmpFile.Name())
		}
	}()

	hasher := sha256.New()
	mw := io.MultiWriter(tmpFile, hasher)

	packetIndex := -1
	var declaredHash []byte
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		packetIndex++

		if packetIndex == 0 {
			// TODO: Verify req.GetStore()
		}

		if _, err := mw.Write(req.GetData()); err != nil {
			return err
		}

		if req.Done {
			declaredHash = req.GetSha256()
			break
		}
	}

	actualHash := hasher.Sum(nil)

	klog.Infof("got blob; actual hash is %x vs declared hash %x", actualHash, declaredHash)
	if !bytes.Equal(actualHash, declaredHash) {
		return status.Errorf(codes.InvalidArgument, "hash did not match uploaded data")
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		return err
	}

	blobInfo, err := s.store.CreateBlob(ctx, actualHash, tmpFile)
	if err != nil {
		return err
	}

	// TODO: Check / return blobInfo?
	log.Info("uploaded blob", "blobInfo", blobInfo)

	reply := &pb.CreateBlobReply{}
	if err := stream.SendAndClose(reply); err != nil {
		return err
	}

	return nil
}
