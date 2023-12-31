// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: userids/userids_service.proto

package useridsconnect

import (
	context "context"
	userids "davensi.com/core/gen/userids"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion0_1_0

const (
	// ServiceName is the fully-qualified name of the Service service.
	ServiceName = "userids.Service"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ServiceCreateProcedure is the fully-qualified name of the Service's Create RPC.
	ServiceCreateProcedure = "/userids.Service/Create"
	// ServiceUpdateProcedure is the fully-qualified name of the Service's Update RPC.
	ServiceUpdateProcedure = "/userids.Service/Update"
	// ServiceGetProcedure is the fully-qualified name of the Service's Get RPC.
	ServiceGetProcedure = "/userids.Service/Get"
	// ServiceGetListProcedure is the fully-qualified name of the Service's GetList RPC.
	ServiceGetListProcedure = "/userids.Service/GetList"
	// ServiceDeleteProcedure is the fully-qualified name of the Service's Delete RPC.
	ServiceDeleteProcedure = "/userids.Service/Delete"
	// ServiceSetAddressesProcedure is the fully-qualified name of the Service's SetAddresses RPC.
	ServiceSetAddressesProcedure = "/userids.Service/SetAddresses"
	// ServiceAddAddressesProcedure is the fully-qualified name of the Service's AddAddresses RPC.
	ServiceAddAddressesProcedure = "/userids.Service/AddAddresses"
	// ServiceUpdateAddressProcedure is the fully-qualified name of the Service's UpdateAddress RPC.
	ServiceUpdateAddressProcedure = "/userids.Service/UpdateAddress"
	// ServiceRemoveAddressesProcedure is the fully-qualified name of the Service's RemoveAddresses RPC.
	ServiceRemoveAddressesProcedure = "/userids.Service/RemoveAddresses"
	// ServiceSetContactsProcedure is the fully-qualified name of the Service's SetContacts RPC.
	ServiceSetContactsProcedure = "/userids.Service/SetContacts"
	// ServiceAddContactsProcedure is the fully-qualified name of the Service's AddContacts RPC.
	ServiceAddContactsProcedure = "/userids.Service/AddContacts"
	// ServiceUpdateContactProcedure is the fully-qualified name of the Service's UpdateContact RPC.
	ServiceUpdateContactProcedure = "/userids.Service/UpdateContact"
	// ServiceRemoveContactsProcedure is the fully-qualified name of the Service's RemoveContacts RPC.
	ServiceRemoveContactsProcedure = "/userids.Service/RemoveContacts"
	// ServiceSetIncomesProcedure is the fully-qualified name of the Service's SetIncomes RPC.
	ServiceSetIncomesProcedure = "/userids.Service/SetIncomes"
	// ServiceAddIncomesProcedure is the fully-qualified name of the Service's AddIncomes RPC.
	ServiceAddIncomesProcedure = "/userids.Service/AddIncomes"
	// ServiceUpdateIncomeProcedure is the fully-qualified name of the Service's UpdateIncome RPC.
	ServiceUpdateIncomeProcedure = "/userids.Service/UpdateIncome"
	// ServiceRemoveIncomesProcedure is the fully-qualified name of the Service's RemoveIncomes RPC.
	ServiceRemoveIncomesProcedure = "/userids.Service/RemoveIncomes"
)

// ServiceClient is a client for the userids.Service service.
type ServiceClient interface {
	Create(context.Context, *connect_go.Request[userids.CreateRequest]) (*connect_go.Response[userids.CreateResponse], error)
	Update(context.Context, *connect_go.Request[userids.UpdateRequest]) (*connect_go.Response[userids.UpdateResponse], error)
	Get(context.Context, *connect_go.Request[userids.GetRequest]) (*connect_go.Response[userids.GetResponse], error)
	GetList(context.Context, *connect_go.Request[userids.GetListRequest]) (*connect_go.ServerStreamForClient[userids.GetListResponse], error)
	Delete(context.Context, *connect_go.Request[userids.DeleteRequest]) (*connect_go.Response[userids.DeleteResponse], error)
	SetAddresses(context.Context, *connect_go.Request[userids.SetAddressesRequest]) (*connect_go.Response[userids.SetAddressesResponse], error)
	AddAddresses(context.Context, *connect_go.Request[userids.AddAddressesRequest]) (*connect_go.Response[userids.AddAddressesResponse], error)
	UpdateAddress(context.Context, *connect_go.Request[userids.UpdateAddressRequest]) (*connect_go.Response[userids.UpdateAddressResponse], error)
	RemoveAddresses(context.Context, *connect_go.Request[userids.RemoveAddressesRequest]) (*connect_go.Response[userids.RemoveAddressesResponse], error)
	SetContacts(context.Context, *connect_go.Request[userids.SetContactsRequest]) (*connect_go.Response[userids.SetContactsResponse], error)
	AddContacts(context.Context, *connect_go.Request[userids.AddContactsRequest]) (*connect_go.Response[userids.AddContactsResponse], error)
	UpdateContact(context.Context, *connect_go.Request[userids.UpdateContactRequest]) (*connect_go.Response[userids.UpdateContactResponse], error)
	RemoveContacts(context.Context, *connect_go.Request[userids.RemoveContactsRequest]) (*connect_go.Response[userids.RemoveContactsResponse], error)
	SetIncomes(context.Context, *connect_go.Request[userids.SetIncomesRequest]) (*connect_go.Response[userids.SetIncomesResponse], error)
	AddIncomes(context.Context, *connect_go.Request[userids.AddIncomesRequest]) (*connect_go.Response[userids.AddIncomesResponse], error)
	UpdateIncome(context.Context, *connect_go.Request[userids.UpdateIncomeRequest]) (*connect_go.Response[userids.UpdateIncomeResponse], error)
	RemoveIncomes(context.Context, *connect_go.Request[userids.RemoveIncomesRequest]) (*connect_go.Response[userids.RemoveIncomesResponse], error)
}

// NewServiceClient constructs a client for the userids.Service service. By default, it uses the
// Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) ServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &serviceClient{
		create: connect_go.NewClient[userids.CreateRequest, userids.CreateResponse](
			httpClient,
			baseURL+ServiceCreateProcedure,
			opts...,
		),
		update: connect_go.NewClient[userids.UpdateRequest, userids.UpdateResponse](
			httpClient,
			baseURL+ServiceUpdateProcedure,
			opts...,
		),
		get: connect_go.NewClient[userids.GetRequest, userids.GetResponse](
			httpClient,
			baseURL+ServiceGetProcedure,
			opts...,
		),
		getList: connect_go.NewClient[userids.GetListRequest, userids.GetListResponse](
			httpClient,
			baseURL+ServiceGetListProcedure,
			opts...,
		),
		delete: connect_go.NewClient[userids.DeleteRequest, userids.DeleteResponse](
			httpClient,
			baseURL+ServiceDeleteProcedure,
			opts...,
		),
		setAddresses: connect_go.NewClient[userids.SetAddressesRequest, userids.SetAddressesResponse](
			httpClient,
			baseURL+ServiceSetAddressesProcedure,
			opts...,
		),
		addAddresses: connect_go.NewClient[userids.AddAddressesRequest, userids.AddAddressesResponse](
			httpClient,
			baseURL+ServiceAddAddressesProcedure,
			opts...,
		),
		updateAddress: connect_go.NewClient[userids.UpdateAddressRequest, userids.UpdateAddressResponse](
			httpClient,
			baseURL+ServiceUpdateAddressProcedure,
			opts...,
		),
		removeAddresses: connect_go.NewClient[userids.RemoveAddressesRequest, userids.RemoveAddressesResponse](
			httpClient,
			baseURL+ServiceRemoveAddressesProcedure,
			opts...,
		),
		setContacts: connect_go.NewClient[userids.SetContactsRequest, userids.SetContactsResponse](
			httpClient,
			baseURL+ServiceSetContactsProcedure,
			opts...,
		),
		addContacts: connect_go.NewClient[userids.AddContactsRequest, userids.AddContactsResponse](
			httpClient,
			baseURL+ServiceAddContactsProcedure,
			opts...,
		),
		updateContact: connect_go.NewClient[userids.UpdateContactRequest, userids.UpdateContactResponse](
			httpClient,
			baseURL+ServiceUpdateContactProcedure,
			opts...,
		),
		removeContacts: connect_go.NewClient[userids.RemoveContactsRequest, userids.RemoveContactsResponse](
			httpClient,
			baseURL+ServiceRemoveContactsProcedure,
			opts...,
		),
		setIncomes: connect_go.NewClient[userids.SetIncomesRequest, userids.SetIncomesResponse](
			httpClient,
			baseURL+ServiceSetIncomesProcedure,
			opts...,
		),
		addIncomes: connect_go.NewClient[userids.AddIncomesRequest, userids.AddIncomesResponse](
			httpClient,
			baseURL+ServiceAddIncomesProcedure,
			opts...,
		),
		updateIncome: connect_go.NewClient[userids.UpdateIncomeRequest, userids.UpdateIncomeResponse](
			httpClient,
			baseURL+ServiceUpdateIncomeProcedure,
			opts...,
		),
		removeIncomes: connect_go.NewClient[userids.RemoveIncomesRequest, userids.RemoveIncomesResponse](
			httpClient,
			baseURL+ServiceRemoveIncomesProcedure,
			opts...,
		),
	}
}

