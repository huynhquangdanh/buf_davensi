// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: blockchains/blockchains_service.proto

package blockchainsconnect

import (
	context "context"
	blockchains "davensi.com/core/gen/blockchains"
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
	ServiceName = "blockchains.Service"
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
	ServiceCreateProcedure = "/blockchains.Service/Create"
	// ServiceUpdateProcedure is the fully-qualified name of the Service's Update RPC.
	ServiceUpdateProcedure = "/blockchains.Service/Update"
	// ServiceGetProcedure is the fully-qualified name of the Service's Get RPC.
	ServiceGetProcedure = "/blockchains.Service/Get"
	// ServiceGetListProcedure is the fully-qualified name of the Service's GetList RPC.
	ServiceGetListProcedure = "/blockchains.Service/GetList"
	// ServiceDeleteProcedure is the fully-qualified name of the Service's Delete RPC.
	ServiceDeleteProcedure = "/blockchains.Service/Delete"
	// ServiceSetCryptosProcedure is the fully-qualified name of the Service's SetCryptos RPC.
	ServiceSetCryptosProcedure = "/blockchains.Service/SetCryptos"
	// ServiceAddCryptosProcedure is the fully-qualified name of the Service's AddCryptos RPC.
	ServiceAddCryptosProcedure = "/blockchains.Service/AddCryptos"
	// ServiceUpdateCryptoProcedure is the fully-qualified name of the Service's UpdateCrypto RPC.
	ServiceUpdateCryptoProcedure = "/blockchains.Service/UpdateCrypto"
	// ServiceRemoveCryptosProcedure is the fully-qualified name of the Service's RemoveCryptos RPC.
	ServiceRemoveCryptosProcedure = "/blockchains.Service/RemoveCryptos"
)

// ServiceClient is a client for the blockchains.Service service.
type ServiceClient interface {
	Create(context.Context, *connect_go.Request[blockchains.CreateRequest]) (*connect_go.Response[blockchains.CreateResponse], error)
	Update(context.Context, *connect_go.Request[blockchains.UpdateRequest]) (*connect_go.Response[blockchains.UpdateResponse], error)
	Get(context.Context, *connect_go.Request[blockchains.GetRequest]) (*connect_go.Response[blockchains.GetResponse], error)
	GetList(context.Context, *connect_go.Request[blockchains.GetListRequest]) (*connect_go.ServerStreamForClient[blockchains.GetListResponse], error)
	Delete(context.Context, *connect_go.Request[blockchains.DeleteRequest]) (*connect_go.Response[blockchains.DeleteResponse], error)
	SetCryptos(context.Context, *connect_go.Request[blockchains.SetCryptosRequest]) (*connect_go.Response[blockchains.SetCryptosResponse], error)
	AddCryptos(context.Context, *connect_go.Request[blockchains.AddCryptosRequest]) (*connect_go.Response[blockchains.AddCryptosResponse], error)
	UpdateCrypto(context.Context, *connect_go.Request[blockchains.UpdateCryptoRequest]) (*connect_go.Response[blockchains.UpdateCryptoResponse], error)
	RemoveCryptos(context.Context, *connect_go.Request[blockchains.RemoveCryptosRequest]) (*connect_go.Response[blockchains.RemoveCryptosResponse], error)
}

// NewServiceClient constructs a client for the blockchains.Service service. By default, it uses the
// Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) ServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &serviceClient{
		create: connect_go.NewClient[blockchains.CreateRequest, blockchains.CreateResponse](
			httpClient,
			baseURL+ServiceCreateProcedure,
			opts...,
		),
		update: connect_go.NewClient[blockchains.UpdateRequest, blockchains.UpdateResponse](
			httpClient,
			baseURL+ServiceUpdateProcedure,
			opts...,
		),
		get: connect_go.NewClient[blockchains.GetRequest, blockchains.GetResponse](
			httpClient,
			baseURL+ServiceGetProcedure,
			opts...,
		),
		getList: connect_go.NewClient[blockchains.GetListRequest, blockchains.GetListResponse](
			httpClient,
			baseURL+ServiceGetListProcedure,
			opts...,
		),
		delete: connect_go.NewClient[blockchains.DeleteRequest, blockchains.DeleteResponse](
			httpClient,
			baseURL+ServiceDeleteProcedure,
			opts...,
		),
		setCryptos: connect_go.NewClient[blockchains.SetCryptosRequest, blockchains.SetCryptosResponse](
			httpClient,
			baseURL+ServiceSetCryptosProcedure,
			opts...,
		),
		addCryptos: connect_go.NewClient[blockchains.AddCryptosRequest, blockchains.AddCryptosResponse](
			httpClient,
			baseURL+ServiceAddCryptosProcedure,
			opts...,
		),
		updateCrypto: connect_go.NewClient[blockchains.UpdateCryptoRequest, blockchains.UpdateCryptoResponse](
			httpClient,
			baseURL+ServiceUpdateCryptoProcedure,
			opts...,
		),
		removeCryptos: connect_go.NewClient[blockchains.RemoveCryptosRequest, blockchains.RemoveCryptosResponse](
			httpClient,
			baseURL+ServiceRemoveCryptosProcedure,
			opts...,
		),
	}
}

