// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: tradingpairs/tradingpairs_service.proto

package tradingpairs

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

var File_tradingpairs_tradingpairs_service_proto protoreflect.FileDescriptor

var file_tradingpairs_tradingpairs_service_proto_rawDesc = []byte{
	0x0a, 0x27, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x2f, 0x74,
	0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x5f, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x74, 0x72, 0x61, 0x64, 0x69,
	0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x1a, 0x1f, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67,
	0x70, 0x61, 0x69, 0x72, 0x73, 0x2f, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69,
	0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xe8, 0x02, 0x0a, 0x07, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x45, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x1b,
	0x2e, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x2e, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x74, 0x72,
	0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x45, 0x0a, 0x06, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1b, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x70,
	0x61, 0x69, 0x72, 0x73, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72,
	0x73, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x3c, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x18, 0x2e, 0x74, 0x72, 0x61, 0x64,
	0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69,
	0x72, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x4a, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x1c, 0x2e, 0x74, 0x72,
	0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4c, 0x69,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x74, 0x72, 0x61, 0x64,
	0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x45, 0x0a, 0x06,
	0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x1b, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67,
	0x70, 0x61, 0x69, 0x72, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69,
	0x72, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0x9f, 0x01, 0x0a, 0x10, 0x63, 0x6f, 0x6d, 0x2e, 0x74, 0x72, 0x61, 0x64,
	0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x42, 0x18, 0x54, 0x72, 0x61, 0x64, 0x69, 0x6e,
	0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x21, 0x64, 0x61, 0x76, 0x65, 0x6e, 0x73, 0x69, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x74, 0x72, 0x61, 0x64, 0x69,
	0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0xa2, 0x02, 0x03, 0x54, 0x58, 0x58, 0xaa, 0x02, 0x0c,
	0x54, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0xca, 0x02, 0x0c, 0x54,
	0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0xe2, 0x02, 0x18, 0x54, 0x72,
	0x61, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x61, 0x69, 0x72, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0c, 0x54, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67,
	0x70, 0x61, 0x69, 0x72, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_tradingpairs_tradingpairs_service_proto_goTypes = []interface{}{
	(*CreateRequest)(nil),   // 0: tradingpairs.CreateRequest
	(*UpdateRequest)(nil),   // 1: tradingpairs.UpdateRequest
	(*GetRequest)(nil),      // 2: tradingpairs.GetRequest
	(*GetListRequest)(nil),  // 3: tradingpairs.GetListRequest
	(*DeleteRequest)(nil),   // 4: tradingpairs.DeleteRequest
	(*CreateResponse)(nil),  // 5: tradingpairs.CreateResponse
	(*UpdateResponse)(nil),  // 6: tradingpairs.UpdateResponse
	(*GetResponse)(nil),     // 7: tradingpairs.GetResponse
	(*GetListResponse)(nil), // 8: tradingpairs.GetListResponse
	(*DeleteResponse)(nil),  // 9: tradingpairs.DeleteResponse
}
var file_tradingpairs_tradingpairs_service_proto_depIdxs = []int32{
	0, // 0: tradingpairs.Service.Create:input_type -> tradingpairs.CreateRequest
	1, // 1: tradingpairs.Service.Update:input_type -> tradingpairs.UpdateRequest
	2, // 2: tradingpairs.Service.Get:input_type -> tradingpairs.GetRequest
	3, // 3: tradingpairs.Service.GetList:input_type -> tradingpairs.GetListRequest
	4, // 4: tradingpairs.Service.Delete:input_type -> tradingpairs.DeleteRequest
	5, // 5: tradingpairs.Service.Create:output_type -> tradingpairs.CreateResponse
	6, // 6: tradingpairs.Service.Update:output_type -> tradingpairs.UpdateResponse
	7, // 7: tradingpairs.Service.Get:output_type -> tradingpairs.GetResponse
	8, // 8: tradingpairs.Service.GetList:output_type -> tradingpairs.GetListResponse
	9, // 9: tradingpairs.Service.Delete:output_type -> tradingpairs.DeleteResponse
	5, // [5:10] is the sub-list for method output_type
	0, // [0:5] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_tradingpairs_tradingpairs_service_proto_init() }
func file_tradingpairs_tradingpairs_service_proto_init() {
	if File_tradingpairs_tradingpairs_service_proto != nil {
		return
	}
	file_tradingpairs_tradingpairs_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_tradingpairs_tradingpairs_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_tradingpairs_tradingpairs_service_proto_goTypes,
		DependencyIndexes: file_tradingpairs_tradingpairs_service_proto_depIdxs,
	}.Build()
	File_tradingpairs_tradingpairs_service_proto = out.File
	file_tradingpairs_tradingpairs_service_proto_rawDesc = nil
	file_tradingpairs_tradingpairs_service_proto_goTypes = nil
	file_tradingpairs_tradingpairs_service_proto_depIdxs = nil
}
