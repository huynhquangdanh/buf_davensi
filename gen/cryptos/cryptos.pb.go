// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: cryptos/cryptos.proto

package cryptos

import (
	common "davensi.com/core/gen/common"
	uoms "davensi.com/core/gen/uoms"
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

type Type int32

const (
	Type_TYPE_UNSPECIFIED Type = 0
	Type_TYPE_UTILITY     Type = 1
	Type_TYPE_SECURITY    Type = 2
	Type_TYPE_PAYMENT     Type = 3
	Type_TYPE_EXCHANGE    Type = 4
	Type_TYPE_NFT         Type = 5
	Type_TYPE_STABLECOIN  Type = 6
	Type_TYPE_DEFI        Type = 7
	Type_TYPE_TOKEN       Type = 8
	Type_TYPE_ASSETBACKED Type = 9
)

// Enum value maps for Type.
var (
	Type_name = map[int32]string{
		0: "TYPE_UNSPECIFIED",
		1: "TYPE_UTILITY",
		2: "TYPE_SECURITY",
		3: "TYPE_PAYMENT",
		4: "TYPE_EXCHANGE",
		5: "TYPE_NFT",
		6: "TYPE_STABLECOIN",
		7: "TYPE_DEFI",
		8: "TYPE_TOKEN",
		9: "TYPE_ASSETBACKED",
	}
	Type_value = map[string]int32{
		"TYPE_UNSPECIFIED": 0,
		"TYPE_UTILITY":     1,
		"TYPE_SECURITY":    2,
		"TYPE_PAYMENT":     3,
		"TYPE_EXCHANGE":    4,
		"TYPE_NFT":         5,
		"TYPE_STABLECOIN":  6,
		"TYPE_DEFI":        7,
		"TYPE_TOKEN":       8,
		"TYPE_ASSETBACKED": 9,
	}
)

func (x Type) Enum() *Type {
	p := new(Type)
	*p = x
	return p
}

func (x Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Type) Descriptor() protoreflect.EnumDescriptor {
	return file_cryptos_cryptos_proto_enumTypes[0].Descriptor()
}

func (Type) Type() protoreflect.EnumType {
	return &file_cryptos_cryptos_proto_enumTypes[0]
}

func (x Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Type.Descriptor instead.
func (Type) EnumDescriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{0}
}

type TypeList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	List []Type `protobuf:"varint,1,rep,packed,name=list,proto3,enum=cryptos.Type" json:"list,omitempty"`
}

func (x *TypeList) Reset() {
	*x = TypeList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TypeList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TypeList) ProtoMessage() {}

func (x *TypeList) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TypeList.ProtoReflect.Descriptor instead.
func (*TypeList) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{0}
}

func (x *TypeList) GetList() []Type {
	if x != nil {
		return x.List
	}
	return nil
}

// Backed by tables 'uoms' + 'cryptos'
// cryptos are uoms of type 2
type Crypto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Fields from table 'uoms'
	Uom *uoms.UoM `protobuf:"bytes,1,opt,name=uom,proto3" json:"uom,omitempty"`
	// Fields from table 'cryptos'
	CryptoType Type `protobuf:"varint,2,opt,name=crypto_type,json=cryptoType,proto3,enum=cryptos.Type" json:"crypto_type,omitempty"`
}

func (x *Crypto) Reset() {
	*x = Crypto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Crypto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Crypto) ProtoMessage() {}

func (x *Crypto) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Crypto.ProtoReflect.Descriptor instead.
func (*Crypto) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{1}
}

func (x *Crypto) GetUom() *uoms.UoM {
	if x != nil {
		return x.Uom
	}
	return nil
}

func (x *Crypto) GetCryptoType() Type {
	if x != nil {
		return x.CryptoType
	}
	return Type_TYPE_UNSPECIFIED
}

type List struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	List []*Crypto `protobuf:"bytes,1,rep,name=list,proto3" json:"list,omitempty"`
}

func (x *List) Reset() {
	*x = List{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *List) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*List) ProtoMessage() {}

func (x *List) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use List.ProtoReflect.Descriptor instead.
func (*List) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{2}
}

func (x *List) GetList() []*Crypto {
	if x != nil {
		return x.List
	}
	return nil
}

type CreateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uom        *uoms.CreateRequest `protobuf:"bytes,1,opt,name=uom,proto3" json:"uom,omitempty"`
	CryptoType *Type               `protobuf:"varint,2,opt,name=crypto_type,json=cryptoType,proto3,enum=cryptos.Type,oneof" json:"crypto_type,omitempty"` // Default: TYPE_UNSPECIFIED
}