// serviceClient implements ServiceClient.
type serviceClient struct {
	create          *connect_go.Client[userids.CreateRequest, userids.CreateResponse]
	update          *connect_go.Client[userids.UpdateRequest, userids.UpdateResponse]
	get             *connect_go.Client[userids.GetRequest, userids.GetResponse]
	getList         *connect_go.Client[userids.GetListRequest, userids.GetListResponse]
	delete          *connect_go.Client[userids.DeleteRequest, userids.DeleteResponse]
	setAddresses    *connect_go.Client[userids.SetAddressesRequest, userids.SetAddressesResponse]
	addAddresses    *connect_go.Client[userids.AddAddressesRequest, userids.AddAddressesResponse]
	updateAddress   *connect_go.Client[userids.UpdateAddressRequest, userids.UpdateAddressResponse]
	removeAddresses *connect_go.Client[userids.RemoveAddressesRequest, userids.RemoveAddressesResponse]
	setContacts     *connect_go.Client[userids.SetContactsRequest, userids.SetContactsResponse]
	addContacts     *connect_go.Client[userids.AddContactsRequest, userids.AddContactsResponse]
	updateContact   *connect_go.Client[userids.UpdateContactRequest, userids.UpdateContactResponse]
	removeContacts  *connect_go.Client[userids.RemoveContactsRequest, userids.RemoveContactsResponse]
	setIncomes      *connect_go.Client[userids.SetIncomesRequest, userids.SetIncomesResponse]
	addIncomes      *connect_go.Client[userids.AddIncomesRequest, userids.AddIncomesResponse]
	updateIncome    *connect_go.Client[userids.UpdateIncomeRequest, userids.UpdateIncomeResponse]
	removeIncomes   *connect_go.Client[userids.RemoveIncomesRequest, userids.RemoveIncomesResponse]
}

