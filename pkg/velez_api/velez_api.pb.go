// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.0
// source: grpc/velez_api.proto

package velez_api

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type Version struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Version) Reset() {
	*x = Version{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_velez_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Version) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Version) ProtoMessage() {}

func (x *Version) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_velez_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Version.ProtoReflect.Descriptor instead.
func (*Version) Descriptor() ([]byte, []int) {
	return file_grpc_velez_api_proto_rawDescGZIP(), []int{0}
}

type CreateService struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CreateService) Reset() {
	*x = CreateService{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_velez_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateService) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateService) ProtoMessage() {}

func (x *CreateService) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_velez_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateService.ProtoReflect.Descriptor instead.
func (*CreateService) Descriptor() ([]byte, []int) {
	return file_grpc_velez_api_proto_rawDescGZIP(), []int{1}
}

type Version_Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Version_Request) Reset() {
	*x = Version_Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_velez_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Version_Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Version_Request) ProtoMessage() {}

func (x *Version_Request) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_velez_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Version_Request.ProtoReflect.Descriptor instead.
func (*Version_Request) Descriptor() ([]byte, []int) {
	return file_grpc_velez_api_proto_rawDescGZIP(), []int{0, 0}
}

type Version_Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version string `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *Version_Response) Reset() {
	*x = Version_Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_velez_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Version_Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Version_Response) ProtoMessage() {}

func (x *Version_Response) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_velez_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Version_Response.ProtoReflect.Descriptor instead.
func (*Version_Response) Descriptor() ([]byte, []int) {
	return file_grpc_velez_api_proto_rawDescGZIP(), []int{0, 1}
}

func (x *Version_Response) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

type CreateService_Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ImageName    string  `protobuf:"bytes,1,opt,name=imageName,proto3" json:"imageName,omitempty"`
	CpuAmount    float32 `protobuf:"fixed32,2,opt,name=cpu_amount,json=cpuAmount,proto3" json:"cpu_amount,omitempty"`
	RamMb        int32   `protobuf:"varint,4,opt,name=ram_mb,json=ramMb,proto3" json:"ram_mb,omitempty"`
	MemorySwapMb int32   `protobuf:"varint,5,opt,name=memory_swap_mb,json=memorySwapMb,proto3" json:"memory_swap_mb,omitempty"`
}

func (x *CreateService_Request) Reset() {
	*x = CreateService_Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_velez_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateService_Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateService_Request) ProtoMessage() {}

func (x *CreateService_Request) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_velez_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateService_Request.ProtoReflect.Descriptor instead.
func (*CreateService_Request) Descriptor() ([]byte, []int) {
	return file_grpc_velez_api_proto_rawDescGZIP(), []int{1, 0}
}

func (x *CreateService_Request) GetImageName() string {
	if x != nil {
		return x.ImageName
	}
	return ""
}

func (x *CreateService_Request) GetCpuAmount() float32 {
	if x != nil {
		return x.CpuAmount
	}
	return 0
}

func (x *CreateService_Request) GetRamMb() int32 {
	if x != nil {
		return x.RamMb
	}
	return 0
}

func (x *CreateService_Request) GetMemorySwapMb() int32 {
	if x != nil {
		return x.MemorySwapMb
	}
	return 0
}

type CreateService_Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uuid string `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"` // later may be extended to return ip and port of given service
}

func (x *CreateService_Response) Reset() {
	*x = CreateService_Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_velez_api_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateService_Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateService_Response) ProtoMessage() {}

func (x *CreateService_Response) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_velez_api_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateService_Response.ProtoReflect.Descriptor instead.
func (*CreateService_Response) Descriptor() ([]byte, []int) {
	return file_grpc_velez_api_proto_rawDescGZIP(), []int{1, 1}
}

func (x *CreateService_Response) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

var File_grpc_velez_api_proto protoreflect.FileDescriptor

