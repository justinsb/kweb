package login

import (
	"encoding/base64"

	"google.golang.org/protobuf/proto"
	"k8s.io/klog/v2"

	"github.com/justinsb/kweb/components/login/pb"
)

func encodeState(data *pb.StateData) string {
	b, err := proto.Marshal(data)
	if err != nil {
		klog.Fatalf("error serializing data: %v", err)
	}

	return base64.URLEncoding.EncodeToString(b)
}
