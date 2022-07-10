package kube

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

type KindInfo struct {
	Group    string
	Version  string
	Kind     string
	Resource string
}

func (k *KindInfo) APIVersion() string {
	if k.Group == "" {
		return k.Version
	} else {
		return k.Group + "/" + k.Version
	}
}
func (k *KindInfo) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    k.Group,
		Version:  k.Version,
		Resource: k.Resource,
	}
}

func (k *KindInfo) GroupResource() schema.GroupResource {
	return schema.GroupResource{
		Group:    k.Group,
		Resource: k.Resource,
	}
}

func GetKindInfo(msg proto.Message) *KindInfo {
	// TODO: Cache
	messageDescriptor := msg.ProtoReflect().Descriptor()
	messageOptions := messageDescriptor.Options().(*descriptorpb.MessageOptions)
	kindVal := messageOptions.ProtoReflect().Get(E_Kind.TypeDescriptor())
	kind, ok := kindVal.Message().Interface().(*Kind)
	if !ok {
		klog.Fatalf("unexpected type for kind annotation, got %T", kindVal.Message().Interface())
	}

	info := &KindInfo{}
	info.Kind = kind.Kind
	if info.Kind == "" {
		info.Kind = string(messageDescriptor.Name())
	}

	info.Resource = kind.Resource
	if info.Resource == "" {
		resource := info.Kind
		resource = strings.ToLower(resource)
		resource += "s" // TODO: worry about pluralization rules?
		info.Resource = resource
	}

	fileDescriptor := messageDescriptor.ParentFile()
	fileOptions := fileDescriptor.Options().(*descriptorpb.FileOptions)
	groupVersionVal := fileOptions.ProtoReflect().Get(E_GroupVersion.TypeDescriptor())
	groupVersion, ok := groupVersionVal.Message().Interface().(*GroupVersion)
	if !ok {
		klog.Fatalf("unexpected type for group_version annotation, got %T", groupVersionVal.Message().Interface())
	}

	info.Group = groupVersion.Group
	info.Version = groupVersion.Version
	if info.Group == "" {
		// Default from path?
		klog.Fatalf("group_version not found; unable to determine group for %v", string(messageDescriptor.Name()))
	}
	if info.Version == "" {
		// Default from path?
		klog.Fatalf("group_version not found; unable to determine version for %v", string(messageDescriptor.Name()))
	}

	return info
}