var file_grpc_velez_api_proto_rawDesc = []byte{
	0x0a, 0x14, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x76, 0x65, 0x6c, 0x65, 0x7a, 0x5f, 0x61, 0x70, 0x69,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x76, 0x65, 0x6c, 0x65, 0x7a, 0x5f, 0x61, 0x70,
	0x69, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e,
	0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x3a, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x1a, 0x09, 0x0a, 0x07, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0xb5, 0x01, 0x0a, 0x0d,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x1a, 0x83, 0x01,
	0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x6d, 0x61,
	0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x69, 0x6d,
	0x61, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x70, 0x75, 0x5f, 0x61,
	0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x52, 0x09, 0x63, 0x70, 0x75,
	0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x72, 0x61, 0x6d, 0x5f, 0x6d, 0x62,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x72, 0x61, 0x6d, 0x4d, 0x62, 0x12, 0x24, 0x0a,
	0x0e, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x5f, 0x73, 0x77, 0x61, 0x70, 0x5f, 0x6d, 0x62, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x53, 0x77, 0x61,
	0x70, 0x4d, 0x62, 0x1a, 0x1e, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75,
	0x75, 0x69, 0x64, 0x32, 0xd5, 0x01, 0x0a, 0x08, 0x56, 0x65, 0x6c, 0x65, 0x7a, 0x41, 0x50, 0x49,
	0x12, 0x57, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1a, 0x2e, 0x76, 0x65,
	0x6c, 0x65, 0x7a, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x2e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x76, 0x65, 0x6c, 0x65, 0x7a, 0x5f,
	0x61, 0x70, 0x69, 0x2e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x13, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0d, 0x3a, 0x01, 0x2a, 0x22,
	0x08, 0x2f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x70, 0x0a, 0x0d, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x20, 0x2e, 0x76, 0x65, 0x6c,
	0x65, 0x7a, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x76,
	0x65, 0x6c, 0x65, 0x7a, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x1a, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x14, 0x3a, 0x01, 0x2a, 0x22, 0x0f, 0x2f, 0x63, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x42, 0x0c, 0x5a, 0x0a, 0x2f,
	0x76, 0x65, 0x6c, 0x65, 0x7a, 0x5f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_grpc_velez_api_proto_rawDescOnce sync.Once
	file_grpc_velez_api_proto_rawDescData = file_grpc_velez_api_proto_rawDesc
)

func file_grpc_velez_api_proto_rawDescGZIP() []byte {
	file_grpc_velez_api_proto_rawDescOnce.Do(func() {
		file_grpc_velez_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_grpc_velez_api_proto_rawDescData)
	})
	return file_grpc_velez_api_proto_rawDescData
}

var file_grpc_velez_api_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_grpc_velez_api_proto_goTypes = []interface{}{
	(*Version)(nil),                // 0: velez_api.Version
	(*CreateService)(nil),          // 1: velez_api.CreateService
	(*Version_Request)(nil),        // 2: velez_api.Version.Request
	(*Version_Response)(nil),       // 3: velez_api.Version.Response
	(*CreateService_Request)(nil),  // 4: velez_api.CreateService.Request
	(*CreateService_Response)(nil), // 5: velez_api.CreateService.Response
}
var file_grpc_velez_api_proto_depIdxs = []int32{
	2, // 0: velez_api.VelezAPI.Version:input_type -> velez_api.Version.Request
	4, // 1: velez_api.VelezAPI.CreateService:input_type -> velez_api.CreateService.Request
	3, // 2: velez_api.VelezAPI.Version:output_type -> velez_api.Version.Response
	5, // 3: velez_api.VelezAPI.CreateService:output_type -> velez_api.CreateService.Response
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_grpc_velez_api_proto_init() }
func file_grpc_velez_api_proto_init() {
	if File_grpc_velez_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_grpc_velez_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Version); i {
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
		file_grpc_velez_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateService); i {
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
		file_grpc_velez_api_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Version_Request); i {
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
		file_grpc_velez_api_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Version_Response); i {
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
		file_grpc_velez_api_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateService_Request); i {
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
		file_grpc_velez_api_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateService_Response); i {
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
			RawDescriptor: file_grpc_velez_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_grpc_velez_api_proto_goTypes,
		DependencyIndexes: file_grpc_velez_api_proto_depIdxs,
		MessageInfos:      file_grpc_velez_api_proto_msgTypes,
	}.Build()
	File_grpc_velez_api_proto = out.File
	file_grpc_velez_api_proto_rawDesc = nil
	file_grpc_velez_api_proto_goTypes = nil
	file_grpc_velez_api_proto_depIdxs = nil
}