// Create calls userids.Service.Create.
func (c *serviceClient) Create(ctx context.Context, req *connect_go.Request[userids.CreateRequest]) (*connect_go.Response[userids.CreateResponse], error) {
	return c.create.CallUnary(ctx, req)
}

// Update calls userids.Service.Update.
func (c *serviceClient) Update(ctx context.Context, req *connect_go.Request[userids.UpdateRequest]) (*connect_go.Response[userids.UpdateResponse], error) {
	return c.update.CallUnary(ctx, req)
}

// Get calls userids.Service.Get.
func (c *serviceClient) Get(ctx context.Context, req *connect_go.Request[userids.GetRequest]) (*connect_go.Response[userids.GetResponse], error) {
	return c.get.CallUnary(ctx, req)
}

// GetList calls userids.Service.GetList.
func (c *serviceClient) GetList(ctx context.Context, req *connect_go.Request[userids.GetListRequest]) (*connect_go.ServerStreamForClient[userids.GetListResponse], error) {
	return c.getList.CallServerStream(ctx, req)
}

// Delete calls userids.Service.Delete.
func (c *serviceClient) Delete(ctx context.Context, req *connect_go.Request[userids.DeleteRequest]) (*connect_go.Response[userids.DeleteResponse], error) {
	return c.delete.CallUnary(ctx, req)
}

