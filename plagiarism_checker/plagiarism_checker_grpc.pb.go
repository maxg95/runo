// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package plagiarism_checker

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

// PlagiarismCheckerClient is the client API for PlagiarismChecker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PlagiarismCheckerClient interface {
	CheckPlagiarism(ctx context.Context, in *CheckPlagiarismRequest, opts ...grpc.CallOption) (*CheckPlagiarismResponse, error)
}

type plagiarismCheckerClient struct {
	cc grpc.ClientConnInterface
}

func NewPlagiarismCheckerClient(cc grpc.ClientConnInterface) PlagiarismCheckerClient {
	return &plagiarismCheckerClient{cc}
}

func (c *plagiarismCheckerClient) CheckPlagiarism(ctx context.Context, in *CheckPlagiarismRequest, opts ...grpc.CallOption) (*CheckPlagiarismResponse, error) {
	out := new(CheckPlagiarismResponse)
	err := c.cc.Invoke(ctx, "/plagiarism_checker.PlagiarismChecker/CheckPlagiarism", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PlagiarismCheckerServer is the server API for PlagiarismChecker service.
// All implementations must embed UnimplementedPlagiarismCheckerServer
// for forward compatibility
type PlagiarismCheckerServer interface {
	CheckPlagiarism(context.Context, *CheckPlagiarismRequest) (*CheckPlagiarismResponse, error)
	mustEmbedUnimplementedPlagiarismCheckerServer()
}

// UnimplementedPlagiarismCheckerServer must be embedded to have forward compatible implementations.
type UnimplementedPlagiarismCheckerServer struct {
}

func (UnimplementedPlagiarismCheckerServer) CheckPlagiarism(context.Context, *CheckPlagiarismRequest) (*CheckPlagiarismResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckPlagiarism not implemented")
}
func (UnimplementedPlagiarismCheckerServer) mustEmbedUnimplementedPlagiarismCheckerServer() {}

// UnsafePlagiarismCheckerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PlagiarismCheckerServer will
// result in compilation errors.
type UnsafePlagiarismCheckerServer interface {
	mustEmbedUnimplementedPlagiarismCheckerServer()
}

func RegisterPlagiarismCheckerServer(s grpc.ServiceRegistrar, srv PlagiarismCheckerServer) {
	s.RegisterService(&PlagiarismChecker_ServiceDesc, srv)
}

func _PlagiarismChecker_CheckPlagiarism_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckPlagiarismRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PlagiarismCheckerServer).CheckPlagiarism(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plagiarism_checker.PlagiarismChecker/CheckPlagiarism",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PlagiarismCheckerServer).CheckPlagiarism(ctx, req.(*CheckPlagiarismRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PlagiarismChecker_ServiceDesc is the grpc.ServiceDesc for PlagiarismChecker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PlagiarismChecker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "plagiarism_checker.PlagiarismChecker",
	HandlerType: (*PlagiarismCheckerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckPlagiarism",
			Handler:    _PlagiarismChecker_CheckPlagiarism_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "plagiarism_checker.proto",
}
