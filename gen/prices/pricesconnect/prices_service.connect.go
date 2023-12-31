// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: prices/prices_service.proto

package pricesconnect

import (
	context "context"
	prices "davensi.com/core/gen/prices"
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
	ServiceName = "prices.Service"
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
	ServiceCreateProcedure = "/prices.Service/Create"
	// ServiceUpdateProcedure is the fully-qualified name of the Service's Update RPC.
	ServiceUpdateProcedure = "/prices.Service/Update"
	// ServiceGetProcedure is the fully-qualified name of the Service's Get RPC.
	ServiceGetProcedure = "/prices.Service/Get"
	// ServiceGetListProcedure is the fully-qualified name of the Service's GetList RPC.
	ServiceGetListProcedure = "/prices.Service/GetList"
	// ServiceDeleteProcedure is the fully-qualified name of the Service's Delete RPC.
	ServiceDeleteProcedure = "/prices.Service/Delete"
	// ServiceGetTimeSeriesProcedure is the fully-qualified name of the Service's GetTimeSeries RPC.
	ServiceGetTimeSeriesProcedure = "/prices.Service/GetTimeSeries"
)

// ServiceClient is a client for the prices.Service service.
type ServiceClient interface {
	Create(context.Context, *connect_go.Request[prices.CreateRequest]) (*connect_go.Response[prices.CreateResponse], error)
	Update(context.Context, *connect_go.Request[prices.UpdateRequest]) (*connect_go.Response[prices.UpdateResponse], error)
	Get(context.Context, *connect_go.Request[prices.GetRequest]) (*connect_go.Response[prices.GetResponse], error)
	GetList(context.Context, *connect_go.Request[prices.GetListRequest]) (*connect_go.ServerStreamForClient[prices.GetListResponse], error)
	Delete(context.Context, *connect_go.Request[prices.DeleteRequest]) (*connect_go.Response[prices.DeleteResponse], error)
	GetTimeSeries(context.Context, *connect_go.Request[prices.GetTimeSeriesRequest]) (*connect_go.ServerStreamForClient[prices.GetTimeSeriesResponse], error)
}

// NewServiceClient constructs a client for the prices.Service service. By default, it uses the
// Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) ServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &serviceClient{
		create: connect_go.NewClient[prices.CreateRequest, prices.CreateResponse](
			httpClient,
			baseURL+ServiceCreateProcedure,
			opts...,
		),
		update: connect_go.NewClient[prices.UpdateRequest, prices.UpdateResponse](
			httpClient,
			baseURL+ServiceUpdateProcedure,
			opts...,
		),
		get: connect_go.NewClient[prices.GetRequest, prices.GetResponse](
			httpClient,
			baseURL+ServiceGetProcedure,
			opts...,
		),
		getList: connect_go.NewClient[prices.GetListRequest, prices.GetListResponse](
			httpClient,
			baseURL+ServiceGetListProcedure,
			opts...,
		),
		delete: connect_go.NewClient[prices.DeleteRequest, prices.DeleteResponse](
			httpClient,
			baseURL+ServiceDeleteProcedure,
			opts...,
		),
		getTimeSeries: connect_go.NewClient[prices.GetTimeSeriesRequest, prices.GetTimeSeriesResponse](
			httpClient,
			baseURL+ServiceGetTimeSeriesProcedure,
			opts...,
		),
	}
}

// serviceClient implements ServiceClient.
type serviceClient struct {
	create        *connect_go.Client[prices.CreateRequest, prices.CreateResponse]
	update        *connect_go.Client[prices.UpdateRequest, prices.UpdateResponse]
	get           *connect_go.Client[prices.GetRequest, prices.GetResponse]
	getList       *connect_go.Client[prices.GetListRequest, prices.GetListResponse]
	delete        *connect_go.Client[prices.DeleteRequest, prices.DeleteResponse]
	getTimeSeries *connect_go.Client[prices.GetTimeSeriesRequest, prices.GetTimeSeriesResponse]
}

// Create calls prices.Service.Create.
func (c *serviceClient) Create(ctx context.Context, req *connect_go.Request[prices.CreateRequest]) (*connect_go.Response[prices.CreateResponse], error) {
	return c.create.CallUnary(ctx, req)
}

