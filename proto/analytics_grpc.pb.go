package proto

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Client ---

// AnalyticsServiceClient is the client API for AnalyticsService.
type AnalyticsServiceClient interface {
	RecordClick(ctx context.Context, in *RecordClickRequest, opts ...grpc.CallOption) (*RecordClickResponse, error)
	GetStats(ctx context.Context, in *GetStatsRequest, opts ...grpc.CallOption) (*GetStatsResponse, error)
}

type analyticsServiceClient struct {
	cc grpc.ClientConnInterface
}

// NewAnalyticsServiceClient creates a new gRPC client.
func NewAnalyticsServiceClient(cc grpc.ClientConnInterface) AnalyticsServiceClient {
	return &analyticsServiceClient{cc}
}

func (c *analyticsServiceClient) RecordClick(ctx context.Context, in *RecordClickRequest, opts ...grpc.CallOption) (*RecordClickResponse, error) {
	out := new(RecordClickResponse)
	err := c.cc.Invoke(ctx, "/analytics.AnalyticsService/RecordClick", in, out, opts...)
	return out, err
}

func (c *analyticsServiceClient) GetStats(ctx context.Context, in *GetStatsRequest, opts ...grpc.CallOption) (*GetStatsResponse, error) {
	out := new(GetStatsResponse)
	err := c.cc.Invoke(ctx, "/analytics.AnalyticsService/GetStats", in, out, opts...)
	return out, err
}

// --- Server ---

// AnalyticsServiceServer is the server API for AnalyticsService.
type AnalyticsServiceServer interface {
	RecordClick(context.Context, *RecordClickRequest) (*RecordClickResponse, error)
	GetStats(context.Context, *GetStatsRequest) (*GetStatsResponse, error)
}

// UnimplementedAnalyticsServiceServer provides default "not implemented" responses.
type UnimplementedAnalyticsServiceServer struct{}

func (UnimplementedAnalyticsServiceServer) RecordClick(context.Context, *RecordClickRequest) (*RecordClickResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "RecordClick not implemented")
}

func (UnimplementedAnalyticsServiceServer) GetStats(context.Context, *GetStatsRequest) (*GetStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetStats not implemented")
}

// RegisterAnalyticsServiceServer registers the server implementation.
func RegisterAnalyticsServiceServer(s *grpc.Server, srv AnalyticsServiceServer) {
	s.RegisterService(&analyticsServiceDesc, srv)
}

// --- Handlers ---

func recordClickHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RecordClickRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AnalyticsServiceServer).RecordClick(ctx, in)
}

func getStatsHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AnalyticsServiceServer).GetStats(ctx, in)
}

// analyticsServiceDesc describes the gRPC service for registration.
var analyticsServiceDesc = grpc.ServiceDesc{
	ServiceName: "analytics.AnalyticsService",
	HandlerType: (*AnalyticsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "RecordClick", Handler: recordClickHandler},
		{MethodName: "GetStats", Handler: getStatsHandler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/analytics.proto",
}