// SetAddresses calls userids.Service.SetAddresses.
func (c *serviceClient) SetAddresses(ctx context.Context, req *connect_go.Request[userids.SetAddressesRequest]) (*connect_go.Response[userids.SetAddressesResponse], error) {
	return c.setAddresses.CallUnary(ctx, req)
}

// AddAddresses calls userids.Service.AddAddresses.
func (c *serviceClient) AddAddresses(ctx context.Context, req *connect_go.Request[userids.AddAddressesRequest]) (*connect_go.Response[userids.AddAddressesResponse], error) {
	return c.addAddresses.CallUnary(ctx, req)
}

// UpdateAddress calls userids.Service.UpdateAddress.
func (c *serviceClient) UpdateAddress(ctx context.Context, req *connect_go.Request[userids.UpdateAddressRequest]) (*connect_go.Response[userids.UpdateAddressResponse], error) {
	return c.updateAddress.CallUnary(ctx, req)
}

// RemoveAddresses calls userids.Service.RemoveAddresses.
func (c *serviceClient) RemoveAddresses(ctx context.Context, req *connect_go.Request[userids.RemoveAddressesRequest]) (*connect_go.Response[userids.RemoveAddressesResponse], error) {
	return c.removeAddresses.CallUnary(ctx, req)
}

// SetContacts calls userids.Service.SetContacts.
func (c *serviceClient) SetContacts(ctx context.Context, req *connect_go.Request[userids.SetContactsRequest]) (*connect_go.Response[userids.SetContactsResponse], error) {
	return c.setContacts.CallUnary(ctx, req)
}

// AddContacts calls userids.Service.AddContacts.
func (c *serviceClient) AddContacts(ctx context.Context, req *connect_go.Request[userids.AddContactsRequest]) (*connect_go.Response[userids.AddContactsResponse], error) {
	return c.addContacts.CallUnary(ctx, req)
}

// UpdateContact calls userids.Service.UpdateContact.
func (c *serviceClient) UpdateContact(ctx context.Context, req *connect_go.Request[userids.UpdateContactRequest]) (*connect_go.Response[userids.UpdateContactResponse], error) {
	return c.updateContact.CallUnary(ctx, req)
}

// RemoveContacts calls userids.Service.RemoveContacts.
func (c *serviceClient) RemoveContacts(ctx context.Context, req *connect_go.Request[userids.RemoveContactsRequest]) (*connect_go.Response[userids.RemoveContactsResponse], error) {
	return c.removeContacts.CallUnary(ctx, req)
}

// SetIncomes calls userids.Service.SetIncomes.
func (c *serviceClient) SetIncomes(ctx context.Context, req *connect_go.Request[userids.SetIncomesRequest]) (*connect_go.Response[userids.SetIncomesResponse], error) {
	return c.setIncomes.CallUnary(ctx, req)
}

// AddIncomes calls userids.Service.AddIncomes.
func (c *serviceClient) AddIncomes(ctx context.Context, req *connect_go.Request[userids.AddIncomesRequest]) (*connect_go.Response[userids.AddIncomesResponse], error) {
	return c.addIncomes.CallUnary(ctx, req)
}

