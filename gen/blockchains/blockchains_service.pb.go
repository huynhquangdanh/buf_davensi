// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: blockchains/blockchains_service.proto

package blockchains

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

var File_blockchains_blockchains_service_proto protoreflect.FileDescriptor

var file_blockchains_blockchains_service_proto_rawDesc = []byte{
	0x0a, 0x25, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2f, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x73, 0x1a, 0x1d, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x73, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x32, 0xb1, 0x05, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x43, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x1a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61,
	0x69, 0x6e, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1a,
	0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x03, 0x47, 0x65, 0x74,
	0x12, 0x17, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x48, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74,
	0x12, 0x1b, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x47,
	0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4c,
	0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12,
	0x43, 0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x1a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61,
	0x69, 0x6e, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x4f, 0x0a, 0x0a, 0x53, 0x65, 0x74, 0x43, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x73, 0x12, 0x1e, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73,
	0x2e, 0x53, 0x65, 0x74, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73,
	0x2e, 0x53, 0x65, 0x74, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4f, 0x0a, 0x0a, 0x41, 0x64, 0x64, 0x43, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x73, 0x12, 0x1e, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x73, 0x2e, 0x41, 0x64, 0x64, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x73, 0x2e, 0x41, 0x64, 0x64, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x55, 0x0a, 0x0c, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x12, 0x20, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x73, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x43, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x43, 0x72, 0x79,
	0x70, 0x74, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x58, 0x0a,
	0x0d, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x12, 0x21,
	0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x52, 0x65, 0x6d,
	0x6f, 0x76, 0x65, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x22, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x2e,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x98, 0x01, 0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x42, 0x17, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x20, 0x64, 0x61, 0x76, 0x65, 0x6e, 0x73, 0x69, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0xa2, 0x02, 0x03, 0x42, 0x58, 0x58, 0xaa, 0x02,
	0x0b, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0xca, 0x02, 0x0b, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0xe2, 0x02, 0x17, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0b, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69,
	0x6e, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_blockchains_blockchains_service_proto_goTypes = []interface{}{
	(*CreateRequest)(nil),         // 0: blockchains.CreateRequest
	(*UpdateRequest)(nil),         // 1: blockchains.UpdateRequest
	(*GetRequest)(nil),            // 2: blockchains.GetRequest
	(*GetListRequest)(nil),        // 3: blockchains.GetListRequest
	(*DeleteRequest)(nil),         // 4: blockchains.DeleteRequest
	(*SetCryptosRequest)(nil),     // 5: blockchains.SetCryptosRequest
	(*AddCryptosRequest)(nil),     // 6: blockchains.AddCryptosRequest
	(*UpdateCryptoRequest)(nil),   // 7: blockchains.UpdateCryptoRequest
	(*RemoveCryptosRequest)(nil),  // 8: blockchains.RemoveCryptosRequest
	(*CreateResponse)(nil),        // 9: blockchains.CreateResponse
	(*UpdateResponse)(nil),        // 10: blockchains.UpdateResponse
	(*GetResponse)(nil),           // 11: blockchains.GetResponse
	(*GetListResponse)(nil),       // 12: blockchains.GetListResponse
	(*DeleteResponse)(nil),        // 13: blockchains.DeleteResponse
	(*SetCryptosResponse)(nil),    // 14: blockchains.SetCryptosResponse
	(*AddCryptosResponse)(nil),    // 15: blockchains.AddCryptosResponse
	(*UpdateCryptoResponse)(nil),  // 16: blockchains.UpdateCryptoResponse
	(*RemoveCryptosResponse)(nil), // 17: blockchains.RemoveCryptosResponse
}
var file_blockchains_blockchains_service_proto_depIdxs = []int32{
	0,  // 0: blockchains.Service.Create:input_type -> blockchains.CreateRequest
	1,  // 1: blockchains.Service.Update:input_type -> blockchains.UpdateRequest
	2,  // 2: blockchains.Service.Get:input_type -> blockchains.GetRequest
	3,  // 3: blockchains.Service.GetList:input_type -> blockchains.GetListRequest
	4,  // 4: blockchains.Service.Delete:input_type -> blockchains.DeleteRequest
	5,  // 5: blockchains.Service.SetCryptos:input_type -> blockchains.SetCryptosRequest
	6,  // 6: blockchains.Service.AddCryptos:input_type -> blockchains.AddCryptosRequest
	7,  // 7: blockchains.Service.UpdateCrypto:input_type -> blockchains.UpdateCryptoRequest
	8,  // 8: blockchains.Service.RemoveCryptos:input_type -> blockchains.RemoveCryptosRequest
	9,  // 9: blockchains.Service.Create:output_type -> blockchains.CreateResponse
	10, // 10: blockchains.Service.Update:output_type -> blockchains.UpdateResponse
	11, // 11: blockchains.Service.Get:output_type -> blockchains.GetResponse
	12, // 12: blockchains.Service.GetList:output_type -> blockchains.GetListResponse
	13, // 13: blockchains.Service.Delete:output_type -> blockchains.DeleteResponse
	14, // 14: blockchains.Service.SetCryptos:output_type -> blockchains.SetCryptosResponse
	15, // 15: blockchains.Service.AddCryptos:output_type -> blockchains.AddCryptosResponse
	16, // 16: blockchains.Service.UpdateCrypto:output_type -> blockchains.UpdateCryptoResponse
	17, // 17: blockchains.Service.RemoveCryptos:output_type -> blockchains.RemoveCryptosResponse
	9,  // [9:18] is the sub-list for method output_type
	0,  // [0:9] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_blockchains_blockchains_service_proto_init() }
func file_blockchains_blockchains_service_proto_init() {
	if File_blockchains_blockchains_service_proto != nil {
		return
	}
	file_blockchains_blockchains_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_blockchains_blockchains_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_blockchains_blockchains_service_proto_goTypes,
		DependencyIndexes: file_blockchains_blockchains_service_proto_depIdxs,
	}.Build()
	File_blockchains_blockchains_service_proto = out.File
	file_blockchains_blockchains_service_proto_rawDesc = nil
	file_blockchains_blockchains_service_proto_goTypes = nil
	file_blockchains_blockchains_service_proto_depIdxs = nil
}
