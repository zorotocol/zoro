// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: grpc.proto

package proto

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Gun_OK_FullMethodName  = "/Gun/OK"
	Gun_Tun_FullMethodName = "/Gun/Tun"
)

// GunClient is the client API for Gun service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GunClient interface {
	OK(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error)
	Tun(ctx context.Context, opts ...grpc.CallOption) (Gun_TunClient, error)
}

type gunClient struct {
	cc grpc.ClientConnInterface
}

func NewGunClient(cc grpc.ClientConnInterface) GunClient {
	return &gunClient{cc}
}

func (c *gunClient) OK(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, Gun_OK_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gunClient) Tun(ctx context.Context, opts ...grpc.CallOption) (Gun_TunClient, error) {
	stream, err := c.cc.NewStream(ctx, &Gun_ServiceDesc.Streams[0], Gun_Tun_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &gunTunClient{stream}
	return x, nil
}

type Gun_TunClient interface {
	Send(*Hunk) error
	Recv() (*Hunk, error)
	grpc.ClientStream
}

type gunTunClient struct {
	grpc.ClientStream
}

func (x *gunTunClient) Send(m *Hunk) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gunTunClient) Recv() (*Hunk, error) {
	m := new(Hunk)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GunServer is the server API for Gun service.
// All implementations must embed UnimplementedGunServer
// for forward compatibility
type GunServer interface {
	OK(context.Context, *empty.Empty) (*empty.Empty, error)
	Tun(Gun_TunServer) error
	mustEmbedUnimplementedGunServer()
}

// UnimplementedGunServer must be embedded to have forward compatible implementations.
type UnimplementedGunServer struct {
}

func (UnimplementedGunServer) OK(context.Context, *empty.Empty) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OK not implemented")
}
func (UnimplementedGunServer) Tun(Gun_TunServer) error {
	return status.Errorf(codes.Unimplemented, "method Tun not implemented")
}
func (UnimplementedGunServer) mustEmbedUnimplementedGunServer() {}

// UnsafeGunServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GunServer will
// result in compilation errors.
type UnsafeGunServer interface {
	mustEmbedUnimplementedGunServer()
}

func RegisterGunServer(s grpc.ServiceRegistrar, srv GunServer) {
	s.RegisterService(&Gun_ServiceDesc, srv)
}

func _Gun_OK_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GunServer).OK(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Gun_OK_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GunServer).OK(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gun_Tun_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GunServer).Tun(&gunTunServer{stream})
}

type Gun_TunServer interface {
	Send(*Hunk) error
	Recv() (*Hunk, error)
	grpc.ServerStream
}

type gunTunServer struct {
	grpc.ServerStream
}

func (x *gunTunServer) Send(m *Hunk) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gunTunServer) Recv() (*Hunk, error) {
	m := new(Hunk)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Gun_ServiceDesc is the grpc.ServiceDesc for Gun service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Gun_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Gun",
	HandlerType: (*GunServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "OK",
			Handler:    _Gun_OK_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Tun",
			Handler:       _Gun_Tun_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "grpc.proto",
}
