// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: apps/sso/pb/app.proto

package pb

import (
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

type AppSessionData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Redirect string `protobuf:"bytes,1,opt,name=redirect,proto3" json:"redirect,omitempty"`
}

func (x *AppSessionData) Reset() {
	*x = AppSessionData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apps_sso_pb_app_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AppSessionData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppSessionData) ProtoMessage() {}

func (x *AppSessionData) ProtoReflect() protoreflect.Message {
	mi := &file_apps_sso_pb_app_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppSessionData.ProtoReflect.Descriptor instead.
func (*AppSessionData) Descriptor() ([]byte, []int) {
	return file_apps_sso_pb_app_proto_rawDescGZIP(), []int{0}
}

func (x *AppSessionData) GetRedirect() string {
	if x != nil {
		return x.Redirect
	}
	return ""
}

var File_apps_sso_pb_app_proto protoreflect.FileDescriptor

var file_apps_sso_pb_app_proto_rawDesc = []byte{
	0x0a, 0x15, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x73, 0x73, 0x6f, 0x2f, 0x70, 0x62, 0x2f, 0x61, 0x70,
	0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x22, 0x2c, 0x0a, 0x0e, 0x41,
	0x70, 0x70, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1a, 0x0a,
	0x08, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x42, 0x60, 0x0a, 0x06, 0x63, 0x6f, 0x6d,
	0x2e, 0x70, 0x62, 0x42, 0x08, 0x41, 0x70, 0x70, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x75, 0x73, 0x74,
	0x69, 0x6e, 0x73, 0x62, 0x2f, 0x6b, 0x77, 0x65, 0x62, 0x2f, 0x61, 0x70, 0x70, 0x73, 0x2f, 0x73,
	0x73, 0x6f, 0x2f, 0x70, 0x62, 0xa2, 0x02, 0x03, 0x50, 0x58, 0x58, 0xaa, 0x02, 0x02, 0x50, 0x62,
	0xca, 0x02, 0x02, 0x50, 0x62, 0xe2, 0x02, 0x0e, 0x50, 0x62, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x02, 0x50, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_apps_sso_pb_app_proto_rawDescOnce sync.Once
	file_apps_sso_pb_app_proto_rawDescData = file_apps_sso_pb_app_proto_rawDesc
)

func file_apps_sso_pb_app_proto_rawDescGZIP() []byte {
	file_apps_sso_pb_app_proto_rawDescOnce.Do(func() {
		file_apps_sso_pb_app_proto_rawDescData = protoimpl.X.CompressGZIP(file_apps_sso_pb_app_proto_rawDescData)
	})
	return file_apps_sso_pb_app_proto_rawDescData
}

var file_apps_sso_pb_app_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_apps_sso_pb_app_proto_goTypes = []interface{}{
	(*AppSessionData)(nil), // 0: pb.AppSessionData
}
var file_apps_sso_pb_app_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_apps_sso_pb_app_proto_init() }
func file_apps_sso_pb_app_proto_init() {
	if File_apps_sso_pb_app_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_apps_sso_pb_app_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AppSessionData); i {
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
			RawDescriptor: file_apps_sso_pb_app_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_apps_sso_pb_app_proto_goTypes,
		DependencyIndexes: file_apps_sso_pb_app_proto_depIdxs,
		MessageInfos:      file_apps_sso_pb_app_proto_msgTypes,
	}.Build()
	File_apps_sso_pb_app_proto = out.File
	file_apps_sso_pb_app_proto_rawDesc = nil
	file_apps_sso_pb_app_proto_goTypes = nil
	file_apps_sso_pb_app_proto_depIdxs = nil
}
