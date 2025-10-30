package grpc

import (
	"context"
	"errors"
	"github.com/deimossy/order-processing-system/internal/user/usecase"
	errs "github.com/deimossy/order-processing-system/pkg/errors"
	user_v1 "github.com/deimossy/order-processing-system/protobuf/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	domainTP, err := c.svc.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, writeGRPCStatus(err)
	}

	TP := &user_v1.TokenPair{
		AccessToken:          domainTP.AccessToken,
		RefreshToken:         domainTP.RefreshToken,
		AccessTokenExpiresAt: timestamppb.New(domainTP.ExpiresAt),
		UserId:               domainTP.UserID,
	}

	return TP, status.Error(codes.OK, "user registered")
}

func (c *Controller) Login(ctx context.Context, req *user_v1.LoginRequest) (*user_v1.TokenPair, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	domainTP, err := c.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, writeGRPCStatus(err)
	}

	TP := &user_v1.TokenPair{
		AccessToken:          domainTP.AccessToken,
		RefreshToken:         domainTP.RefreshToken,
		AccessTokenExpiresAt: timestamppb.New(domainTP.ExpiresAt),
		UserId:               domainTP.UserID,
	}

	return TP, status.Error(codes.OK, "user logged in")
}

func (c *Controller) Refresh(ctx context.Context, req *user_v1.RefreshRequest) (*user_v1.TokenPair, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	domainTP, err := c.svc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, writeGRPCStatus(err)
	}

	TP := &user_v1.TokenPair{
		AccessToken:          domainTP.AccessToken,
		RefreshToken:         domainTP.RefreshToken,
		AccessTokenExpiresAt: timestamppb.New(domainTP.ExpiresAt),
		UserId:               domainTP.UserID,
	}

	return TP, status.Error(codes.OK, "user refreshed")
}

func (c *Controller) Logout(ctx context.Context, req *user_v1.LogoutRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := c.svc.Logout(ctx, req.RefreshToken)
	if err != nil {
		return nil, writeGRPCStatus(err)
	}

	return &emptypb.Empty{}, status.Error(codes.OK, "user logout")
}

func (c *Controller) GetProfile(ctx context.Context, req *user_v1.GetProfileRequest) (*user_v1.UserProfile, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, status.Errorf(codes.OK, "method: GetProfile not implemented")
}

func writeGRPCStatus(err error) error {
	switch {
	case errors.Is(err, errs.ErrUserNotFound), errors.Is(err, errs.ErrRefreshTokenNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, errs.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, errs.ErrAccessTokenInvalid), errors.Is(err, errs.ErrPasswordMismatch):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, errs.ErrAccessTokenExpired):
		return status.Error(codes.FailedPrecondition, err.Error())

	case errors.Is(err, errs.ErrRefreshTokenExpired), errors.Is(err, errs.ErrRefreshTokenRevoked):
		return status.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, errs.ErrMaxRetriesAttemptsExceeded):
		return status.Error(codes.Unavailable, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