// UpdateIncome calls userids.Service.UpdateIncome.
func (c *serviceClient) UpdateIncome(ctx context.Context, req *connect_go.Request[userids.UpdateIncomeRequest]) (*connect_go.Response[userids.UpdateIncomeResponse], error) {
	return c.updateIncome.CallUnary(ctx, req)
}

// RemoveIncomes calls userids.Service.RemoveIncomes.
func (c *serviceClient) RemoveIncomes(ctx context.Context, req *connect_go.Request[userids.RemoveIncomesRequest]) (*connect_go.Response[userids.RemoveIncomesResponse], error) {
	return c.removeIncomes.CallUnary(ctx, req)
}

// ServiceHandler is an implementation of the userids.Service service.
type ServiceHandler interface {
	Create(context.Context, *connect_go.Request[userids.CreateRequest]) (*connect_go.Response[userids.CreateResponse], error)
	Update(context.Context, *connect_go.Request[userids.UpdateRequest]) (*connect_go.Response[userids.UpdateResponse], error)
	Get(context.Context, *connect_go.Request[userids.GetRequest]) (*connect_go.Response[userids.GetResponse], error)
	GetList(context.Context, *connect_go.Request[userids.GetListRequest], *connect_go.ServerStream[userids.GetListResponse]) error
	Delete(context.Context, *connect_go.Request[userids.DeleteRequest]) (*connect_go.Response[userids.DeleteResponse], error)
	SetAddresses(context.Context, *connect_go.Request[userids.SetAddressesRequest]) (*connect_go.Response[userids.SetAddressesResponse], error)
	AddAddresses(context.Context, *connect_go.Request[userids.AddAddressesRequest]) (*connect_go.Response[userids.AddAddressesResponse], error)
	UpdateAddress(context.Context, *connect_go.Request[userids.UpdateAddressRequest]) (*connect_go.Response[userids.UpdateAddressResponse], error)
	RemoveAddresses(context.Context, *connect_go.Request[userids.RemoveAddressesRequest]) (*connect_go.Response[userids.RemoveAddressesResponse], error)
	SetContacts(context.Context, *connect_go.Request[userids.SetContactsRequest]) (*connect_go.Response[userids.SetContactsResponse], error)
	AddContacts(context.Context, *connect_go.Request[userids.AddContactsRequest]) (*connect_go.Response[userids.AddContactsResponse], error)
	UpdateContact(context.Context, *connect_go.Request[userids.UpdateContactRequest]) (*connect_go.Response[userids.UpdateContactResponse], error)
	RemoveContacts(context.Context, *connect_go.Request[userids.RemoveContactsRequest]) (*connect_go.Response[userids.RemoveContactsResponse], error)
	SetIncomes(context.Context, *connect_go.Request[userids.SetIncomesRequest]) (*connect_go.Response[userids.SetIncomesResponse], error)
	AddIncomes(context.Context, *connect_go.Request[userids.AddIncomesRequest]) (*connect_go.Response[userids.AddIncomesResponse], error)
	UpdateIncome(context.Context, *connect_go.Request[userids.UpdateIncomeRequest]) (*connect_go.Response[userids.UpdateIncomeResponse], error)
	RemoveIncomes(context.Context, *connect_go.Request[userids.RemoveIncomesRequest]) (*connect_go.Response[userids.RemoveIncomesResponse], error)
}

