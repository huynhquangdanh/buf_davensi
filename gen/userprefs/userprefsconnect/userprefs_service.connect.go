// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: userprefs/userprefs_service.proto

package userprefsconnect

import (
	context "context"
	userprefs "davensi.com/core/gen/userprefs"
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
	ServiceName = "userprefs.Service"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ServiceSetProcedure is the fully-qualified name of the Service's Set RPC.
	ServiceSetProcedure = "/userprefs.Service/Set"
	// ServiceRemoveProcedure is the fully-qualified name of the Service's Remove RPC.
	ServiceRemoveProcedure = "/userprefs.Service/Remove"
	// ServiceGetProcedure is the fully-qualified name of the Service's Get RPC.
	ServiceGetProcedure = "/userprefs.Service/Get"
	// ServiceGetListProcedure is the fully-qualified name of the Service's GetList RPC.
	ServiceGetListProcedure = "/userprefs.Service/GetList"
	// ServiceResetProcedure is the fully-qualified name of the Service's Reset RPC.
	ServiceResetProcedure = "/userprefs.Service/Reset"
)

// ServiceClient is a client for the userprefs.Service service.
type ServiceClient interface {
	// Set: before upsert, please validate the userprefs.key in userprefs_default.key
	Set(context.Context, *connect_go.Request[userprefs.SetRequest]) (*connect_go.Response[userprefs.SetResponse], error)
	Remove(context.Context, *connect_go.Request[userprefs.RemoveRequest]) (*connect_go.Response[userprefs.RemoveResponse], error)
	// Only get item with status = 1. show NOT found if status <> 1.
	// While getting if the item with key is not exist in userprefs => get item key from userprefs_default.
	Get(context.Context, *connect_go.Request[userprefs.GetRequest]) (*connect_go.Response[userprefs.GetResponse], error)
	GetList(context.Context, *connect_go.Request[userprefs.GetListRequest]) (*connect_go.ServerStreamForClient[userprefs.GetListResponse], error)
	Reset(context.Context, *connect_go.Request[userprefs.ResetRequest]) (*connect_go.ServerStreamForClient[userprefs.ResetResponse], error)
}

// NewServiceClient constructs a client for the userprefs.Service service. By default, it uses the
// Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) ServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &serviceClient{
		set: connect_go.NewClient[userprefs.SetRequest, userprefs.SetResponse](
			httpClient,
			baseURL+ServiceSetProcedure,
			opts...,
		),
		remove: connect_go.NewClient[userprefs.RemoveRequest, userprefs.RemoveResponse](
			httpClient,
			baseURL+ServiceRemoveProcedure,
			opts...,
		),
		get: connect_go.NewClient[userprefs.GetRequest, userprefs.GetResponse](
			httpClient,
			baseURL+ServiceGetProcedure,
			opts...,
		),
		getList: connect_go.NewClient[userprefs.GetListRequest, userprefs.GetListResponse](
			httpClient,
			baseURL+ServiceGetListProcedure,
			opts...,
		),
		reset: connect_go.NewClient[userprefs.ResetRequest, userprefs.ResetResponse](
			httpClient,
			baseURL+ServiceResetProcedure,
			opts...,
		),
	}
}

// serviceClient implements ServiceClient.
type serviceClient struct {
	set     *connect_go.Client[userprefs.SetRequest, userprefs.SetResponse]
	remove  *connect_go.Client[userprefs.RemoveRequest, userprefs.RemoveResponse]
	get     *connect_go.Client[userprefs.GetRequest, userprefs.GetResponse]
	getList *connect_go.Client[userprefs.GetListRequest, userprefs.GetListResponse]
	reset   *connect_go.Client[userprefs.ResetRequest, userprefs.ResetResponse]
}

// Set calls userprefs.Service.Set.
func (c *serviceClient) Set(ctx context.Context, req *connect_go.Request[userprefs.SetRequest]) (*connect_go.Response[userprefs.SetResponse], error) {
	return c.set.CallUnary(ctx, req)
}

// Remove calls userprefs.Service.Remove.
func (c *serviceClient) Remove(ctx context.Context, req *connect_go.Request[userprefs.RemoveRequest]) (*connect_go.Response[userprefs.RemoveResponse], error) {
	return c.remove.CallUnary(ctx, req)
}

