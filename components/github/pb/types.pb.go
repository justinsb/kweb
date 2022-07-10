// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.12.4
// source: components/github/pb/types.proto

package pb

import (
	kube "github.com/justinsb/kweb/components/kube"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type AppInstallation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Typemeta *kube.TypeMeta       `protobuf:"bytes,1,opt,name=typemeta,proto3" json:"typemeta,omitempty"`
	Metadata *kube.ObjectMeta     `protobuf:"bytes,2,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Spec     *AppInstallationSpec `protobuf:"bytes,3,opt,name=spec,proto3" json:"spec,omitempty"`
}

func (x *AppInstallation) Reset() {
	*x = AppInstallation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_components_github_pb_types_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AppInstallation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppInstallation) ProtoMessage() {}

func (x *AppInstallation) ProtoReflect() protoreflect.Message {
	mi := &file_components_github_pb_types_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppInstallation.ProtoReflect.Descriptor instead.
func (*AppInstallation) Descriptor() ([]byte, []int) {
	return file_components_github_pb_types_proto_rawDescGZIP(), []int{0}
}

func (x *AppInstallation) GetTypemeta() *kube.TypeMeta {
	if x != nil {
		return x.Typemeta
	}
	return nil
}

func (x *AppInstallation) GetMetadata() *kube.ObjectMeta {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *AppInstallation) GetSpec() *AppInstallationSpec {
	if x != nil {
		return x.Spec
	}
	return nil
}

type AppInstallationSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The capitilization isn't normal proto, but it avoids name mangling
	Id      int64          `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	UserID  string         `protobuf:"bytes,2,opt,name=userID,proto3" json:"userID,omitempty"`
	Account *GithubAccount `protobuf:"bytes,3,opt,name=account,proto3" json:"account,omitempty"`
}

func (x *AppInstallationSpec) Reset() {
	*x = AppInstallationSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_components_github_pb_types_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AppInstallationSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppInstallationSpec) ProtoMessage() {}

func (x *AppInstallationSpec) ProtoReflect() protoreflect.Message {
	mi := &file_components_github_pb_types_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppInstallationSpec.ProtoReflect.Descriptor instead.
func (*AppInstallationSpec) Descriptor() ([]byte, []int) {
	return file_components_github_pb_types_proto_rawDescGZIP(), []int{1}
}

func (x *AppInstallationSpec) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *AppInstallationSpec) GetUserID() string {
	if x != nil {
		return x.UserID
	}
	return ""
}

func (x *AppInstallationSpec) GetAccount() *GithubAccount {
	if x != nil {
		return x.Account
	}
	return nil
}

type GithubAccount struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Login string `protobuf:"bytes,2,opt,name=login,proto3" json:"login,omitempty"`
}