func (x *CreateRequest) Reset() {
	*x = CreateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateRequest) ProtoMessage() {}

func (x *CreateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateRequest.ProtoReflect.Descriptor instead.
func (*CreateRequest) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{3}
}

func (x *CreateRequest) GetUom() *uoms.CreateRequest {
	if x != nil {
		return x.Uom
	}
	return nil
}

func (x *CreateRequest) GetCryptoType() Type {
	if x != nil && x.CryptoType != nil {
		return *x.CryptoType
	}
	return Type_TYPE_UNSPECIFIED
}

type CreateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Response:
	//
	//	*CreateResponse_Error
	//	*CreateResponse_Crypto
	Response isCreateResponse_Response `protobuf_oneof:"response"`
}

func (x *CreateResponse) Reset() {
	*x = CreateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateResponse) ProtoMessage() {}

func (x *CreateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateResponse.ProtoReflect.Descriptor instead.
func (*CreateResponse) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{4}
}

func (m *CreateResponse) GetResponse() isCreateResponse_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (x *CreateResponse) GetError() *common.Error {
	if x, ok := x.GetResponse().(*CreateResponse_Error); ok {
		return x.Error
	}
	return nil
}

func (x *CreateResponse) GetCrypto() *Crypto {
	if x, ok := x.GetResponse().(*CreateResponse_Crypto); ok {
		return x.Crypto
	}
	return nil
}

type isCreateResponse_Response interface {
	isCreateResponse_Response()
}

type CreateResponse_Error struct {
	Error *common.Error `protobuf:"bytes,1,opt,name=error,proto3,oneof"`
}

type CreateResponse_Crypto struct {
	Crypto *Crypto `protobuf:"bytes,2,opt,name=crypto,proto3,oneof"`
}

func (*CreateResponse_Error) isCreateResponse_Response() {}

func (*CreateResponse_Crypto) isCreateResponse_Response() {}

type UpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uom        *uoms.UpdateRequest `protobuf:"bytes,1,opt,name=uom,proto3" json:"uom,omitempty"`
	CryptoType *Type               `protobuf:"varint,2,opt,name=crypto_type,json=cryptoType,proto3,enum=cryptos.Type,oneof" json:"crypto_type,omitempty"`
}

func (x *UpdateRequest) Reset() {
	*x = UpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRequest) ProtoMessage() {}

func (x *UpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRequest.ProtoReflect.Descriptor instead.
func (*UpdateRequest) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{5}
}

func (x *UpdateRequest) GetUom() *uoms.UpdateRequest {
	if x != nil {
		return x.Uom
	}
	return nil
}

func (x *UpdateRequest) GetCryptoType() Type {
	if x != nil && x.CryptoType != nil {
		return *x.CryptoType
	}
	return Type_TYPE_UNSPECIFIED
}

type UpdateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Response:
	//
	//	*UpdateResponse_Error
	//	*UpdateResponse_Crypto
	Response isUpdateResponse_Response `protobuf_oneof:"response"`
}

func (x *UpdateResponse) Reset() {
	*x = UpdateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateResponse) ProtoMessage() {}

func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateResponse.ProtoReflect.Descriptor instead.
func (*UpdateResponse) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{6}
}

func (m *UpdateResponse) GetResponse() isUpdateResponse_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (x *UpdateResponse) GetError() *common.Error {
	if x, ok := x.GetResponse().(*UpdateResponse_Error); ok {
		return x.Error
	}
	return nil
}

func (x *UpdateResponse) GetCrypto() *Crypto {
	if x, ok := x.GetResponse().(*UpdateResponse_Crypto); ok {
		return x.Crypto
	}
	return nil
}

type isUpdateResponse_Response interface {
	isUpdateResponse_Response()
}

type UpdateResponse_Error struct {
	Error *common.Error `protobuf:"bytes,1,opt,name=error,proto3,oneof"`
}

type UpdateResponse_Crypto struct {
	Crypto *Crypto `protobuf:"bytes,2,opt,name=crypto,proto3,oneof"`
}

func (*UpdateResponse_Error) isUpdateResponse_Response() {}

func (*UpdateResponse_Crypto) isUpdateResponse_Response() {}

// An error is returned if there is more than one record found.
type GetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Response:
	//
	//	*GetResponse_Error
	//	*GetResponse_Crypto
	Response isGetResponse_Response `protobuf_oneof:"response"`
}