// NewServiceHandler builds an HTTP handler from the service implementation. It returns the path on
// which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewServiceHandler(svc ServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	serviceCreateHandler := connect_go.NewUnaryHandler(
		ServiceCreateProcedure,
		svc.Create,
		opts...,
	)
	serviceUpdateHandler := connect_go.NewUnaryHandler(
		ServiceUpdateProcedure,
		svc.Update,
		opts...,
	)
	serviceGetHandler := connect_go.NewUnaryHandler(
		ServiceGetProcedure,
		svc.Get,
		opts...,
	)
	serviceGetListHandler := connect_go.NewServerStreamHandler(
		ServiceGetListProcedure,
		svc.GetList,
		opts...,
	)
	serviceDeleteHandler := connect_go.NewUnaryHandler(
		ServiceDeleteProcedure,
		svc.Delete,
		opts...,
	)
	serviceSetAddressesHandler := connect_go.NewUnaryHandler(
		ServiceSetAddressesProcedure,
		svc.SetAddresses,
		opts...,
	)
	serviceAddAddressesHandler := connect_go.NewUnaryHandler(
		ServiceAddAddressesProcedure,
		svc.AddAddresses,
		opts...,
	)
	serviceUpdateAddressHandler := connect_go.NewUnaryHandler(
		ServiceUpdateAddressProcedure,
		svc.UpdateAddress,
		opts...,
	)
	serviceRemoveAddressesHandler := connect_go.NewUnaryHandler(
		ServiceRemoveAddressesProcedure,
		svc.RemoveAddresses,
		opts...,
	)
	serviceSetContactsHandler := connect_go.NewUnaryHandler(
		ServiceSetContactsProcedure,
		svc.SetContacts,
		opts...,
	)
	serviceAddContactsHandler := connect_go.NewUnaryHandler(
		ServiceAddContactsProcedure,
		svc.AddContacts,
		opts...,
	)
	serviceUpdateContactHandler := connect_go.NewUnaryHandler(
		ServiceUpdateContactProcedure,
		svc.UpdateContact,
		opts...,
	)
	serviceRemoveContactsHandler := connect_go.NewUnaryHandler(
		ServiceRemoveContactsProcedure,
		svc.RemoveContacts,
		opts...,
	)
	serviceSetIncomesHandler := connect_go.NewUnaryHandler(
		ServiceSetIncomesProcedure,
		svc.SetIncomes,
		opts...,
	)
	serviceAddIncomesHandler := connect_go.NewUnaryHandler(
		ServiceAddIncomesProcedure,
		svc.AddIncomes,
		opts...,
	)
	serviceUpdateIncomeHandler := connect_go.NewUnaryHandler(
		ServiceUpdateIncomeProcedure,
		svc.UpdateIncome,
		opts...,
	)
	serviceRemoveIncomesHandler := connect_go.NewUnaryHandler(
		ServiceRemoveIncomesProcedure,
		svc.RemoveIncomes,
		opts...,
	)
	return "/userids.Service/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ServiceCreateProcedure:
			serviceCreateHandler.ServeHTTP(w, r)
		case ServiceUpdateProcedure:
			serviceUpdateHandler.ServeHTTP(w, r)
		case ServiceGetProcedure:
			serviceGetHandler.ServeHTTP(w, r)
		case ServiceGetListProcedure:
			serviceGetListHandler.ServeHTTP(w, r)
		case ServiceDeleteProcedure:
			serviceDeleteHandler.ServeHTTP(w, r)
		case ServiceSetAddressesProcedure:
			serviceSetAddressesHandler.ServeHTTP(w, r)
		case ServiceAddAddressesProcedure:
			serviceAddAddressesHandler.ServeHTTP(w, r)
		case ServiceUpdateAddressProcedure:
			serviceUpdateAddressHandler.ServeHTTP(w, r)
		case ServiceRemoveAddressesProcedure:
			serviceRemoveAddressesHandler.ServeHTTP(w, r)
		case ServiceSetContactsProcedure:
			serviceSetContactsHandler.ServeHTTP(w, r)
		case ServiceAddContactsProcedure:
			serviceAddContactsHandler.ServeHTTP(w, r)
		case ServiceUpdateContactProcedure:
			serviceUpdateContactHandler.ServeHTTP(w, r)
		case ServiceRemoveContactsProcedure:
			serviceRemoveContactsHandler.ServeHTTP(w, r)
		case ServiceSetIncomesProcedure:
			serviceSetIncomesHandler.ServeHTTP(w, r)
		case ServiceAddIncomesProcedure:
			serviceAddIncomesHandler.ServeHTTP(w, r)
		case ServiceUpdateIncomeProcedure:
			serviceUpdateIncomeHandler.ServeHTTP(w, r)
		case ServiceRemoveIncomesProcedure:
			serviceRemoveIncomesHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedServiceHandler struct{}

