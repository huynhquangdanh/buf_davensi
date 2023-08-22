// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: common/timeseries.proto

package common

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

type Timescale int32

const (
	Timescale_TIMESCALE_UNSPECIFIED Timescale = 0
	Timescale_TIMESCALE_1S          Timescale = 1
	Timescale_TIMESCALE_1MN         Timescale = 2
	Timescale_TIMESCALE_5MN         Timescale = 3
	Timescale_TIMESCALE_15MN        Timescale = 4
	Timescale_TIMESCALE_30MN        Timescale = 5
	Timescale_TIMESCALE_1H          Timescale = 6
	Timescale_TIMESCALE_2H          Timescale = 7
	Timescale_TIMESCALE_4H          Timescale = 8
	Timescale_TIMESCALE_8H          Timescale = 9
	Timescale_TIMESCALE_12H         Timescale = 10
	Timescale_TIMESCALE_1D          Timescale = 11
	Timescale_TIMESCALE_1W          Timescale = 12
	Timescale_TIMESCALE_1MTH        Timescale = 13
	Timescale_TIMESCALE_3MTH        Timescale = 14
	Timescale_TIMESCALE_6MTH        Timescale = 15
	Timescale_TIMESCALE_1Y          Timescale = 16
)

// Enum value maps for Timescale.
var (
	Timescale_name = map[int32]string{
		0:  "TIMESCALE_UNSPECIFIED",
		1:  "TIMESCALE_1S",
		2:  "TIMESCALE_1MN",
		3:  "TIMESCALE_5MN",
		4:  "TIMESCALE_15MN",
		5:  "TIMESCALE_30MN",
		6:  "TIMESCALE_1H",
		7:  "TIMESCALE_2H",
		8:  "TIMESCALE_4H",
		9:  "TIMESCALE_8H",
		10: "TIMESCALE_12H",
		11: "TIMESCALE_1D",
		12: "TIMESCALE_1W",
		13: "TIMESCALE_1MTH",
		14: "TIMESCALE_3MTH",
		15: "TIMESCALE_6MTH",
		16: "TIMESCALE_1Y",
	}
	Timescale_value = map[string]int32{
		"TIMESCALE_UNSPECIFIED": 0,
		"TIMESCALE_1S":          1,
		"TIMESCALE_1MN":         2,
		"TIMESCALE_5MN":         3,
		"TIMESCALE_15MN":        4,
		"TIMESCALE_30MN":        5,
		"TIMESCALE_1H":          6,
		"TIMESCALE_2H":          7,
		"TIMESCALE_4H":          8,
		"TIMESCALE_8H":          9,
		"TIMESCALE_12H":         10,
		"TIMESCALE_1D":          11,
		"TIMESCALE_1W":          12,
		"TIMESCALE_1MTH":        13,
		"TIMESCALE_3MTH":        14,
		"TIMESCALE_6MTH":        15,
		"TIMESCALE_1Y":          16,
	}
)

func (x Timescale) Enum() *Timescale {
	p := new(Timescale)
	*p = x
	return p
}

func (x Timescale) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Timescale) Descriptor() protoreflect.EnumDescriptor {
	return file_common_timeseries_proto_enumTypes[0].Descriptor()
}

func (Timescale) Type() protoreflect.EnumType {
	return &file_common_timeseries_proto_enumTypes[0]
}

func (x Timescale) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Timescale.Descriptor instead.
func (Timescale) EnumDescriptor() ([]byte, []int) {
	return file_common_timeseries_proto_rawDescGZIP(), []int{0}
}

var File_common_timeseries_proto protoreflect.FileDescriptor

