// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.7
// source: pkg/grpc/user_devices.proto

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// UserDeviceServiceClient is the client API for UserDeviceService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserDeviceServiceClient interface {
	GetUserDevice(ctx context.Context, in *GetUserDeviceRequest, opts ...grpc.CallOption) (*UserDevice, error)
	GetUserDeviceByTokenId(ctx context.Context, in *GetUserDeviceByTokenIdRequest, opts ...grpc.CallOption) (*UserDevice, error)
	ListUserDevicesForUser(ctx context.Context, in *ListUserDevicesForUserRequest, opts ...grpc.CallOption) (*ListUserDevicesForUserResponse, error)
	ApplyHardwareTemplate(ctx context.Context, in *ApplyHardwareTemplateRequest, opts ...grpc.CallOption) (*ApplyHardwareTemplateResponse, error)
	GetAllUserDeviceValuation(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ValuationResponse, error)
	GetUserDeviceByAutoPIUnitId(ctx context.Context, in *GetUserDeviceByAutoPIUnitIdRequest, opts ...grpc.CallOption) (*UserDeviceAutoPIUnitResponse, error)
	GetClaimedVehiclesGrowth(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ClaimedVehiclesGrowth, error)
	CreateTemplate(ctx context.Context, in *CreateTemplateRequest, opts ...grpc.CallOption) (*CreateTemplateResponse, error)
	RegisterUserDeviceFromVIN(ctx context.Context, in *RegisterUserDeviceFromVINRequest, opts ...grpc.CallOption) (*RegisterUserDeviceFromVINResponse, error)
	UpdateDeviceIntegrationStatus(ctx context.Context, in *UpdateDeviceIntegrationStatusRequest, opts ...grpc.CallOption) (*UserDevice, error)
}

type userDeviceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserDeviceServiceClient(cc grpc.ClientConnInterface) UserDeviceServiceClient {
	return &userDeviceServiceClient{cc}
}