func (x *GetResponse) Reset() {
	*x = GetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetResponse) ProtoMessage() {}

func (x *GetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetResponse.ProtoReflect.Descriptor instead.
func (*GetResponse) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{7}
}

func (m *GetResponse) GetResponse() isGetResponse_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (x *GetResponse) GetError() *common.Error {
	if x, ok := x.GetResponse().(*GetResponse_Error); ok {
		return x.Error
	}
	return nil
}

func (x *GetResponse) GetCrypto() *Crypto {
	if x, ok := x.GetResponse().(*GetResponse_Crypto); ok {
		return x.Crypto
	}
	return nil
}

type isGetResponse_Response interface {
	isGetResponse_Response()
}

type GetResponse_Error struct {
	Error *common.Error `protobuf:"bytes,1,opt,name=error,proto3,oneof"`
}

type GetResponse_Crypto struct {
	Crypto *Crypto `protobuf:"bytes,2,opt,name=crypto,proto3,oneof"`
}

func (*GetResponse_Error) isGetResponse_Response() {}

func (*GetResponse_Crypto) isGetResponse_Response() {}

// GetCryptoList will use SQL 'LIKE' instead of '=' for string fields
type GetListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uom        *uoms.GetListRequest `protobuf:"bytes,1,opt,name=uom,proto3,oneof" json:"uom,omitempty"`
	CryptoType *Type                `protobuf:"varint,2,opt,name=crypto_type,json=cryptoType,proto3,enum=cryptos.Type,oneof" json:"crypto_type,omitempty"`
}

func (x *GetListRequest) Reset() {
	*x = GetListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetListRequest) ProtoMessage() {}

func (x *GetListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetListRequest.ProtoReflect.Descriptor instead.
func (*GetListRequest) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{8}
}

func (x *GetListRequest) GetUom() *uoms.GetListRequest {
	if x != nil {
		return x.Uom
	}
	return nil
}

func (x *GetListRequest) GetCryptoType() Type {
	if x != nil && x.CryptoType != nil {
		return *x.CryptoType
	}
	return Type_TYPE_UNSPECIFIED
}

type GetListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Response:
	//
	//	*GetListResponse_Error
	//	*GetListResponse_Crypto
	Response isGetListResponse_Response `protobuf_oneof:"response"`
}

func (x *GetListResponse) Reset() {
	*x = GetListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetListResponse) ProtoMessage() {}

func (x *GetListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetListResponse.ProtoReflect.Descriptor instead.
func (*GetListResponse) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{9}
}

func (m *GetListResponse) GetResponse() isGetListResponse_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (x *GetListResponse) GetError() *common.Error {
	if x, ok := x.GetResponse().(*GetListResponse_Error); ok {
		return x.Error
	}
	return nil
}

func (x *GetListResponse) GetCrypto() *Crypto {
	if x, ok := x.GetResponse().(*GetListResponse_Crypto); ok {
		return x.Crypto
	}
	return nil
}

type isGetListResponse_Response interface {
	isGetListResponse_Response()
}

type GetListResponse_Error struct {
	Error *common.Error `protobuf:"bytes,1,opt,name=error,proto3,oneof"`
}

type GetListResponse_Crypto struct {
	Crypto *Crypto `protobuf:"bytes,2,opt,name=crypto,proto3,oneof"`
}

func (*GetListResponse_Error) isGetListResponse_Response() {}

func (*GetListResponse_Crypto) isGetListResponse_Response() {}

type DeleteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Response:
	//
	//	*DeleteResponse_Error
	//	*DeleteResponse_Crypto
	Response isDeleteResponse_Response `protobuf_oneof:"response"`
}

func (x *DeleteResponse) Reset() {
	*x = DeleteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptos_cryptos_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteResponse) ProtoMessage() {}