// Update calls prices.Service.Update.
func (c *serviceClient) Update(ctx context.Context, req *connect_go.Request[prices.UpdateRequest]) (*connect_go.Response[prices.UpdateResponse], error) {
	return c.update.CallUnary(ctx, req)
}

// Get calls prices.Service.Get.
func (c *serviceClient) Get(ctx context.Context, req *connect_go.Request[prices.GetRequest]) (*connect_go.Response[prices.GetResponse], error) {
	return c.get.CallUnary(ctx, req)
}

// GetList calls prices.Service.GetList.
func (c *serviceClient) GetList(ctx context.Context, req *connect_go.Request[prices.GetListRequest]) (*connect_go.ServerStreamForClient[prices.GetListResponse], error) {
	return c.getList.CallServerStream(ctx, req)
}

// Delete calls prices.Service.Delete.
func (c *serviceClient) Delete(ctx context.Context, req *connect_go.Request[prices.DeleteRequest]) (*connect_go.Response[prices.DeleteResponse], error) {
	return c.delete.CallUnary(ctx, req)
}

// GetTimeSeries calls prices.Service.GetTimeSeries.
func (c *serviceClient) GetTimeSeries(ctx context.Context, req *connect_go.Request[prices.GetTimeSeriesRequest]) (*connect_go.ServerStreamForClient[prices.GetTimeSeriesResponse], error) {
	return c.getTimeSeries.CallServerStream(ctx, req)
}

// ServiceHandler is an implementation of the prices.Service service.
type ServiceHandler interface {
	Create(context.Context, *connect_go.Request[prices.CreateRequest]) (*connect_go.Response[prices.CreateResponse], error)
	Update(context.Context, *connect_go.Request[prices.UpdateRequest]) (*connect_go.Response[prices.UpdateResponse], error)
	Get(context.Context, *connect_go.Request[prices.GetRequest]) (*connect_go.Response[prices.GetResponse], error)
	GetList(context.Context, *connect_go.Request[prices.GetListRequest], *connect_go.ServerStream[prices.GetListResponse]) error
	Delete(context.Context, *connect_go.Request[prices.DeleteRequest]) (*connect_go.Response[prices.DeleteResponse], error)
	GetTimeSeries(context.Context, *connect_go.Request[prices.GetTimeSeriesRequest], *connect_go.ServerStream[prices.GetTimeSeriesResponse]) error
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
	serviceGetTimeSeriesHandler := connect_go.NewServerStreamHandler(
		ServiceGetTimeSeriesProcedure,
		svc.GetTimeSeries,
		opts...,
	)
	return "/prices.Service/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		case ServiceGetTimeSeriesProcedure:
			serviceGetTimeSeriesHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedServiceHandler struct{}

func (UnimplementedServiceHandler) Create(context.Context, *connect_go.Request[prices.CreateRequest]) (*connect_go.Response[prices.CreateResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("prices.Service.Create is not implemented"))
}

func (UnimplementedServiceHandler) Update(context.Context, *connect_go.Request[prices.UpdateRequest]) (*connect_go.Response[prices.UpdateResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("prices.Service.Update is not implemented"))
}

func (UnimplementedServiceHandler) Get(context.Context, *connect_go.Request[prices.GetRequest]) (*connect_go.Response[prices.GetResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("prices.Service.Get is not implemented"))
}

func (UnimplementedServiceHandler) GetList(context.Context, *connect_go.Request[prices.GetListRequest], *connect_go.ServerStream[prices.GetListResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("prices.Service.GetList is not implemented"))
}

func (UnimplementedServiceHandler) Delete(context.Context, *connect_go.Request[prices.DeleteRequest]) (*connect_go.Response[prices.DeleteResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("prices.Service.Delete is not implemented"))
}

func (UnimplementedServiceHandler) GetTimeSeries(context.Context, *connect_go.Request[prices.GetTimeSeriesRequest], *connect_go.ServerStream[prices.GetTimeSeriesResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("prices.Service.GetTimeSeries is not implemented"))
}