func (c *userDeviceServiceClient) GetUserDevice(ctx context.Context, in *GetUserDeviceRequest, opts ...grpc.CallOption) (*UserDevice, error) {
	out := new(UserDevice)
	err := c.cc.Invoke(ctx, "/devices.UserDeviceService/GetUserDevice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) GetUserDeviceByTokenId(ctx context.Context, in *GetUserDeviceByTokenIdRequest, opts ...grpc.CallOption) (*UserDevice, error) {
	out := new(UserDevice)
	err := c.cc.Invoke(ctx, "/devices.UserDeviceService/GetUserDeviceByTokenId", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) ListUserDevicesForUser(ctx context.Context, in *ListUserDevicesForUserRequest, opts ...grpc.CallOption) (*ListUserDevicesForUserResponse, error) {
	out := new(ListUserDevicesForUserResponse)
	err := c.cc.Invoke(ctx, "/devices.UserDeviceService/ListUserDevicesForUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) ApplyHardwareTemplate(ctx context.Context, in *ApplyHardwareTemplateRequest, opts ...grpc.CallOption) (*ApplyHardwareTemplateResponse, error) {
	out := new(ApplyHardwareTemplateResponse)
	err := c.cc.Invoke(ctx, "/devices.UserDeviceService/ApplyHardwareTemplate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) GetAllUserDeviceValuation(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ValuationResponse, error) {
	out := new(ValuationResponse)
	err := c.cc.Invoke(ctx, "/devices.UserDeviceService/GetAllUserDeviceValuation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) GetUserDeviceByAutoPIUnitId(ctx context.Context, in *GetUserDeviceByAutoPIUnitIdRequest, opts ...grpc.CallOption) (*UserDeviceAutoPIUnitResponse, error) {
	out := new(UserDeviceAutoPIUnitResponse)
	err := c.cc.Invoke(ctx, "/devices.UserDeviceService/GetUserDeviceByAutoPIUnitId", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) GetClaimedVehiclesGrowth(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ClaimedVehiclesGrowth, error) {
	out := new(ClaimedVehiclesGrowth)
	err := c.cc.Invoke(ctx, "/devices.UserDeviceService/GetClaimedVehiclesGrowth", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) CreateTemplate(ctx context.Context, in *CreateTemplateRequest, opts ...grpc.CallOption) (*CreateTemplateResponse, error) {
	out := new(CreateTemplateResponse)
	err := c.cc.Invoke(ctx, "/devices.UserDeviceService/CreateTemplate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) RegisterUserDeviceFromVIN(ctx context.Context, in *RegisterUserDeviceFromVINRequest, opts ...grpc.CallOption) (*RegisterUserDeviceFromVINResponse, error) {
	out := new(RegisterUserDeviceFromVINResponse)
	err := c.cc.Invoke(ctx, "/devices.UserDeviceService/RegisterUserDeviceFromVIN", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) UpdateDeviceIntegrationStatus(ctx context.Context, in *UpdateDeviceIntegrationStatusRequest, opts ...grpc.CallOption) (*UserDevice, error) {
	out := new(UserDevice)
	err := c.cc.Invoke(ctx, "/devices.UserDeviceService/UpdateDeviceIntegrationStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserDeviceServiceServer is the server API for UserDeviceService service.
// All implementations must embed UnimplementedUserDeviceServiceServer
// for forward compatibility
type UserDeviceServiceServer interface {
	GetUserDevice(context.Context, *GetUserDeviceRequest) (*UserDevice, error)
	GetUserDeviceByTokenId(context.Context, *GetUserDeviceByTokenIdRequest) (*UserDevice, error)
	ListUserDevicesForUser(context.Context, *ListUserDevicesForUserRequest) (*ListUserDevicesForUserResponse, error)
	ApplyHardwareTemplate(context.Context, *ApplyHardwareTemplateRequest) (*ApplyHardwareTemplateResponse, error)
	GetAllUserDeviceValuation(context.Context, *emptypb.Empty) (*ValuationResponse, error)
	GetUserDeviceByAutoPIUnitId(context.Context, *GetUserDeviceByAutoPIUnitIdRequest) (*UserDeviceAutoPIUnitResponse, error)
	GetClaimedVehiclesGrowth(context.Context, *emptypb.Empty) (*ClaimedVehiclesGrowth, error)
	CreateTemplate(context.Context, *CreateTemplateRequest) (*CreateTemplateResponse, error)
	RegisterUserDeviceFromVIN(context.Context, *RegisterUserDeviceFromVINRequest) (*RegisterUserDeviceFromVINResponse, error)
	UpdateDeviceIntegrationStatus(context.Context, *UpdateDeviceIntegrationStatusRequest) (*UserDevice, error)
	mustEmbedUnimplementedUserDeviceServiceServer()
}

// UnimplementedUserDeviceServiceServer must be embedded to have forward compatible implementations.
type UnimplementedUserDeviceServiceServer struct {
}

func (UnimplementedUserDeviceServiceServer) GetUserDevice(context.Context, *GetUserDeviceRequest) (*UserDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserDevice not implemented")
}
func (UnimplementedUserDeviceServiceServer) GetUserDeviceByTokenId(context.Context, *GetUserDeviceByTokenIdRequest) (*UserDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserDeviceByTokenId not implemented")
}
func (UnimplementedUserDeviceServiceServer) ListUserDevicesForUser(context.Context, *ListUserDevicesForUserRequest) (*ListUserDevicesForUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUserDevicesForUser not implemented")
}
func (UnimplementedUserDeviceServiceServer) ApplyHardwareTemplate(context.Context, *ApplyHardwareTemplateRequest) (*ApplyHardwareTemplateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplyHardwareTemplate not implemented")
}
func (UnimplementedUserDeviceServiceServer) GetAllUserDeviceValuation(context.Context, *emptypb.Empty) (*ValuationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllUserDeviceValuation not implemented")
}
func (UnimplementedUserDeviceServiceServer) GetUserDeviceByAutoPIUnitId(context.Context, *GetUserDeviceByAutoPIUnitIdRequest) (*UserDeviceAutoPIUnitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserDeviceByAutoPIUnitId not implemented")
}
func (UnimplementedUserDeviceServiceServer) GetClaimedVehiclesGrowth(context.Context, *emptypb.Empty) (*ClaimedVehiclesGrowth, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClaimedVehiclesGrowth not implemented")
}
func (UnimplementedUserDeviceServiceServer) CreateTemplate(context.Context, *CreateTemplateRequest) (*CreateTemplateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTemplate not implemented")
}
func (UnimplementedUserDeviceServiceServer) RegisterUserDeviceFromVIN(context.Context, *RegisterUserDeviceFromVINRequest) (*RegisterUserDeviceFromVINResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterUserDeviceFromVIN not implemented")
}
func (UnimplementedUserDeviceServiceServer) UpdateDeviceIntegrationStatus(context.Context, *UpdateDeviceIntegrationStatusRequest) (*UserDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDeviceIntegrationStatus not implemented")
}
func (UnimplementedUserDeviceServiceServer) mustEmbedUnimplementedUserDeviceServiceServer() {}

// UnsafeUserDeviceServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserDeviceServiceServer will
// result in compilation errors.
type UnsafeUserDeviceServiceServer interface {
	mustEmbedUnimplementedUserDeviceServiceServer()
}

func RegisterUserDeviceServiceServer(s grpc.ServiceRegistrar, srv UserDeviceServiceServer) {
	s.RegisterService(&UserDeviceService_ServiceDesc, srv)
}

func _UserDeviceService_GetUserDevice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserDeviceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetUserDevice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/devices.UserDeviceService/GetUserDevice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetUserDevice(ctx, req.(*GetUserDeviceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_GetUserDeviceByTokenId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserDeviceByTokenIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetUserDeviceByTokenId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/devices.UserDeviceService/GetUserDeviceByTokenId",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetUserDeviceByTokenId(ctx, req.(*GetUserDeviceByTokenIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_ListUserDevicesForUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListUserDevicesForUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).ListUserDevicesForUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/devices.UserDeviceService/ListUserDevicesForUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).ListUserDevicesForUser(ctx, req.(*ListUserDevicesForUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_ApplyHardwareTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApplyHardwareTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).ApplyHardwareTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/devices.UserDeviceService/ApplyHardwareTemplate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).ApplyHardwareTemplate(ctx, req.(*ApplyHardwareTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_GetAllUserDeviceValuation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetAllUserDeviceValuation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/devices.UserDeviceService/GetAllUserDeviceValuation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetAllUserDeviceValuation(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_GetUserDeviceByAutoPIUnitId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserDeviceByAutoPIUnitIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetUserDeviceByAutoPIUnitId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/devices.UserDeviceService/GetUserDeviceByAutoPIUnitId",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetUserDeviceByAutoPIUnitId(ctx, req.(*GetUserDeviceByAutoPIUnitIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_GetClaimedVehiclesGrowth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetClaimedVehiclesGrowth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/devices.UserDeviceService/GetClaimedVehiclesGrowth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetClaimedVehiclesGrowth(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_CreateTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).CreateTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/devices.UserDeviceService/CreateTemplate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).CreateTemplate(ctx, req.(*CreateTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_RegisterUserDeviceFromVIN_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterUserDeviceFromVINRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).RegisterUserDeviceFromVIN(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/devices.UserDeviceService/RegisterUserDeviceFromVIN",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).RegisterUserDeviceFromVIN(ctx, req.(*RegisterUserDeviceFromVINRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_UpdateDeviceIntegrationStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateDeviceIntegrationStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).UpdateDeviceIntegrationStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/devices.UserDeviceService/UpdateDeviceIntegrationStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).UpdateDeviceIntegrationStatus(ctx, req.(*UpdateDeviceIntegrationStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UserDeviceService_ServiceDesc is the grpc.ServiceDesc for UserDeviceService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserDeviceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "devices.UserDeviceService",
	HandlerType: (*UserDeviceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetUserDevice",
			Handler:    _UserDeviceService_GetUserDevice_Handler,
		},
		{
			MethodName: "GetUserDeviceByTokenId",
			Handler:    _UserDeviceService_GetUserDeviceByTokenId_Handler,
		},
		{
			MethodName: "ListUserDevicesForUser",
			Handler:    _UserDeviceService_ListUserDevicesForUser_Handler,
		},
		{
			MethodName: "ApplyHardwareTemplate",
			Handler:    _UserDeviceService_ApplyHardwareTemplate_Handler,
		},
		{
			MethodName: "GetAllUserDeviceValuation",
			Handler:    _UserDeviceService_GetAllUserDeviceValuation_Handler,
		},
		{
			MethodName: "GetUserDeviceByAutoPIUnitId",
			Handler:    _UserDeviceService_GetUserDeviceByAutoPIUnitId_Handler,
		},
		{
			MethodName: "GetClaimedVehiclesGrowth",
			Handler:    _UserDeviceService_GetClaimedVehiclesGrowth_Handler,
		},
		{
			MethodName: "CreateTemplate",
			Handler:    _UserDeviceService_CreateTemplate_Handler,
		},
		{
			MethodName: "RegisterUserDeviceFromVIN",
			Handler:    _UserDeviceService_RegisterUserDeviceFromVIN_Handler,
		},
		{
			MethodName: "UpdateDeviceIntegrationStatus",
			Handler:    _UserDeviceService_UpdateDeviceIntegrationStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/user_devices.proto",
}
