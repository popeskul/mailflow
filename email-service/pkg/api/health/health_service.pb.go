// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        (unknown)
// source: api/health/health_service.proto

package health

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

// HealthStatus is the status of the service.
type HealthStatus int32

const (
	HealthStatus_UNKNOWN   HealthStatus = 0
	HealthStatus_HEALTHY   HealthStatus = 1
	HealthStatus_UNHEALTHY HealthStatus = 2
	HealthStatus_ALIVE     HealthStatus = 3
	HealthStatus_READY     HealthStatus = 4
)

// Enum value maps for HealthStatus.
var (
	HealthStatus_name = map[int32]string{
		0: "UNKNOWN",
		1: "HEALTHY",
		2: "UNHEALTHY",
		3: "ALIVE",
		4: "READY",
	}
	HealthStatus_value = map[string]int32{
		"UNKNOWN":   0,
		"HEALTHY":   1,
		"UNHEALTHY": 2,
		"ALIVE":     3,
		"READY":     4,
	}
)

func (x HealthStatus) Enum() *HealthStatus {
	p := new(HealthStatus)
	*p = x
	return p
}

func (x HealthStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (HealthStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_api_health_health_service_proto_enumTypes[0].Descriptor()
}

func (HealthStatus) Type() protoreflect.EnumType {
	return &file_api_health_health_service_proto_enumTypes[0]
}

func (x HealthStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use HealthStatus.Descriptor instead.
func (HealthStatus) EnumDescriptor() ([]byte, []int) {
	return file_api_health_health_service_proto_rawDescGZIP(), []int{0}
}

// HealthCheckRequest is the request message for health check RPCs.
type HealthCheckRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HealthCheckRequest) Reset() {
	*x = HealthCheckRequest{}
	mi := &file_api_health_health_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HealthCheckRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthCheckRequest) ProtoMessage() {}

func (x *HealthCheckRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_health_health_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthCheckRequest.ProtoReflect.Descriptor instead.
func (*HealthCheckRequest) Descriptor() ([]byte, []int) {
	return file_api_health_health_service_proto_rawDescGZIP(), []int{0}
}

// HealthCheckResponse is the response message for health check RPCs.
type HealthCheckResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// status is the health status of the service.
	Status HealthStatus `protobuf:"varint,1,opt,name=status,proto3,enum=health.HealthStatus" json:"status,omitempty"`
}

func (x *HealthCheckResponse) Reset() {
	*x = HealthCheckResponse{}
	mi := &file_api_health_health_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HealthCheckResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthCheckResponse) ProtoMessage() {}

func (x *HealthCheckResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_health_health_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthCheckResponse.ProtoReflect.Descriptor instead.
func (*HealthCheckResponse) Descriptor() ([]byte, []int) {
	return file_api_health_health_service_proto_rawDescGZIP(), []int{1}
}

func (x *HealthCheckResponse) GetStatus() HealthStatus {
	if x != nil {
		return x.Status
	}
	return HealthStatus_UNKNOWN
}

var File_api_health_health_service_proto protoreflect.FileDescriptor

var file_api_health_health_service_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2f, 0x68, 0x65, 0x61,
	0x6c, 0x74, 0x68, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x06, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x14, 0x0a, 0x12, 0x48, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x43, 0x0a,
	0x13, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x48, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x2a, 0x4d, 0x0a, 0x0c, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12,
	0x0b, 0x0a, 0x07, 0x48, 0x45, 0x41, 0x4c, 0x54, 0x48, 0x59, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09,
	0x55, 0x4e, 0x48, 0x45, 0x41, 0x4c, 0x54, 0x48, 0x59, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x41,
	0x4c, 0x49, 0x56, 0x45, 0x10, 0x03, 0x12, 0x09, 0x0a, 0x05, 0x52, 0x45, 0x41, 0x44, 0x59, 0x10,
	0x04, 0x32, 0xf6, 0x02, 0x0a, 0x0d, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x54, 0x0a, 0x05, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x12, 0x1a, 0x2e, 0x68,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63,
	0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x12, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0c, 0x12, 0x0a, 0x2f,
	0x76, 0x31, 0x2f, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12, 0x59, 0x0a, 0x08, 0x4c, 0x69, 0x76,
	0x65, 0x6e, 0x65, 0x73, 0x73, 0x12, 0x1a, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x48,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1b, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x14,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0e, 0x12, 0x0c, 0x2f, 0x76, 0x31, 0x2f, 0x6c, 0x69, 0x76, 0x65,
	0x6e, 0x65, 0x73, 0x73, 0x12, 0x5b, 0x0a, 0x09, 0x52, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x65, 0x73,
	0x73, 0x12, 0x1a, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e,
	0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65,
	0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x15, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x0f, 0x12, 0x0d, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x65, 0x73,
	0x73, 0x12, 0x57, 0x0a, 0x07, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x7a, 0x12, 0x1a, 0x2e, 0x68,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63,
	0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x13, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0d, 0x12, 0x0b, 0x2f,
	0x76, 0x31, 0x2f, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x7a, 0x42, 0x50, 0x5a, 0x4e, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x6f, 0x70, 0x65, 0x73, 0x6b, 0x75,
	0x6c, 0x2f, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2d,
	0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2f, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x2d, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x68,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x3b, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_health_health_service_proto_rawDescOnce sync.Once
	file_api_health_health_service_proto_rawDescData = file_api_health_health_service_proto_rawDesc
)

func file_api_health_health_service_proto_rawDescGZIP() []byte {
	file_api_health_health_service_proto_rawDescOnce.Do(func() {
		file_api_health_health_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_health_health_service_proto_rawDescData)
	})
	return file_api_health_health_service_proto_rawDescData
}

var file_api_health_health_service_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_api_health_health_service_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_health_health_service_proto_goTypes = []any{
	(HealthStatus)(0),           // 0: health.HealthStatus
	(*HealthCheckRequest)(nil),  // 1: health.HealthCheckRequest
	(*HealthCheckResponse)(nil), // 2: health.HealthCheckResponse
}
var file_api_health_health_service_proto_depIdxs = []int32{
	0, // 0: health.HealthCheckResponse.status:type_name -> health.HealthStatus
	1, // 1: health.HealthService.Check:input_type -> health.HealthCheckRequest
	1, // 2: health.HealthService.Liveness:input_type -> health.HealthCheckRequest
	1, // 3: health.HealthService.Readiness:input_type -> health.HealthCheckRequest
	1, // 4: health.HealthService.Healthz:input_type -> health.HealthCheckRequest
	2, // 5: health.HealthService.Check:output_type -> health.HealthCheckResponse
	2, // 6: health.HealthService.Liveness:output_type -> health.HealthCheckResponse
	2, // 7: health.HealthService.Readiness:output_type -> health.HealthCheckResponse
	2, // 8: health.HealthService.Healthz:output_type -> health.HealthCheckResponse
	5, // [5:9] is the sub-list for method output_type
	1, // [1:5] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_health_health_service_proto_init() }
func file_api_health_health_service_proto_init() {
	if File_api_health_health_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_health_health_service_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_health_health_service_proto_goTypes,
		DependencyIndexes: file_api_health_health_service_proto_depIdxs,
		EnumInfos:         file_api_health_health_service_proto_enumTypes,
		MessageInfos:      file_api_health_health_service_proto_msgTypes,
	}.Build()
	File_api_health_health_service_proto = out.File
	file_api_health_health_service_proto_rawDesc = nil
	file_api_health_health_service_proto_goTypes = nil
	file_api_health_health_service_proto_depIdxs = nil
}
