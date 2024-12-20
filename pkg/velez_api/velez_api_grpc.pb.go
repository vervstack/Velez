// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.3
// source: api/grpc/velez_api.proto

package velez_api

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
	VelezAPI_Version_FullMethodName     = "/velez_api.VelezAPI/Version"
	VelezAPI_CreateSmerd_FullMethodName = "/velez_api.VelezAPI/CreateSmerd"
	VelezAPI_ListSmerds_FullMethodName  = "/velez_api.VelezAPI/ListSmerds"
	VelezAPI_DropSmerd_FullMethodName   = "/velez_api.VelezAPI/DropSmerd"
	VelezAPI_GetHardware_FullMethodName = "/velez_api.VelezAPI/GetHardware"
	VelezAPI_FetchConfig_FullMethodName = "/velez_api.VelezAPI/FetchConfig"
)

// VelezAPIClient is the client API for VelezAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VelezAPIClient interface {
	Version(ctx context.Context, in *Version_Request, opts ...grpc.CallOption) (*Version_Response, error)
	CreateSmerd(ctx context.Context, in *CreateSmerd_Request, opts ...grpc.CallOption) (*Smerd, error)
	ListSmerds(ctx context.Context, in *ListSmerds_Request, opts ...grpc.CallOption) (*ListSmerds_Response, error)
	DropSmerd(ctx context.Context, in *DropSmerd_Request, opts ...grpc.CallOption) (*DropSmerd_Response, error)
	GetHardware(ctx context.Context, in *GetHardware_Request, opts ...grpc.CallOption) (*GetHardware_Response, error)
	FetchConfig(ctx context.Context, in *FetchConfig_Request, opts ...grpc.CallOption) (*FetchConfig_Response, error)
}

type velezAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewVelezAPIClient(cc grpc.ClientConnInterface) VelezAPIClient {
	return &velezAPIClient{cc}
}

func (c *velezAPIClient) Version(ctx context.Context, in *Version_Request, opts ...grpc.CallOption) (*Version_Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Version_Response)
	err := c.cc.Invoke(ctx, VelezAPI_Version_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *velezAPIClient) CreateSmerd(ctx context.Context, in *CreateSmerd_Request, opts ...grpc.CallOption) (*Smerd, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Smerd)
	err := c.cc.Invoke(ctx, VelezAPI_CreateSmerd_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *velezAPIClient) ListSmerds(ctx context.Context, in *ListSmerds_Request, opts ...grpc.CallOption) (*ListSmerds_Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListSmerds_Response)
	err := c.cc.Invoke(ctx, VelezAPI_ListSmerds_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *velezAPIClient) DropSmerd(ctx context.Context, in *DropSmerd_Request, opts ...grpc.CallOption) (*DropSmerd_Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DropSmerd_Response)
	err := c.cc.Invoke(ctx, VelezAPI_DropSmerd_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *velezAPIClient) GetHardware(ctx context.Context, in *GetHardware_Request, opts ...grpc.CallOption) (*GetHardware_Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetHardware_Response)
	err := c.cc.Invoke(ctx, VelezAPI_GetHardware_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *velezAPIClient) FetchConfig(ctx context.Context, in *FetchConfig_Request, opts ...grpc.CallOption) (*FetchConfig_Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(FetchConfig_Response)
	err := c.cc.Invoke(ctx, VelezAPI_FetchConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VelezAPIServer is the server API for VelezAPI service.
// All implementations must embed UnimplementedVelezAPIServer
// for forward compatibility.
type VelezAPIServer interface {
	Version(context.Context, *Version_Request) (*Version_Response, error)
	CreateSmerd(context.Context, *CreateSmerd_Request) (*Smerd, error)
	ListSmerds(context.Context, *ListSmerds_Request) (*ListSmerds_Response, error)
	DropSmerd(context.Context, *DropSmerd_Request) (*DropSmerd_Response, error)
	GetHardware(context.Context, *GetHardware_Request) (*GetHardware_Response, error)
	FetchConfig(context.Context, *FetchConfig_Request) (*FetchConfig_Response, error)
	mustEmbedUnimplementedVelezAPIServer()
}

// UnimplementedVelezAPIServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedVelezAPIServer struct{}

func (UnimplementedVelezAPIServer) Version(context.Context, *Version_Request) (*Version_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Version not implemented")
}
func (UnimplementedVelezAPIServer) CreateSmerd(context.Context, *CreateSmerd_Request) (*Smerd, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSmerd not implemented")
}
func (UnimplementedVelezAPIServer) ListSmerds(context.Context, *ListSmerds_Request) (*ListSmerds_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSmerds not implemented")
}
func (UnimplementedVelezAPIServer) DropSmerd(context.Context, *DropSmerd_Request) (*DropSmerd_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DropSmerd not implemented")
}
func (UnimplementedVelezAPIServer) GetHardware(context.Context, *GetHardware_Request) (*GetHardware_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHardware not implemented")
}
func (UnimplementedVelezAPIServer) FetchConfig(context.Context, *FetchConfig_Request) (*FetchConfig_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchConfig not implemented")
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
	in := new(Version_Request)
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
		return srv.(VelezAPIServer).Version(ctx, req.(*Version_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _VelezAPI_CreateSmerd_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSmerd_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VelezAPIServer).CreateSmerd(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: VelezAPI_CreateSmerd_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VelezAPIServer).CreateSmerd(ctx, req.(*CreateSmerd_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _VelezAPI_ListSmerds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListSmerds_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VelezAPIServer).ListSmerds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: VelezAPI_ListSmerds_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VelezAPIServer).ListSmerds(ctx, req.(*ListSmerds_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _VelezAPI_DropSmerd_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DropSmerd_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VelezAPIServer).DropSmerd(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: VelezAPI_DropSmerd_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VelezAPIServer).DropSmerd(ctx, req.(*DropSmerd_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _VelezAPI_GetHardware_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetHardware_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VelezAPIServer).GetHardware(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: VelezAPI_GetHardware_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VelezAPIServer).GetHardware(ctx, req.(*GetHardware_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _VelezAPI_FetchConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchConfig_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VelezAPIServer).FetchConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: VelezAPI_FetchConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VelezAPIServer).FetchConfig(ctx, req.(*FetchConfig_Request))
	}
	return interceptor(ctx, in, info, handler)
}

// VelezAPI_ServiceDesc is the grpc.ServiceDesc for VelezAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var VelezAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "velez_api.VelezAPI",
	HandlerType: (*VelezAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Version",
			Handler:    _VelezAPI_Version_Handler,
		},
		{
			MethodName: "CreateSmerd",
			Handler:    _VelezAPI_CreateSmerd_Handler,
		},
		{
			MethodName: "ListSmerds",
			Handler:    _VelezAPI_ListSmerds_Handler,
		},
		{
			MethodName: "DropSmerd",
			Handler:    _VelezAPI_DropSmerd_Handler,
		},
		{
			MethodName: "GetHardware",
			Handler:    _VelezAPI_GetHardware_Handler,
		},
		{
			MethodName: "FetchConfig",
			Handler:    _VelezAPI_FetchConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/grpc/velez_api.proto",
}
