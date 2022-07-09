package kube

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// TODO: Can we generate these?

func SetMetadata(obj Object, metadata *ObjectMeta) {
	// TODO: Cache
	metadataField := obj.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name("metadata"))
	if metadataField == nil {
		klog.Fatalf("metadata field not found")
	}
	if metadataField.Kind() != protoreflect.MessageKind {
		klog.Fatalf("metadata field was of unexpected type %v", metadataField.Kind())
	}
	obj.ProtoReflect().Set(metadataField, protoreflect.ValueOfMessage(metadata.ProtoReflect()))
}

func InitObject(obj Object, id types.NamespacedName) {
	metadata := obj.GetMetadata()
	if metadata == nil {
		metadata = &ObjectMeta{}
	}
	metadata.Name = id.Name
	metadata.Namespace = id.Namespace
	SetMetadata(obj, metadata)
}
