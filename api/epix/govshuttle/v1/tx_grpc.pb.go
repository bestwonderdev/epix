// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: epix/govshuttle/v1/tx.proto

package govshuttlev1

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

const (
	Msg_LendingMarketProposal_FullMethodName = "/epix.govshuttle.v1.Msg/LendingMarketProposal"
	Msg_TreasuryProposal_FullMethodName      = "/epix.govshuttle.v1.Msg/TreasuryProposal"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	// LendingMarketProposal append the lending market proposal of the
	// x/govshuttle module.
	LendingMarketProposal(ctx context.Context, in *MsgLendingMarketProposal, opts ...grpc.CallOption) (*MsgLendingMarketProposalResponse, error)
	// TreasuryProposal append the treasury proposal of the x/govshuttle module.
	TreasuryProposal(ctx context.Context, in *MsgTreasuryProposal, opts ...grpc.CallOption) (*MsgTreasuryProposalResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) LendingMarketProposal(ctx context.Context, in *MsgLendingMarketProposal, opts ...grpc.CallOption) (*MsgLendingMarketProposalResponse, error) {
	out := new(MsgLendingMarketProposalResponse)
	err := c.cc.Invoke(ctx, Msg_LendingMarketProposal_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) TreasuryProposal(ctx context.Context, in *MsgTreasuryProposal, opts ...grpc.CallOption) (*MsgTreasuryProposalResponse, error) {
	out := new(MsgTreasuryProposalResponse)
	err := c.cc.Invoke(ctx, Msg_TreasuryProposal_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
type MsgServer interface {
	// LendingMarketProposal append the lending market proposal of the
	// x/govshuttle module.
	LendingMarketProposal(context.Context, *MsgLendingMarketProposal) (*MsgLendingMarketProposalResponse, error)
	// TreasuryProposal append the treasury proposal of the x/govshuttle module.
	TreasuryProposal(context.Context, *MsgTreasuryProposal) (*MsgTreasuryProposalResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) LendingMarketProposal(context.Context, *MsgLendingMarketProposal) (*MsgLendingMarketProposalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LendingMarketProposal not implemented")
}
func (UnimplementedMsgServer) TreasuryProposal(context.Context, *MsgTreasuryProposal) (*MsgTreasuryProposalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TreasuryProposal not implemented")
}
func (UnimplementedMsgServer) mustEmbedUnimplementedMsgServer() {}

// UnsafeMsgServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MsgServer will
// result in compilation errors.
type UnsafeMsgServer interface {
	mustEmbedUnimplementedMsgServer()
}

func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&Msg_ServiceDesc, srv)
}

func _Msg_LendingMarketProposal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgLendingMarketProposal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).LendingMarketProposal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_LendingMarketProposal_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).LendingMarketProposal(ctx, req.(*MsgLendingMarketProposal))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_TreasuryProposal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgTreasuryProposal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).TreasuryProposal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_TreasuryProposal_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).TreasuryProposal(ctx, req.(*MsgTreasuryProposal))
	}
	return interceptor(ctx, in, info, handler)
}

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "epix.govshuttle.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LendingMarketProposal",
			Handler:    _Msg_LendingMarketProposal_Handler,
		},
		{
			MethodName: "TreasuryProposal",
			Handler:    _Msg_TreasuryProposal_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "epix/govshuttle/v1/tx.proto",
}