func (x *DeleteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cryptos_cryptos_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteResponse.ProtoReflect.Descriptor instead.
func (*DeleteResponse) Descriptor() ([]byte, []int) {
	return file_cryptos_cryptos_proto_rawDescGZIP(), []int{10}
}

func (m *DeleteResponse) GetResponse() isDeleteResponse_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (x *DeleteResponse) GetError() *common.Error {
	if x, ok := x.GetResponse().(*DeleteResponse_Error); ok {
		return x.Error
	}
	return nil
}

func (x *DeleteResponse) GetCrypto() *Crypto {
	if x, ok := x.GetResponse().(*DeleteResponse_Crypto); ok {
		return x.Crypto
	}
	return nil
}

type isDeleteResponse_Response interface {
	isDeleteResponse_Response()
}

type DeleteResponse_Error struct {
	Error *common.Error `protobuf:"bytes,1,opt,name=error,proto3,oneof"`
}

type DeleteResponse_Crypto struct {
	Crypto *Crypto `protobuf:"bytes,2,opt,name=crypto,proto3,oneof"`
}

func (*DeleteResponse_Error) isDeleteResponse_Response() {}

func (*DeleteResponse_Crypto) isDeleteResponse_Response() {}

var File_cryptos_cryptos_proto protoreflect.FileDescriptor

var file_cryptos_cryptos_proto_rawDesc = []byte{
	0x0a, 0x15, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2f, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73,
	0x1a, 0x13, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x75, 0x6f, 0x6d, 0x73, 0x2f, 0x75, 0x6f, 0x6d, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2d, 0x0a, 0x08, 0x54, 0x79, 0x70, 0x65, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x21, 0x0a, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0e,
	0x32, 0x0d, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x04, 0x6c, 0x69, 0x73, 0x74, 0x22, 0x55, 0x0a, 0x06, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x12,
	0x1b, 0x0a, 0x03, 0x75, 0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x75,
	0x6f, 0x6d, 0x73, 0x2e, 0x55, 0x6f, 0x4d, 0x52, 0x03, 0x75, 0x6f, 0x6d, 0x12, 0x2e, 0x0a, 0x0b,
	0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x0d, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2e, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x0a, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x54, 0x79, 0x70, 0x65, 0x22, 0x2b, 0x0a, 0x04,
	0x4c, 0x69, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2e, 0x43, 0x72, 0x79,
	0x70, 0x74, 0x6f, 0x52, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x22, 0x7b, 0x0a, 0x0d, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x25, 0x0a, 0x03, 0x75, 0x6f,
	0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x75, 0x6f, 0x6d, 0x73, 0x2e, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x03, 0x75, 0x6f,
	0x6d, 0x12, 0x33, 0x0a, 0x0b, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73,
	0x2e, 0x54, 0x79, 0x70, 0x65, 0x48, 0x00, 0x52, 0x0a, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x54,
	0x79, 0x70, 0x65, 0x88, 0x01, 0x01, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x22, 0x6e, 0x0a, 0x0e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x25, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x48, 0x00, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12,
	0x29, 0x0a, 0x06, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0f, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2e, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f,
	0x48, 0x00, 0x52, 0x06, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x42, 0x0a, 0x0a, 0x08, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x7b, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x25, 0x0a, 0x03, 0x75, 0x6f, 0x6d, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x75, 0x6f, 0x6d, 0x73, 0x2e, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x03, 0x75, 0x6f, 0x6d, 0x12, 0x33,
	0x0a, 0x0b, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2e, 0x54, 0x79,
	0x70, 0x65, 0x48, 0x00, 0x52, 0x0a, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x54, 0x79, 0x70, 0x65,
	0x88, 0x01, 0x01, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x22, 0x6e, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x25, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x48, 0x00, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x29, 0x0a, 0x06,
	0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x63,
	0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2e, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x48, 0x00, 0x52,
	0x06, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x42, 0x0a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x6b, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x25, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0d, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x48, 0x00, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x29, 0x0a, 0x06, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x73, 0x2e, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x48, 0x00, 0x52, 0x06, 0x63, 0x72,
	0x79, 0x70, 0x74, 0x6f, 0x42, 0x0a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x8a, 0x01, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x2b, 0x0a, 0x03, 0x75, 0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x14, 0x2e, 0x75, 0x6f, 0x6d, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x03, 0x75, 0x6f, 0x6d, 0x88, 0x01, 0x01,
	0x12, 0x33, 0x0a, 0x0b, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2e,
	0x54, 0x79, 0x70, 0x65, 0x48, 0x01, 0x52, 0x0a, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x54, 0x79,
	0x70, 0x65, 0x88, 0x01, 0x01, 0x42, 0x06, 0x0a, 0x04, 0x5f, 0x75, 0x6f, 0x6d, 0x42, 0x0e, 0x0a,
	0x0c, 0x5f, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x22, 0x6f, 0x0a,
	0x0f, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x25, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0d, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x48, 0x00,
	0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x29, 0x0a, 0x06, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f,
	0x73, 0x2e, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x48, 0x00, 0x52, 0x06, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x42, 0x0a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x6e,
	0x0a, 0x0e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x25, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0d, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x48, 0x00,
	0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x29, 0x0a, 0x06, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f,
	0x73, 0x2e, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x48, 0x00, 0x52, 0x06, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x42, 0x0a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2a, 0xbe,
	0x01, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x10, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x10, 0x0a,
	0x0c, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x54, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x10, 0x01, 0x12,
	0x11, 0x0a, 0x0d, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x45, 0x43, 0x55, 0x52, 0x49, 0x54, 0x59,
	0x10, 0x02, 0x12, 0x10, 0x0a, 0x0c, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x50, 0x41, 0x59, 0x4d, 0x45,
	0x4e, 0x54, 0x10, 0x03, 0x12, 0x11, 0x0a, 0x0d, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x45, 0x58, 0x43,
	0x48, 0x41, 0x4e, 0x47, 0x45, 0x10, 0x04, 0x12, 0x0c, 0x0a, 0x08, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x4e, 0x46, 0x54, 0x10, 0x05, 0x12, 0x13, 0x0a, 0x0f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x54,
	0x41, 0x42, 0x4c, 0x45, 0x43, 0x4f, 0x49, 0x4e, 0x10, 0x06, 0x12, 0x0d, 0x0a, 0x09, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x44, 0x45, 0x46, 0x49, 0x10, 0x07, 0x12, 0x0e, 0x0a, 0x0a, 0x54, 0x59, 0x50,
	0x45, 0x5f, 0x54, 0x4f, 0x4b, 0x45, 0x4e, 0x10, 0x08, 0x12, 0x14, 0x0a, 0x10, 0x54, 0x59, 0x50,
	0x45, 0x5f, 0x41, 0x53, 0x53, 0x45, 0x54, 0x42, 0x41, 0x43, 0x4b, 0x45, 0x44, 0x10, 0x09, 0x42,
	0x75, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x42, 0x0c,
	0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x1c,
	0x64, 0x61, 0x76, 0x65, 0x6e, 0x73, 0x69, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x72, 0x65,
	0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0xa2, 0x02, 0x03, 0x43,
	0x58, 0x58, 0xaa, 0x02, 0x07, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0xca, 0x02, 0x07, 0x43,
	0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0xe2, 0x02, 0x13, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73,
	0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x07, 0x43,
	0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cryptos_cryptos_proto_rawDescOnce sync.Once
	file_cryptos_cryptos_proto_rawDescData = file_cryptos_cryptos_proto_rawDesc
)

func file_cryptos_cryptos_proto_rawDescGZIP() []byte {
	file_cryptos_cryptos_proto_rawDescOnce.Do(func() {
		file_cryptos_cryptos_proto_rawDescData = protoimpl.X.CompressGZIP(file_cryptos_cryptos_proto_rawDescData)
	})
	return file_cryptos_cryptos_proto_rawDescData
}

var file_cryptos_cryptos_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_cryptos_cryptos_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_cryptos_cryptos_proto_goTypes = []interface{}{
	(Type)(0),                   // 0: cryptos.Type
	(*TypeList)(nil),            // 1: cryptos.TypeList
	(*Crypto)(nil),              // 2: cryptos.Crypto
	(*List)(nil),                // 3: cryptos.List
	(*CreateRequest)(nil),       // 4: cryptos.CreateRequest
	(*CreateResponse)(nil),      // 5: cryptos.CreateResponse
	(*UpdateRequest)(nil),       // 6: cryptos.UpdateRequest
	(*UpdateResponse)(nil),      // 7: cryptos.UpdateResponse
	(*GetResponse)(nil),         // 8: cryptos.GetResponse
	(*GetListRequest)(nil),      // 9: cryptos.GetListRequest
	(*GetListResponse)(nil),     // 10: cryptos.GetListResponse
	(*DeleteResponse)(nil),      // 11: cryptos.DeleteResponse
	(*uoms.UoM)(nil),            // 12: uoms.UoM
	(*uoms.CreateRequest)(nil),  // 13: uoms.CreateRequest
	(*common.Error)(nil),        // 14: common.Error
	(*uoms.UpdateRequest)(nil),  // 15: uoms.UpdateRequest
	(*uoms.GetListRequest)(nil), // 16: uoms.GetListRequest
}
var file_cryptos_cryptos_proto_depIdxs = []int32{
	0,  // 0: cryptos.TypeList.list:type_name -> cryptos.Type
	12, // 1: cryptos.Crypto.uom:type_name -> uoms.UoM
	0,  // 2: cryptos.Crypto.crypto_type:type_name -> cryptos.Type
	2,  // 3: cryptos.List.list:type_name -> cryptos.Crypto
	13, // 4: cryptos.CreateRequest.uom:type_name -> uoms.CreateRequest
	0,  // 5: cryptos.CreateRequest.crypto_type:type_name -> cryptos.Type
	14, // 6: cryptos.CreateResponse.error:type_name -> common.Error
	2,  // 7: cryptos.CreateResponse.crypto:type_name -> cryptos.Crypto
	15, // 8: cryptos.UpdateRequest.uom:type_name -> uoms.UpdateRequest
	0,  // 9: cryptos.UpdateRequest.crypto_type:type_name -> cryptos.Type
	14, // 10: cryptos.UpdateResponse.error:type_name -> common.Error
	2,  // 11: cryptos.UpdateResponse.crypto:type_name -> cryptos.Crypto
	14, // 12: cryptos.GetResponse.error:type_name -> common.Error
	2,  // 13: cryptos.GetResponse.crypto:type_name -> cryptos.Crypto
	16, // 14: cryptos.GetListRequest.uom:type_name -> uoms.GetListRequest
	0,  // 15: cryptos.GetListRequest.crypto_type:type_name -> cryptos.Type
	14, // 16: cryptos.GetListResponse.error:type_name -> common.Error
	2,  // 17: cryptos.GetListResponse.crypto:type_name -> cryptos.Crypto
	14, // 18: cryptos.DeleteResponse.error:type_name -> common.Error
	2,  // 19: cryptos.DeleteResponse.crypto:type_name -> cryptos.Crypto
	20, // [20:20] is the sub-list for method output_type
	20, // [20:20] is the sub-list for method input_type
	20, // [20:20] is the sub-list for extension type_name
	20, // [20:20] is the sub-list for extension extendee
	0,  // [0:20] is the sub-list for field type_name
}

func init() { file_cryptos_cryptos_proto_init() }
func file_cryptos_cryptos_proto_init() {
	if File_cryptos_cryptos_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cryptos_cryptos_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TypeList); i {
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
		file_cryptos_cryptos_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Crypto); i {
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
		file_cryptos_cryptos_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*List); i {
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
		file_cryptos_cryptos_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateRequest); i {
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
		file_cryptos_cryptos_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateResponse); i {
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
		file_cryptos_cryptos_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateRequest); i {
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
		file_cryptos_cryptos_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateResponse); i {
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
		file_cryptos_cryptos_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetResponse); i {
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
		file_cryptos_cryptos_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetListRequest); i {
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
		file_cryptos_cryptos_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetListResponse); i {
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
		file_cryptos_cryptos_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteResponse); i {
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
	file_cryptos_cryptos_proto_msgTypes[3].OneofWrappers = []interface{}{}
	file_cryptos_cryptos_proto_msgTypes[4].OneofWrappers = []interface{}{
		(*CreateResponse_Error)(nil),
		(*CreateResponse_Crypto)(nil),
	}
	file_cryptos_cryptos_proto_msgTypes[5].OneofWrappers = []interface{}{}
	file_cryptos_cryptos_proto_msgTypes[6].OneofWrappers = []interface{}{
		(*UpdateResponse_Error)(nil),
		(*UpdateResponse_Crypto)(nil),
	}
	file_cryptos_cryptos_proto_msgTypes[7].OneofWrappers = []interface{}{
		(*GetResponse_Error)(nil),
		(*GetResponse_Crypto)(nil),
	}
	file_cryptos_cryptos_proto_msgTypes[8].OneofWrappers = []interface{}{}
	file_cryptos_cryptos_proto_msgTypes[9].OneofWrappers = []interface{}{
		(*GetListResponse_Error)(nil),
		(*GetListResponse_Crypto)(nil),
	}
	file_cryptos_cryptos_proto_msgTypes[10].OneofWrappers = []interface{}{
		(*DeleteResponse_Error)(nil),
		(*DeleteResponse_Crypto)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cryptos_cryptos_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cryptos_cryptos_proto_goTypes,
		DependencyIndexes: file_cryptos_cryptos_proto_depIdxs,
		EnumInfos:         file_cryptos_cryptos_proto_enumTypes,
		MessageInfos:      file_cryptos_cryptos_proto_msgTypes,
	}.Build()
	File_cryptos_cryptos_proto = out.File
	file_cryptos_cryptos_proto_rawDesc = nil
	file_cryptos_cryptos_proto_goTypes = nil
	file_cryptos_cryptos_proto_depIdxs = nil
}