// Get calls userprefs.Service.Get.
func (c *serviceClient) Get(ctx context.Context, req *connect_go.Request[userprefs.GetRequest]) (*connect_go.Response[userprefs.GetResponse], error) {
	return c.get.CallUnary(ctx, req)
}

// GetList calls userprefs.Service.GetList.
func (c *serviceClient) GetList(ctx context.Context, req *connect_go.Request[userprefs.GetListRequest]) (*connect_go.ServerStreamForClient[userprefs.GetListResponse], error) {
	return c.getList.CallServerStream(ctx, req)
}

// Reset calls userprefs.Service.Reset.
func (c *serviceClient) Reset(ctx context.Context, req *connect_go.Request[userprefs.ResetRequest]) (*connect_go.ServerStreamForClient[userprefs.ResetResponse], error) {
	return c.reset.CallServerStream(ctx, req)
}

// ServiceHandler is an implementation of the userprefs.Service service.
type ServiceHandler interface {
	// Set: before upsert, please validate the userprefs.key in userprefs_default.key
	Set(context.Context, *connect_go.Request[userprefs.SetRequest]) (*connect_go.Response[userprefs.SetResponse], error)
	Remove(context.Context, *connect_go.Request[userprefs.RemoveRequest]) (*connect_go.Response[userprefs.RemoveResponse], error)
	// Only get item with status = 1. show NOT found if status <> 1.
	// While getting if the item with key is not exist in userprefs => get item key from userprefs_default.
	Get(context.Context, *connect_go.Request[userprefs.GetRequest]) (*connect_go.Response[userprefs.GetResponse], error)
	GetList(context.Context, *connect_go.Request[userprefs.GetListRequest], *connect_go.ServerStream[userprefs.GetListResponse]) error
	Reset(context.Context, *connect_go.Request[userprefs.ResetRequest], *connect_go.ServerStream[userprefs.ResetResponse]) error
}

// NewServiceHandler builds an HTTP handler from the service implementation. It returns the path on
// which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewServiceHandler(svc ServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	serviceSetHandler := connect_go.NewUnaryHandler(
		ServiceSetProcedure,
		svc.Set,
		opts...,
	)
	serviceRemoveHandler := connect_go.NewUnaryHandler(
		ServiceRemoveProcedure,
		svc.Remove,
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
	serviceResetHandler := connect_go.NewServerStreamHandler(
		ServiceResetProcedure,
		svc.Reset,
		opts...,
	)
	return "/userprefs.Service/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ServiceSetProcedure:
			serviceSetHandler.ServeHTTP(w, r)
		case ServiceRemoveProcedure:
			serviceRemoveHandler.ServeHTTP(w, r)
		case ServiceGetProcedure:
			serviceGetHandler.ServeHTTP(w, r)
		case ServiceGetListProcedure:
			serviceGetListHandler.ServeHTTP(w, r)
		case ServiceResetProcedure:
			serviceResetHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedServiceHandler struct{}

func (UnimplementedServiceHandler) Set(context.Context, *connect_go.Request[userprefs.SetRequest]) (*connect_go.Response[userprefs.SetResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userprefs.Service.Set is not implemented"))
}

func (UnimplementedServiceHandler) Remove(context.Context, *connect_go.Request[userprefs.RemoveRequest]) (*connect_go.Response[userprefs.RemoveResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userprefs.Service.Remove is not implemented"))
}

func (UnimplementedServiceHandler) Get(context.Context, *connect_go.Request[userprefs.GetRequest]) (*connect_go.Response[userprefs.GetResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userprefs.Service.Get is not implemented"))
}

func (UnimplementedServiceHandler) GetList(context.Context, *connect_go.Request[userprefs.GetListRequest], *connect_go.ServerStream[userprefs.GetListResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userprefs.Service.GetList is not implemented"))
}

func (UnimplementedServiceHandler) Reset(context.Context, *connect_go.Request[userprefs.ResetRequest], *connect_go.ServerStream[userprefs.ResetResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("userprefs.Service.Reset is not implemented"))
}