// serviceClient implements ServiceClient.
type serviceClient struct {
	create        *connect_go.Client[blockchains.CreateRequest, blockchains.CreateResponse]
	update        *connect_go.Client[blockchains.UpdateRequest, blockchains.UpdateResponse]
	get           *connect_go.Client[blockchains.GetRequest, blockchains.GetResponse]
	getList       *connect_go.Client[blockchains.GetListRequest, blockchains.GetListResponse]
	delete        *connect_go.Client[blockchains.DeleteRequest, blockchains.DeleteResponse]
	setCryptos    *connect_go.Client[blockchains.SetCryptosRequest, blockchains.SetCryptosResponse]
	addCryptos    *connect_go.Client[blockchains.AddCryptosRequest, blockchains.AddCryptosResponse]
	updateCrypto  *connect_go.Client[blockchains.UpdateCryptoRequest, blockchains.UpdateCryptoResponse]
	removeCryptos *connect_go.Client[blockchains.RemoveCryptosRequest, blockchains.RemoveCryptosResponse]
}

// Create calls blockchains.Service.Create.
func (c *serviceClient) Create(ctx context.Context, req *connect_go.Request[blockchains.CreateRequest]) (*connect_go.Response[blockchains.CreateResponse], error) {
	return c.create.CallUnary(ctx, req)
}

// Update calls blockchains.Service.Update.
func (c *serviceClient) Update(ctx context.Context, req *connect_go.Request[blockchains.UpdateRequest]) (*connect_go.Response[blockchains.UpdateResponse], error) {
	return c.update.CallUnary(ctx, req)
}

// Get calls blockchains.Service.Get.
func (c *serviceClient) Get(ctx context.Context, req *connect_go.Request[blockchains.GetRequest]) (*connect_go.Response[blockchains.GetResponse], error) {
	return c.get.CallUnary(ctx, req)
}

// GetList calls blockchains.Service.GetList.
func (c *serviceClient) GetList(ctx context.Context, req *connect_go.Request[blockchains.GetListRequest]) (*connect_go.ServerStreamForClient[blockchains.GetListResponse], error) {
	return c.getList.CallServerStream(ctx, req)
}

// Delete calls blockchains.Service.Delete.
func (c *serviceClient) Delete(ctx context.Context, req *connect_go.Request[blockchains.DeleteRequest]) (*connect_go.Response[blockchains.DeleteResponse], error) {
	return c.delete.CallUnary(ctx, req)
}

// SetCryptos calls blockchains.Service.SetCryptos.
func (c *serviceClient) SetCryptos(ctx context.Context, req *connect_go.Request[blockchains.SetCryptosRequest]) (*connect_go.Response[blockchains.SetCryptosResponse], error) {
	return c.setCryptos.CallUnary(ctx, req)
}

// AddCryptos calls blockchains.Service.AddCryptos.
func (c *serviceClient) AddCryptos(ctx context.Context, req *connect_go.Request[blockchains.AddCryptosRequest]) (*connect_go.Response[blockchains.AddCryptosResponse], error) {
	return c.addCryptos.CallUnary(ctx, req)
}

// UpdateCrypto calls blockchains.Service.UpdateCrypto.
func (c *serviceClient) UpdateCrypto(ctx context.Context, req *connect_go.Request[blockchains.UpdateCryptoRequest]) (*connect_go.Response[blockchains.UpdateCryptoResponse], error) {
	return c.updateCrypto.CallUnary(ctx, req)
}

// RemoveCryptos calls blockchains.Service.RemoveCryptos.
func (c *serviceClient) RemoveCryptos(ctx context.Context, req *connect_go.Request[blockchains.RemoveCryptosRequest]) (*connect_go.Response[blockchains.RemoveCryptosResponse], error) {
	return c.removeCryptos.CallUnary(ctx, req)
}

