package kube

import "google.golang.org/protobuf/proto"

type Object interface {
	proto.Message
	GetMetadata() *ObjectMeta
}
