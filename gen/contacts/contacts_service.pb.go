// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: contacts/contacts_service.proto

package contacts

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

var File_contacts_contacts_service_proto protoreflect.FileDescriptor

var file_contacts_contacts_service_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61,
	0x63, 0x74, 0x73, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x1a, 0x17, 0x63, 0x6f, 0x6e,
	0x74, 0x61, 0x63, 0x74, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x32, 0xc0, 0x02, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x3d, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x17, 0x2e, 0x63, 0x6f, 0x6e,
	0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x3d, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x17, 0x2e, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x63, 0x74, 0x73, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x18, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x34,
	0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x14, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73,
	0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x63, 0x6f,
	0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x42, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x12,
	0x18, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4c, 0x69,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x63, 0x74, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x3d, 0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x12, 0x17, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x63, 0x6f,
	0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x83, 0x01, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x2e,
	0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x42, 0x14, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63,
	0x74, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x1d, 0x64, 0x61, 0x76, 0x65, 0x6e, 0x73, 0x69, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f,
	0x72, 0x65, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0xa2,
	0x02, 0x03, 0x43, 0x58, 0x58, 0xaa, 0x02, 0x08, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73,
	0xca, 0x02, 0x08, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0xe2, 0x02, 0x14, 0x43, 0x6f,
	0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0xea, 0x02, 0x08, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_contacts_contacts_service_proto_goTypes = []interface{}{
	(*CreateRequest)(nil),   // 0: contacts.CreateRequest
	(*UpdateRequest)(nil),   // 1: contacts.UpdateRequest
	(*GetRequest)(nil),      // 2: contacts.GetRequest
	(*GetListRequest)(nil),  // 3: contacts.GetListRequest
	(*DeleteRequest)(nil),   // 4: contacts.DeleteRequest
	(*CreateResponse)(nil),  // 5: contacts.CreateResponse
	(*UpdateResponse)(nil),  // 6: contacts.UpdateResponse
	(*GetResponse)(nil),     // 7: contacts.GetResponse
	(*GetListResponse)(nil), // 8: contacts.GetListResponse
	(*DeleteResponse)(nil),  // 9: contacts.DeleteResponse
}
var file_contacts_contacts_service_proto_depIdxs = []int32{
	0, // 0: contacts.Service.Create:input_type -> contacts.CreateRequest
	1, // 1: contacts.Service.Update:input_type -> contacts.UpdateRequest
	2, // 2: contacts.Service.Get:input_type -> contacts.GetRequest
	3, // 3: contacts.Service.GetList:input_type -> contacts.GetListRequest
	4, // 4: contacts.Service.Delete:input_type -> contacts.DeleteRequest
	5, // 5: contacts.Service.Create:output_type -> contacts.CreateResponse
	6, // 6: contacts.Service.Update:output_type -> contacts.UpdateResponse
	7, // 7: contacts.Service.Get:output_type -> contacts.GetResponse
	8, // 8: contacts.Service.GetList:output_type -> contacts.GetListResponse
	9, // 9: contacts.Service.Delete:output_type -> contacts.DeleteResponse
	5, // [5:10] is the sub-list for method output_type
	0, // [0:5] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_contacts_contacts_service_proto_init() }
func file_contacts_contacts_service_proto_init() {
	if File_contacts_contacts_service_proto != nil {
		return
	}
	file_contacts_contacts_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_contacts_contacts_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_contacts_contacts_service_proto_goTypes,
		DependencyIndexes: file_contacts_contacts_service_proto_depIdxs,
	}.Build()
	File_contacts_contacts_service_proto = out.File
	file_contacts_contacts_service_proto_rawDesc = nil
	file_contacts_contacts_service_proto_goTypes = nil
	file_contacts_contacts_service_proto_depIdxs = nil
}