// ServiceHandler is an implementation of the blockchains.Service service.
type ServiceHandler interface {
	Create(context.Context, *connect_go.Request[blockchains.CreateRequest]) (*connect_go.Response[blockchains.CreateResponse], error)
	Update(context.Context, *connect_go.Request[blockchains.UpdateRequest]) (*connect_go.Response[blockchains.UpdateResponse], error)
	Get(context.Context, *connect_go.Request[blockchains.GetRequest]) (*connect_go.Response[blockchains.GetResponse], error)
	GetList(context.Context, *connect_go.Request[blockchains.GetListRequest], *connect_go.ServerStream[blockchains.GetListResponse]) error
	Delete(context.Context, *connect_go.Request[blockchains.DeleteRequest]) (*connect_go.Response[blockchains.DeleteResponse], error)
	SetCryptos(context.Context, *connect_go.Request[blockchains.SetCryptosRequest]) (*connect_go.Response[blockchains.SetCryptosResponse], error)
	AddCryptos(context.Context, *connect_go.Request[blockchains.AddCryptosRequest]) (*connect_go.Response[blockchains.AddCryptosResponse], error)
	UpdateCrypto(context.Context, *connect_go.Request[blockchains.UpdateCryptoRequest]) (*connect_go.Response[blockchains.UpdateCryptoResponse], error)
	RemoveCryptos(context.Context, *connect_go.Request[blockchains.RemoveCryptosRequest]) (*connect_go.Response[blockchains.RemoveCryptosResponse], error)
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
	serviceSetCryptosHandler := connect_go.NewUnaryHandler(
		ServiceSetCryptosProcedure,
		svc.SetCryptos,
		opts...,
	)
	serviceAddCryptosHandler := connect_go.NewUnaryHandler(
		ServiceAddCryptosProcedure,
		svc.AddCryptos,
		opts...,
	)
	serviceUpdateCryptoHandler := connect_go.NewUnaryHandler(
		ServiceUpdateCryptoProcedure,
		svc.UpdateCrypto,
		opts...,
	)
	serviceRemoveCryptosHandler := connect_go.NewUnaryHandler(
		ServiceRemoveCryptosProcedure,
		svc.RemoveCryptos,
		opts...,
	)
	return "/blockchains.Service/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		case ServiceSetCryptosProcedure:
			serviceSetCryptosHandler.ServeHTTP(w, r)
		case ServiceAddCryptosProcedure:
			serviceAddCryptosHandler.ServeHTTP(w, r)
		case ServiceUpdateCryptoProcedure:
			serviceUpdateCryptoHandler.ServeHTTP(w, r)
		case ServiceRemoveCryptosProcedure:
			serviceRemoveCryptosHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedServiceHandler struct{}

func (UnimplementedServiceHandler) Create(context.Context, *connect_go.Request[blockchains.CreateRequest]) (*connect_go.Response[blockchains.CreateResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("blockchains.Service.Create is not implemented"))
}

func (UnimplementedServiceHandler) Update(context.Context, *connect_go.Request[blockchains.UpdateRequest]) (*connect_go.Response[blockchains.UpdateResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("blockchains.Service.Update is not implemented"))
}

func (UnimplementedServiceHandler) Get(context.Context, *connect_go.Request[blockchains.GetRequest]) (*connect_go.Response[blockchains.GetResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("blockchains.Service.Get is not implemented"))
}

func (UnimplementedServiceHandler) GetList(context.Context, *connect_go.Request[blockchains.GetListRequest], *connect_go.ServerStream[blockchains.GetListResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("blockchains.Service.GetList is not implemented"))
}

func (UnimplementedServiceHandler) Delete(context.Context, *connect_go.Request[blockchains.DeleteRequest]) (*connect_go.Response[blockchains.DeleteResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("blockchains.Service.Delete is not implemented"))
}

func (UnimplementedServiceHandler) SetCryptos(context.Context, *connect_go.Request[blockchains.SetCryptosRequest]) (*connect_go.Response[blockchains.SetCryptosResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("blockchains.Service.SetCryptos is not implemented"))
}

func (UnimplementedServiceHandler) AddCryptos(context.Context, *connect_go.Request[blockchains.AddCryptosRequest]) (*connect_go.Response[blockchains.AddCryptosResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("blockchains.Service.AddCryptos is not implemented"))
}

func (UnimplementedServiceHandler) UpdateCrypto(context.Context, *connect_go.Request[blockchains.UpdateCryptoRequest]) (*connect_go.Response[blockchains.UpdateCryptoResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("blockchains.Service.UpdateCrypto is not implemented"))
}

func (UnimplementedServiceHandler) RemoveCryptos(context.Context, *connect_go.Request[blockchains.RemoveCryptosRequest]) (*connect_go.Response[blockchains.RemoveCryptosResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("blockchains.Service.RemoveCryptos is not implemented"))
}
