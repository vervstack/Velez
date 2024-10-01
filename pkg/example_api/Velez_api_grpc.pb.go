// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v4.25.1
// source: grpc/Velez_api.proto

package example_api

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	VelezAPI_Version_FullMethodName = "/velez_api.velezAPI/Version"
)

// VelezAPIClient is the client API for VelezAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VelezAPIClient interface {
	Version(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
}

type velezAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewVelezAPIClient(cc grpc.ClientConnInterface) VelezAPIClient {
	return &velezAPIClient{cc}
}

func (c *velezAPIClient) Version(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, VelezAPI_Version_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VelezAPIServer is the server API for VelezAPI service.
// All implementations must embed UnimplementedVelezAPIServer
// for forward compatibility.
type VelezAPIServer interface {
	Version(context.Context, *PingRequest) (*PingResponse, error)
	mustEmbedUnimplementedVelezAPIServer()
}

// UnimplementedVelezAPIServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedVelezAPIServer struct{}

func (UnimplementedVelezAPIServer) Version(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Version not implemented")
}
func (UnimplementedVelezAPIServer) mustEmbedUnimplementedVelezAPIServer() {}
func (UnimplementedVelezAPIServer) testEmbeddedByValue()                  {}

// UnsafeVelezAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VelezAPIServer will
// result in compilation errors.
type UnsafeVelezAPIServer interface {
	mustEmbedUnimplementedVelezAPIServer()
}

func RegisterVelezAPIServer(s grpc.ServiceRegistrar, srv VelezAPIServer) {
	// If the following call pancis, it indicates UnimplementedVelezAPIServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&VelezAPI_ServiceDesc, srv)
}

func _VelezAPI_Version_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VelezAPIServer).Version(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: VelezAPI_Version_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VelezAPIServer).Version(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// VelezAPI_ServiceDesc is the grpc.ServiceDesc for VelezAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var VelezAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "velez_api.velezAPI",
	HandlerType: (*VelezAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Version",
			Handler:    _VelezAPI_Version_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/Velez_api.proto",
}
