// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: fiats/fiats_service.proto

package fiats

import (
	uoms "davensi.com/core/gen/uoms"
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

var File_fiats_fiats_service_proto protoreflect.FileDescriptor

var file_fiats_fiats_service_proto_rawDesc = []byte{
	0x0a, 0x19, 0x66, 0x69, 0x61, 0x74, 0x73, 0x2f, 0x66, 0x69, 0x61, 0x74, 0x73, 0x5f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x66, 0x69, 0x61,
	0x74, 0x73, 0x1a, 0x11, 0x66, 0x69, 0x61, 0x74, 0x73, 0x2f, 0x66, 0x69, 0x61, 0x74, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x75, 0x6f, 0x6d, 0x73, 0x2f, 0x75, 0x6f, 0x6d, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xa0, 0x02, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x37, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x14, 0x2e, 0x66,
	0x69, 0x61, 0x74, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x15, 0x2e, 0x66, 0x69, 0x61, 0x74, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x37, 0x0a, 0x06, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x14, 0x2e, 0x66, 0x69, 0x61, 0x74, 0x73, 0x2e, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x66, 0x69,
	0x61, 0x74, 0x73, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x2d, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x10, 0x2e, 0x75, 0x6f,
	0x6d, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e,
	0x66, 0x69, 0x61, 0x74, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x12, 0x3c, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x15,
	0x2e, 0x66, 0x69, 0x61, 0x74, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x66, 0x69, 0x61, 0x74, 0x73, 0x2e, 0x47, 0x65,
	0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30,
	0x01, 0x12, 0x36, 0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x13, 0x2e, 0x75, 0x6f,
	0x6d, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x15, 0x2e, 0x66, 0x69, 0x61, 0x74, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x6e, 0x0a, 0x09, 0x63, 0x6f, 0x6d,
	0x2e, 0x66, 0x69, 0x61, 0x74, 0x73, 0x42, 0x11, 0x46, 0x69, 0x61, 0x74, 0x73, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x1a, 0x64, 0x61, 0x76,
	0x65, 0x6e, 0x73, 0x69, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x67, 0x65,
	0x6e, 0x2f, 0x66, 0x69, 0x61, 0x74, 0x73, 0xa2, 0x02, 0x03, 0x46, 0x58, 0x58, 0xaa, 0x02, 0x05,
	0x46, 0x69, 0x61, 0x74, 0x73, 0xca, 0x02, 0x05, 0x46, 0x69, 0x61, 0x74, 0x73, 0xe2, 0x02, 0x11,
	0x46, 0x69, 0x61, 0x74, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0xea, 0x02, 0x05, 0x46, 0x69, 0x61, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var file_fiats_fiats_service_proto_goTypes = []interface{}{
	(*CreateRequest)(nil),      // 0: fiats.CreateRequest
	(*UpdateRequest)(nil),      // 1: fiats.UpdateRequest
	(*uoms.GetRequest)(nil),    // 2: uoms.GetRequest
	(*GetListRequest)(nil),     // 3: fiats.GetListRequest
	(*uoms.DeleteRequest)(nil), // 4: uoms.DeleteRequest
	(*CreateResponse)(nil),     // 5: fiats.CreateResponse
	(*UpdateResponse)(nil),     // 6: fiats.UpdateResponse
	(*GetResponse)(nil),        // 7: fiats.GetResponse
	(*GetListResponse)(nil),    // 8: fiats.GetListResponse
	(*DeleteResponse)(nil),     // 9: fiats.DeleteResponse
}
var file_fiats_fiats_service_proto_depIdxs = []int32{
	0, // 0: fiats.Service.Create:input_type -> fiats.CreateRequest
	1, // 1: fiats.Service.Update:input_type -> fiats.UpdateRequest
	2, // 2: fiats.Service.Get:input_type -> uoms.GetRequest
	3, // 3: fiats.Service.GetList:input_type -> fiats.GetListRequest
	4, // 4: fiats.Service.Delete:input_type -> uoms.DeleteRequest
	5, // 5: fiats.Service.Create:output_type -> fiats.CreateResponse
	6, // 6: fiats.Service.Update:output_type -> fiats.UpdateResponse
	7, // 7: fiats.Service.Get:output_type -> fiats.GetResponse
	8, // 8: fiats.Service.GetList:output_type -> fiats.GetListResponse
	9, // 9: fiats.Service.Delete:output_type -> fiats.DeleteResponse
	5, // [5:10] is the sub-list for method output_type
	0, // [0:5] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_fiats_fiats_service_proto_init() }
func file_fiats_fiats_service_proto_init() {
	if File_fiats_fiats_service_proto != nil {
		return
	}
	file_fiats_fiats_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_fiats_fiats_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_fiats_fiats_service_proto_goTypes,
		DependencyIndexes: file_fiats_fiats_service_proto_depIdxs,
	}.Build()
	File_fiats_fiats_service_proto = out.File
	file_fiats_fiats_service_proto_rawDesc = nil
	file_fiats_fiats_service_proto_goTypes = nil
	file_fiats_fiats_service_proto_depIdxs = nil
}
