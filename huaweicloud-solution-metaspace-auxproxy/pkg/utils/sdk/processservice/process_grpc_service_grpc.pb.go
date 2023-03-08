// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package processservice

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ProcessGrpcSdkServiceClient is the client API for ProcessGrpcSdkService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProcessGrpcSdkServiceClient interface {
	// 接收健康检查请求
	OnHealthCheck(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error)
	// 接收游戏会话
	OnStartServerSession(ctx context.Context, in *StartServerSessionRequest, opts ...grpc.CallOption) (*ProcessResponse, error)
	// 结束游戏进程
	OnProcessTerminate(ctx context.Context, in *ProcessTerminateRequest, opts ...grpc.CallOption) (*ProcessResponse, error)
}

type processGrpcSdkServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProcessGrpcSdkServiceClient(cc grpc.ClientConnInterface) ProcessGrpcSdkServiceClient {
	return &processGrpcSdkServiceClient{cc}
}

func (c *processGrpcSdkServiceClient) OnHealthCheck(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error) {
	out := new(HealthCheckResponse)
	err := c.cc.Invoke(ctx, "/processService.ProcessGrpcSdkService/OnHealthCheck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *processGrpcSdkServiceClient) OnStartServerSession(ctx context.Context, in *StartServerSessionRequest, opts ...grpc.CallOption) (*ProcessResponse, error) {
	out := new(ProcessResponse)
	err := c.cc.Invoke(ctx, "/processService.ProcessGrpcSdkService/OnStartServerSession", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *processGrpcSdkServiceClient) OnProcessTerminate(ctx context.Context, in *ProcessTerminateRequest, opts ...grpc.CallOption) (*ProcessResponse, error) {
	out := new(ProcessResponse)
	err := c.cc.Invoke(ctx, "/processService.ProcessGrpcSdkService/OnProcessTerminate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProcessGrpcSdkServiceServer is the server API for ProcessGrpcSdkService service.
// All implementations must embed UnimplementedProcessGrpcSdkServiceServer
// for forward compatibility
type ProcessGrpcSdkServiceServer interface {
	// 接收健康检查请求
	OnHealthCheck(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error)
	// 接收游戏会话
	OnStartServerSession(context.Context, *StartServerSessionRequest) (*ProcessResponse, error)
	// 结束游戏进程
	OnProcessTerminate(context.Context, *ProcessTerminateRequest) (*ProcessResponse, error)
	mustEmbedUnimplementedProcessGrpcSdkServiceServer()
}

// UnimplementedProcessGrpcSdkServiceServer must be embedded to have forward compatible implementations.
type UnimplementedProcessGrpcSdkServiceServer struct {
}

func (UnimplementedProcessGrpcSdkServiceServer) OnHealthCheck(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnHealthCheck not implemented")
}
func (UnimplementedProcessGrpcSdkServiceServer) OnStartServerSession(context.Context, *StartServerSessionRequest) (*ProcessResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnStartServerSession not implemented")
}
func (UnimplementedProcessGrpcSdkServiceServer) OnProcessTerminate(context.Context, *ProcessTerminateRequest) (*ProcessResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnProcessTerminate not implemented")
}
func (UnimplementedProcessGrpcSdkServiceServer) mustEmbedUnimplementedProcessGrpcSdkServiceServer() {}

// UnsafeProcessGrpcSdkServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ProcessGrpcSdkServiceServer will
// result in compilation errors.
type UnsafeProcessGrpcSdkServiceServer interface {
	mustEmbedUnimplementedProcessGrpcSdkServiceServer()
}

func RegisterProcessGrpcSdkServiceServer(s grpc.ServiceRegistrar, srv ProcessGrpcSdkServiceServer) {
	s.RegisterService(&ProcessGrpcSdkService_ServiceDesc, srv)
}

func _ProcessGrpcSdkService_OnHealthCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProcessGrpcSdkServiceServer).OnHealthCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/processService.ProcessGrpcSdkService/OnHealthCheck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProcessGrpcSdkServiceServer).OnHealthCheck(ctx, req.(*HealthCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProcessGrpcSdkService_OnStartServerSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartServerSessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProcessGrpcSdkServiceServer).OnStartServerSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/processService.ProcessGrpcSdkService/OnStartServerSession",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProcessGrpcSdkServiceServer).OnStartServerSession(ctx, req.(*StartServerSessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProcessGrpcSdkService_OnProcessTerminate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProcessTerminateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProcessGrpcSdkServiceServer).OnProcessTerminate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/processService.ProcessGrpcSdkService/OnProcessTerminate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProcessGrpcSdkServiceServer).OnProcessTerminate(ctx, req.(*ProcessTerminateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ProcessGrpcSdkService_ServiceDesc is the grpc.ServiceDesc for ProcessGrpcSdkService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ProcessGrpcSdkService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "processService.ProcessGrpcSdkService",
	HandlerType: (*ProcessGrpcSdkServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "OnHealthCheck",
			Handler:    _ProcessGrpcSdkService_OnHealthCheck_Handler,
		},
		{
			MethodName: "OnStartServerSession",
			Handler:    _ProcessGrpcSdkService_OnStartServerSession_Handler,
		},
		{
			MethodName: "OnProcessTerminate",
			Handler:    _ProcessGrpcSdkService_OnProcessTerminate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "process_grpc_service.proto",
}