var file_common_timeseries_proto_rawDesc = []byte{
	0x0a, 0x17, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x65, 0x72,
	0x69, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2a, 0xd3, 0x02, 0x0a, 0x09, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x63, 0x61, 0x6c, 0x65, 0x12,
	0x19, 0x0a, 0x15, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x55, 0x4e, 0x53,
	0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c, 0x54, 0x49,
	0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x31, 0x53, 0x10, 0x01, 0x12, 0x11, 0x0a, 0x0d,
	0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x31, 0x4d, 0x4e, 0x10, 0x02, 0x12,
	0x11, 0x0a, 0x0d, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x35, 0x4d, 0x4e,
	0x10, 0x03, 0x12, 0x12, 0x0a, 0x0e, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f,
	0x31, 0x35, 0x4d, 0x4e, 0x10, 0x04, 0x12, 0x12, 0x0a, 0x0e, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43,
	0x41, 0x4c, 0x45, 0x5f, 0x33, 0x30, 0x4d, 0x4e, 0x10, 0x05, 0x12, 0x10, 0x0a, 0x0c, 0x54, 0x49,
	0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x31, 0x48, 0x10, 0x06, 0x12, 0x10, 0x0a, 0x0c,
	0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x32, 0x48, 0x10, 0x07, 0x12, 0x10,
	0x0a, 0x0c, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x34, 0x48, 0x10, 0x08,
	0x12, 0x10, 0x0a, 0x0c, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x38, 0x48,
	0x10, 0x09, 0x12, 0x11, 0x0a, 0x0d, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f,
	0x31, 0x32, 0x48, 0x10, 0x0a, 0x12, 0x10, 0x0a, 0x0c, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41,
	0x4c, 0x45, 0x5f, 0x31, 0x44, 0x10, 0x0b, 0x12, 0x10, 0x0a, 0x0c, 0x54, 0x49, 0x4d, 0x45, 0x53,
	0x43, 0x41, 0x4c, 0x45, 0x5f, 0x31, 0x57, 0x10, 0x0c, 0x12, 0x12, 0x0a, 0x0e, 0x54, 0x49, 0x4d,
	0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x31, 0x4d, 0x54, 0x48, 0x10, 0x0d, 0x12, 0x12, 0x0a,
	0x0e, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x33, 0x4d, 0x54, 0x48, 0x10,
	0x0e, 0x12, 0x12, 0x0a, 0x0e, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41, 0x4c, 0x45, 0x5f, 0x36,
	0x4d, 0x54, 0x48, 0x10, 0x0f, 0x12, 0x10, 0x0a, 0x0c, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x43, 0x41,
	0x4c, 0x45, 0x5f, 0x31, 0x59, 0x10, 0x10, 0x42, 0x72, 0x0a, 0x0a, 0x63, 0x6f, 0x6d, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x42, 0x0f, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x65, 0x72, 0x69, 0x65,
	0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x1b, 0x64, 0x61, 0x76, 0x65, 0x6e, 0x73,
	0x69, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0xa2, 0x02, 0x03, 0x43, 0x58, 0x58, 0xaa, 0x02, 0x06, 0x43, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0xca, 0x02, 0x06, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0xe2, 0x02, 0x12,
	0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0xea, 0x02, 0x06, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_common_timeseries_proto_rawDescOnce sync.Once
	file_common_timeseries_proto_rawDescData = file_common_timeseries_proto_rawDesc
)

func file_common_timeseries_proto_rawDescGZIP() []byte {
	file_common_timeseries_proto_rawDescOnce.Do(func() {
		file_common_timeseries_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_timeseries_proto_rawDescData)
	})
	return file_common_timeseries_proto_rawDescData
}

var file_common_timeseries_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_common_timeseries_proto_goTypes = []interface{}{
	(Timescale)(0), // 0: common.Timescale
}
var file_common_timeseries_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_timeseries_proto_init() }
func file_common_timeseries_proto_init() {
	if File_common_timeseries_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_common_timeseries_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_timeseries_proto_goTypes,
		DependencyIndexes: file_common_timeseries_proto_depIdxs,
		EnumInfos:         file_common_timeseries_proto_enumTypes,
	}.Build()
	File_common_timeseries_proto = out.File
	file_common_timeseries_proto_rawDesc = nil
	file_common_timeseries_proto_goTypes = nil
	file_common_timeseries_proto_depIdxs = nil
}