func (UnimplementedServiceHandler) Create(context.Context, *connect_go.Request[userids.CreateRequest]) (*connect_go.Response[userids.CreateResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.Create is not implemented"))
}

func (UnimplementedServiceHandler) Update(context.Context, *connect_go.Request[userids.UpdateRequest]) (*connect_go.Response[userids.UpdateResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.Update is not implemented"))
}

func (UnimplementedServiceHandler) Get(context.Context, *connect_go.Request[userids.GetRequest]) (*connect_go.Response[userids.GetResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.Get is not implemented"))
}

func (UnimplementedServiceHandler) GetList(context.Context, *connect_go.Request[userids.GetListRequest], *connect_go.ServerStream[userids.GetListResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.GetList is not implemented"))
}

func (UnimplementedServiceHandler) Delete(context.Context, *connect_go.Request[userids.DeleteRequest]) (*connect_go.Response[userids.DeleteResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.Delete is not implemented"))
}

func (UnimplementedServiceHandler) SetAddresses(context.Context, *connect_go.Request[userids.SetAddressesRequest]) (*connect_go.Response[userids.SetAddressesResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.SetAddresses is not implemented"))
}

func (UnimplementedServiceHandler) AddAddresses(context.Context, *connect_go.Request[userids.AddAddressesRequest]) (*connect_go.Response[userids.AddAddressesResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.AddAddresses is not implemented"))
}

func (UnimplementedServiceHandler) UpdateAddress(context.Context, *connect_go.Request[userids.UpdateAddressRequest]) (*connect_go.Response[userids.UpdateAddressResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.UpdateAddress is not implemented"))
}

func (UnimplementedServiceHandler) RemoveAddresses(context.Context, *connect_go.Request[userids.RemoveAddressesRequest]) (*connect_go.Response[userids.RemoveAddressesResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.RemoveAddresses is not implemented"))
}

func (UnimplementedServiceHandler) SetContacts(context.Context, *connect_go.Request[userids.SetContactsRequest]) (*connect_go.Response[userids.SetContactsResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.SetContacts is not implemented"))
}

func (UnimplementedServiceHandler) AddContacts(context.Context, *connect_go.Request[userids.AddContactsRequest]) (*connect_go.Response[userids.AddContactsResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.AddContacts is not implemented"))
}

func (UnimplementedServiceHandler) UpdateContact(context.Context, *connect_go.Request[userids.UpdateContactRequest]) (*connect_go.Response[userids.UpdateContactResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.UpdateContact is not implemented"))
}

func (UnimplementedServiceHandler) RemoveContacts(context.Context, *connect_go.Request[userids.RemoveContactsRequest]) (*connect_go.Response[userids.RemoveContactsResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.RemoveContacts is not implemented"))
}

func (UnimplementedServiceHandler) SetIncomes(context.Context, *connect_go.Request[userids.SetIncomesRequest]) (*connect_go.Response[userids.SetIncomesResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.SetIncomes is not implemented"))
}

func (UnimplementedServiceHandler) AddIncomes(context.Context, *connect_go.Request[userids.AddIncomesRequest]) (*connect_go.Response[userids.AddIncomesResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.AddIncomes is not implemented"))
}

func (UnimplementedServiceHandler) UpdateIncome(context.Context, *connect_go.Request[userids.UpdateIncomeRequest]) (*connect_go.Response[userids.UpdateIncomeResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.UpdateIncome is not implemented"))
}

func (UnimplementedServiceHandler) RemoveIncomes(context.Context, *connect_go.Request[userids.RemoveIncomesRequest]) (*connect_go.Response[userids.RemoveIncomesResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userids.Service.RemoveIncomes is not implemented"))
}
