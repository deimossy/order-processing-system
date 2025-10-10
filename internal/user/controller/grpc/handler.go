package grpc

import (
	"context"
	"github.com/deimossy/order-processing-system/internal/user/usecase"
	user_v1 "github.com/deimossy/order-processing-system/protobuf/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Controller struct {
	user_v1.UnimplementedUserServiceServer

	svc *usecase.UserService
}

func NewController(svc *usecase.UserService) *Controller {
	return &Controller{
		svc: svc,
	}
}

func (c *Controller) Register(ctx context.Context, req *user_v1.RegisterRequest) (*user_v1.TokenPair, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, status.Errorf(codes.OK, "method: Register not implemented")
}

func (c *Controller) Login(ctx context.Context, req *user_v1.LoginRequest) (*user_v1.TokenPair, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, status.Errorf(codes.OK, "method: Login not implemented")
}

func (c *Controller) Refresh(ctx context.Context, req *user_v1.RefreshRequest) (*user_v1.TokenPair, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, status.Errorf(codes.OK, "method: Refresh not implemented")
}

func (c *Controller) Logout(ctx context.Context, req *user_v1.LogoutRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, status.Errorf(codes.OK, "method: Logout not implemented")
}

func (c *Controller) GetProfile(ctx context.Context, req *user_v1.GetProfileRequest) (*user_v1.UserProfile, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, status.Errorf(codes.OK, "method: GetProfile not implemented")
}
