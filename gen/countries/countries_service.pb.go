// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: countries/countries_service.proto

package countries

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_countries_countries_service_proto protoreflect.FileDescriptor

var file_countries_countries_service_proto_rawDesc = []byte{
	0x0a, 0x21, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x72, 0x69, 0x65, 0x73, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x09, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x1a, 0x19,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72,
	0x69, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xe9, 0x09, 0x0a, 0x07, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3f, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12,
	0x18, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x12, 0x18, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x36, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x15,
	0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65,
	0x73, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x44, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x19, 0x2e, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65,
	0x73, 0x2e, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x3f, 0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12,
	0x18, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x45, 0x0a, 0x08, 0x53, 0x65, 0x74, 0x46, 0x69, 0x61,
	0x74, 0x73, 0x12, 0x1a, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x53,
	0x65, 0x74, 0x46, 0x69, 0x61, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b,
	0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x53, 0x65, 0x74, 0x46, 0x69,
	0x61, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x45, 0x0a,
	0x08, 0x41, 0x64, 0x64, 0x46, 0x69, 0x61, 0x74, 0x73, 0x12, 0x1a, 0x2e, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x46, 0x69, 0x61, 0x74, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65,
	0x73, 0x2e, 0x41, 0x64, 0x64, 0x46, 0x69, 0x61, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x45, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x46, 0x69, 0x61, 0x74, 0x73,
	0x12, 0x1a, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x47, 0x65, 0x74,
	0x46, 0x69, 0x61, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x46, 0x69, 0x61, 0x74,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4e, 0x0a, 0x0b, 0x52,
	0x65, 0x6d, 0x6f, 0x76, 0x65, 0x46, 0x69, 0x61, 0x74, 0x73, 0x12, 0x1d, 0x2e, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x46, 0x69, 0x61,
	0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x46, 0x69, 0x61, 0x74,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4b, 0x0a, 0x0a, 0x53,
	0x65, 0x74, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x12, 0x1c, 0x2e, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x53, 0x65, 0x74, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72,
	0x69, 0x65, 0x73, 0x2e, 0x53, 0x65, 0x74, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4b, 0x0a, 0x0a, 0x41, 0x64, 0x64, 0x43,
	0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x12, 0x1c, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69,
	0x65, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73,
	0x2e, 0x41, 0x64, 0x64, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4b, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x43, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x73, 0x12, 0x1c, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e,
	0x47, 0x65, 0x74, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1d, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x47, 0x65,
	0x74, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x54, 0x0a, 0x0d, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x73, 0x12, 0x1f, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73,
	0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4b, 0x0a, 0x0a, 0x53, 0x65, 0x74, 0x4d,
	0x61, 0x72, 0x6b, 0x65, 0x74, 0x73, 0x12, 0x1c, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69,
	0x65, 0x73, 0x2e, 0x53, 0x65, 0x74, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73,
	0x2e, 0x53, 0x65, 0x74, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4b, 0x0a, 0x0a, 0x41, 0x64, 0x64, 0x4d, 0x61, 0x72, 0x6b,
	0x65, 0x74, 0x73, 0x12, 0x1c, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e,
	0x41, 0x64, 0x64, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1d, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x41, 0x64,
	0x64, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x4b, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x73,
	0x12, 0x1c, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x47, 0x65, 0x74,
	0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d,
	0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4d, 0x61,
	0x72, 0x6b, 0x65, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x54, 0x0a, 0x0d, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x73,
	0x12, 0x1f, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x6d,
	0x6f, 0x76, 0x65, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x20, 0x2e, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x52, 0x65,
	0x6d, 0x6f, 0x76, 0x65, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x8a, 0x01, 0x0a, 0x0d, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x42, 0x15, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69,
	0x65, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x1e, 0x64, 0x61, 0x76, 0x65, 0x6e, 0x73, 0x69, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f,
	0x72, 0x65, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73,
	0xa2, 0x02, 0x03, 0x43, 0x58, 0x58, 0xaa, 0x02, 0x09, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69,
	0x65, 0x73, 0xca, 0x02, 0x09, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0xe2, 0x02,
	0x15, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x09, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x69,
	0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_countries_countries_service_proto_goTypes = []interface{}{
	(*CreateRequest)(nil),         // 0: countries.CreateRequest
	(*UpdateRequest)(nil),         // 1: countries.UpdateRequest
	(*GetRequest)(nil),            // 2: countries.GetRequest
	(*GetListRequest)(nil),        // 3: countries.GetListRequest
	(*DeleteRequest)(nil),         // 4: countries.DeleteRequest
	(*SetFiatsRequest)(nil),       // 5: countries.SetFiatsRequest
	(*AddFiatsRequest)(nil),       // 6: countries.AddFiatsRequest
	(*GetFiatsRequest)(nil),       // 7: countries.GetFiatsRequest
	(*RemoveFiatsRequest)(nil),    // 8: countries.RemoveFiatsRequest
	(*SetCryptosRequest)(nil),     // 9: countries.SetCryptosRequest
	(*AddCryptosRequest)(nil),     // 10: countries.AddCryptosRequest
	(*GetCryptosRequest)(nil),     // 11: countries.GetCryptosRequest
	(*RemoveCryptosRequest)(nil),  // 12: countries.RemoveCryptosRequest
	(*SetMarketsRequest)(nil),     // 13: countries.SetMarketsRequest
	(*AddMarketsRequest)(nil),     // 14: countries.AddMarketsRequest
	(*GetMarketsRequest)(nil),     // 15: countries.GetMarketsRequest
	(*RemoveMarketsRequest)(nil),  // 16: countries.RemoveMarketsRequest
	(*CreateResponse)(nil),        // 17: countries.CreateResponse
	(*UpdateResponse)(nil),        // 18: countries.UpdateResponse
	(*GetResponse)(nil),           // 19: countries.GetResponse
	(*GetListResponse)(nil),       // 20: countries.GetListResponse
	(*DeleteResponse)(nil),        // 21: countries.DeleteResponse
	(*SetFiatsResponse)(nil),      // 22: countries.SetFiatsResponse
	(*AddFiatsResponse)(nil),      // 23: countries.AddFiatsResponse
	(*GetFiatsResponse)(nil),      // 24: countries.GetFiatsResponse
	(*RemoveFiatsResponse)(nil),   // 25: countries.RemoveFiatsResponse
	(*SetCryptosResponse)(nil),    // 26: countries.SetCryptosResponse
	(*AddCryptosResponse)(nil),    // 27: countries.AddCryptosResponse
	(*GetCryptosResponse)(nil),    // 28: countries.GetCryptosResponse
	(*RemoveCryptosResponse)(nil), // 29: countries.RemoveCryptosResponse
	(*SetMarketsResponse)(nil),    // 30: countries.SetMarketsResponse
	(*AddMarketsResponse)(nil),    // 31: countries.AddMarketsResponse
	(*GetMarketsResponse)(nil),    // 32: countries.GetMarketsResponse
	(*RemoveMarketsResponse)(nil), // 33: countries.RemoveMarketsResponse
}
var file_countries_countries_service_proto_depIdxs = []int32{
	0,  // 0: countries.Service.Create:input_type -> countries.CreateRequest
	1,  // 1: countries.Service.Update:input_type -> countries.UpdateRequest
	2,  // 2: countries.Service.Get:input_type -> countries.GetRequest
	3,  // 3: countries.Service.GetList:input_type -> countries.GetListRequest
	4,  // 4: countries.Service.Delete:input_type -> countries.DeleteRequest
	5,  // 5: countries.Service.SetFiats:input_type -> countries.SetFiatsRequest
	6,  // 6: countries.Service.AddFiats:input_type -> countries.AddFiatsRequest
	7,  // 7: countries.Service.GetFiats:input_type -> countries.GetFiatsRequest
	8,  // 8: countries.Service.RemoveFiats:input_type -> countries.RemoveFiatsRequest
	9,  // 9: countries.Service.SetCryptos:input_type -> countries.SetCryptosRequest
	10, // 10: countries.Service.AddCryptos:input_type -> countries.AddCryptosRequest
	11, // 11: countries.Service.GetCryptos:input_type -> countries.GetCryptosRequest
	12, // 12: countries.Service.RemoveCryptos:input_type -> countries.RemoveCryptosRequest
	13, // 13: countries.Service.SetMarkets:input_type -> countries.SetMarketsRequest
	14, // 14: countries.Service.AddMarkets:input_type -> countries.AddMarketsRequest
	15, // 15: countries.Service.GetMarkets:input_type -> countries.GetMarketsRequest
	16, // 16: countries.Service.RemoveMarkets:input_type -> countries.RemoveMarketsRequest
	17, // 17: countries.Service.Create:output_type -> countries.CreateResponse
	18, // 18: countries.Service.Update:output_type -> countries.UpdateResponse
	19, // 19: countries.Service.Get:output_type -> countries.GetResponse
	20, // 20: countries.Service.GetList:output_type -> countries.GetListResponse
	21, // 21: countries.Service.Delete:output_type -> countries.DeleteResponse
	22, // 22: countries.Service.SetFiats:output_type -> countries.SetFiatsResponse
	23, // 23: countries.Service.AddFiats:output_type -> countries.AddFiatsResponse
	24, // 24: countries.Service.GetFiats:output_type -> countries.GetFiatsResponse
	25, // 25: countries.Service.RemoveFiats:output_type -> countries.RemoveFiatsResponse
	26, // 26: countries.Service.SetCryptos:output_type -> countries.SetCryptosResponse
	27, // 27: countries.Service.AddCryptos:output_type -> countries.AddCryptosResponse
	28, // 28: countries.Service.GetCryptos:output_type -> countries.GetCryptosResponse
	29, // 29: countries.Service.RemoveCryptos:output_type -> countries.RemoveCryptosResponse
	30, // 30: countries.Service.SetMarkets:output_type -> countries.SetMarketsResponse
	31, // 31: countries.Service.AddMarkets:output_type -> countries.AddMarketsResponse
	32, // 32: countries.Service.GetMarkets:output_type -> countries.GetMarketsResponse
	33, // 33: countries.Service.RemoveMarkets:output_type -> countries.RemoveMarketsResponse
	17, // [17:34] is the sub-list for method output_type
	0,  // [0:17] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_countries_countries_service_proto_init() }
func file_countries_countries_service_proto_init() {
	if File_countries_countries_service_proto != nil {
		return
	}
	file_countries_countries_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_countries_countries_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_countries_countries_service_proto_goTypes,
		DependencyIndexes: file_countries_countries_service_proto_depIdxs,
	}.Build()
	File_countries_countries_service_proto = out.File
	file_countries_countries_service_proto_rawDesc = nil
	file_countries_countries_service_proto_goTypes = nil
	file_countries_countries_service_proto_depIdxs = nil
}