func (x *GithubAccount) Reset() {
	*x = GithubAccount{}
	if protoimpl.UnsafeEnabled {
		mi := &file_components_github_pb_types_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GithubAccount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GithubAccount) ProtoMessage() {}

func (x *GithubAccount) ProtoReflect() protoreflect.Message {
	mi := &file_components_github_pb_types_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GithubAccount.ProtoReflect.Descriptor instead.
func (*GithubAccount) Descriptor() ([]byte, []int) {
	return file_components_github_pb_types_proto_rawDescGZIP(), []int{2}
}

func (x *GithubAccount) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *GithubAccount) GetLogin() string {
	if x != nil {
		return x.Login
	}
	return ""
}

var File_components_github_pb_types_proto protoreflect.FileDescriptor

var file_components_github_pb_types_proto_rawDesc = []byte{
	0x0a, 0x20, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2f, 0x70, 0x62, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x1a, 0x0a, 0x6b, 0x75, 0x62, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xaf, 0x01, 0x0a, 0x0f, 0x41, 0x70, 0x70, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c,
	0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2a, 0x0a, 0x08, 0x74, 0x79, 0x70, 0x65, 0x6d, 0x65,
	0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x2e,
	0x54, 0x79, 0x70, 0x65, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x08, 0x74, 0x79, 0x70, 0x65, 0x6d, 0x65,
	0x74, 0x61, 0x12, 0x2c, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x2e, 0x4f, 0x62, 0x6a, 0x65,
	0x63, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x12, 0x2b, 0x0a, 0x04, 0x73, 0x70, 0x65, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17,
	0x2e, 0x70, 0x62, 0x2e, 0x41, 0x70, 0x70, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x53, 0x70, 0x65, 0x63, 0x52, 0x04, 0x73, 0x70, 0x65, 0x63, 0x3a, 0x15, 0x8a,
	0xb5, 0x18, 0x11, 0x0a, 0x0f, 0x41, 0x70, 0x70, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0x6a, 0x0a, 0x13, 0x41, 0x70, 0x70, 0x49, 0x6e, 0x73, 0x74, 0x61,
	0x6c, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x70, 0x65, 0x63, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x75,
	0x73, 0x65, 0x72, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73, 0x65,
	0x72, 0x49, 0x44, 0x12, 0x2b, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x70, 0x62, 0x2e, 0x47, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x22, 0x35, 0x0a, 0x0d, 0x47, 0x69, 0x74, 0x68, 0x75, 0x62, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x42, 0x4d, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x75, 0x73, 0x74, 0x69, 0x6e, 0x73, 0x62, 0x2f, 0x6b,
	0x77, 0x65, 0x62, 0x2f, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x67,
	0x68, 0x61, 0x70, 0x70, 0x2f, 0x70, 0x62, 0x82, 0xb5, 0x18, 0x1b, 0x0a, 0x0f, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x6b, 0x77, 0x65, 0x62, 0x2e, 0x64, 0x65, 0x76, 0x12, 0x08, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_components_github_pb_types_proto_rawDescOnce sync.Once
	file_components_github_pb_types_proto_rawDescData = file_components_github_pb_types_proto_rawDesc
)

func file_components_github_pb_types_proto_rawDescGZIP() []byte {
	file_components_github_pb_types_proto_rawDescOnce.Do(func() {
		file_components_github_pb_types_proto_rawDescData = protoimpl.X.CompressGZIP(file_components_github_pb_types_proto_rawDescData)
	})
	return file_components_github_pb_types_proto_rawDescData
}

var file_components_github_pb_types_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_components_github_pb_types_proto_goTypes = []interface{}{
	(*AppInstallation)(nil),     // 0: pb.AppInstallation
	(*AppInstallationSpec)(nil), // 1: pb.AppInstallationSpec
	(*GithubAccount)(nil),       // 2: pb.GithubAccount
	(*kube.TypeMeta)(nil),       // 3: kube.TypeMeta
	(*kube.ObjectMeta)(nil),     // 4: kube.ObjectMeta
}
var file_components_github_pb_types_proto_depIdxs = []int32{
	3, // 0: pb.AppInstallation.typemeta:type_name -> kube.TypeMeta
	4, // 1: pb.AppInstallation.metadata:type_name -> kube.ObjectMeta
	1, // 2: pb.AppInstallation.spec:type_name -> pb.AppInstallationSpec
	2, // 3: pb.AppInstallationSpec.account:type_name -> pb.GithubAccount
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_components_github_pb_types_proto_init() }
func file_components_github_pb_types_proto_init() {
	if File_components_github_pb_types_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_components_github_pb_types_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AppInstallation); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_components_github_pb_types_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AppInstallationSpec); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_components_github_pb_types_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GithubAccount); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_components_github_pb_types_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_components_github_pb_types_proto_goTypes,
		DependencyIndexes: file_components_github_pb_types_proto_depIdxs,
		MessageInfos:      file_components_github_pb_types_proto_msgTypes,
	}.Build()
	File_components_github_pb_types_proto = out.File
	file_components_github_pb_types_proto_rawDesc = nil
	file_components_github_pb_types_proto_goTypes = nil
	file_components_github_pb_types_proto_depIdxs = nil
}
