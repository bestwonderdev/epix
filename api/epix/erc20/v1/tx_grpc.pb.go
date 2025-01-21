// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: epix/erc20/v1/tx.proto

package erc20v1

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
	Msg_ConvertCoin_FullMethodName                   = "/epix.erc20.v1.Msg/ConvertCoin"
	Msg_ConvertERC20_FullMethodName                  = "/epix.erc20.v1.Msg/ConvertERC20"
	Msg_UpdateParams_FullMethodName                  = "/epix.erc20.v1.Msg/UpdateParams"
	Msg_RegisterCoinProposal_FullMethodName          = "/epix.erc20.v1.Msg/RegisterCoinProposal"
	Msg_RegisterERC20Proposal_FullMethodName         = "/epix.erc20.v1.Msg/RegisterERC20Proposal"
	Msg_ToggleTokenConversionProposal_FullMethodName = "/epix.erc20.v1.Msg/ToggleTokenConversionProposal"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	// ConvertCoin mints a ERC20 representation of the native Cosmos coin denom
	// that is registered on the token mapping.
	ConvertCoin(ctx context.Context, in *MsgConvertCoin, opts ...grpc.CallOption) (*MsgConvertCoinResponse, error)
	// ConvertERC20 mints a native Cosmos coin representation of the ERC20 token
	// contract that is registered on the token mapping.
	ConvertERC20(ctx context.Context, in *MsgConvertERC20, opts ...grpc.CallOption) (*MsgConvertERC20Response, error)
	// UpdateParams updates the parameters of the x/erc20 module.
	UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error)
	// RegisterCoinProposal defines a method to create a proposal to register a
	// token pair for a native Cosmos coin.
	RegisterCoinProposal(ctx context.Context, in *MsgRegisterCoin, opts ...grpc.CallOption) (*MsgRegisterCoinResponse, error)
	// RegisterERC20Proposal defines a method to create a proposal to register a
	// token pair for an ERC20 token.
	RegisterERC20Proposal(ctx context.Context, in *MsgRegisterERC20, opts ...grpc.CallOption) (*MsgRegisterERC20Response, error)
	// ToggleTokenConversionProposal defines a method to create a proposal to
	// toggle the conversion of a token pair.
	ToggleTokenConversionProposal(ctx context.Context, in *MsgToggleTokenConversion, opts ...grpc.CallOption) (*MsgToggleTokenConversionResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) ConvertCoin(ctx context.Context, in *MsgConvertCoin, opts ...grpc.CallOption) (*MsgConvertCoinResponse, error) {
	out := new(MsgConvertCoinResponse)
	err := c.cc.Invoke(ctx, Msg_ConvertCoin_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ConvertERC20(ctx context.Context, in *MsgConvertERC20, opts ...grpc.CallOption) (*MsgConvertERC20Response, error) {
	out := new(MsgConvertERC20Response)
	err := c.cc.Invoke(ctx, Msg_ConvertERC20_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error) {
	out := new(MsgUpdateParamsResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateParams_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RegisterCoinProposal(ctx context.Context, in *MsgRegisterCoin, opts ...grpc.CallOption) (*MsgRegisterCoinResponse, error) {
	out := new(MsgRegisterCoinResponse)
	err := c.cc.Invoke(ctx, Msg_RegisterCoinProposal_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RegisterERC20Proposal(ctx context.Context, in *MsgRegisterERC20, opts ...grpc.CallOption) (*MsgRegisterERC20Response, error) {
	out := new(MsgRegisterERC20Response)
	err := c.cc.Invoke(ctx, Msg_RegisterERC20Proposal_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ToggleTokenConversionProposal(ctx context.Context, in *MsgToggleTokenConversion, opts ...grpc.CallOption) (*MsgToggleTokenConversionResponse, error) {
	out := new(MsgToggleTokenConversionResponse)
	err := c.cc.Invoke(ctx, Msg_ToggleTokenConversionProposal_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
type MsgServer interface {
	// ConvertCoin mints a ERC20 representation of the native Cosmos coin denom
	// that is registered on the token mapping.
	ConvertCoin(context.Context, *MsgConvertCoin) (*MsgConvertCoinResponse, error)
	// ConvertERC20 mints a native Cosmos coin representation of the ERC20 token
	// contract that is registered on the token mapping.
	ConvertERC20(context.Context, *MsgConvertERC20) (*MsgConvertERC20Response, error)
	// UpdateParams updates the parameters of the x/erc20 module.
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
	// RegisterCoinProposal defines a method to create a proposal to register a
	// token pair for a native Cosmos coin.
	RegisterCoinProposal(context.Context, *MsgRegisterCoin) (*MsgRegisterCoinResponse, error)
	// RegisterERC20Proposal defines a method to create a proposal to register a
	// token pair for an ERC20 token.
	RegisterERC20Proposal(context.Context, *MsgRegisterERC20) (*MsgRegisterERC20Response, error)
	// ToggleTokenConversionProposal defines a method to create a proposal to
	// toggle the conversion of a token pair.
	ToggleTokenConversionProposal(context.Context, *MsgToggleTokenConversion) (*MsgToggleTokenConversionResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) ConvertCoin(context.Context, *MsgConvertCoin) (*MsgConvertCoinResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConvertCoin not implemented")
}
func (UnimplementedMsgServer) ConvertERC20(context.Context, *MsgConvertERC20) (*MsgConvertERC20Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConvertERC20 not implemented")
}
func (UnimplementedMsgServer) UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateParams not implemented")
}
func (UnimplementedMsgServer) RegisterCoinProposal(context.Context, *MsgRegisterCoin) (*MsgRegisterCoinResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterCoinProposal not implemented")
}
func (UnimplementedMsgServer) RegisterERC20Proposal(context.Context, *MsgRegisterERC20) (*MsgRegisterERC20Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterERC20Proposal not implemented")
}
func (UnimplementedMsgServer) ToggleTokenConversionProposal(context.Context, *MsgToggleTokenConversion) (*MsgToggleTokenConversionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleTokenConversionProposal not implemented")
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

func _Msg_ConvertCoin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgConvertCoin)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ConvertCoin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_ConvertCoin_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ConvertCoin(ctx, req.(*MsgConvertCoin))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ConvertERC20_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgConvertERC20)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ConvertERC20(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_ConvertERC20_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ConvertERC20(ctx, req.(*MsgConvertERC20))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateParams_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateParams(ctx, req.(*MsgUpdateParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RegisterCoinProposal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRegisterCoin)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RegisterCoinProposal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_RegisterCoinProposal_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RegisterCoinProposal(ctx, req.(*MsgRegisterCoin))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RegisterERC20Proposal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRegisterERC20)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RegisterERC20Proposal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_RegisterERC20Proposal_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RegisterERC20Proposal(ctx, req.(*MsgRegisterERC20))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ToggleTokenConversionProposal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgToggleTokenConversion)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ToggleTokenConversionProposal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_ToggleTokenConversionProposal_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ToggleTokenConversionProposal(ctx, req.(*MsgToggleTokenConversion))
	}
	return interceptor(ctx, in, info, handler)
}

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "epix.erc20.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ConvertCoin",
			Handler:    _Msg_ConvertCoin_Handler,
		},
		{
			MethodName: "ConvertERC20",
			Handler:    _Msg_ConvertERC20_Handler,
		},
		{
			MethodName: "UpdateParams",
			Handler:    _Msg_UpdateParams_Handler,
		},
		{
			MethodName: "RegisterCoinProposal",
			Handler:    _Msg_RegisterCoinProposal_Handler,
		},
		{
			MethodName: "RegisterERC20Proposal",
			Handler:    _Msg_RegisterERC20Proposal_Handler,
		},
		{
			MethodName: "ToggleTokenConversionProposal",
			Handler:    _Msg_ToggleTokenConversionProposal_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "epix/erc20/v1/tx.proto",
}
