package service

import (
	"context"
	"log/slog"

	"github.com/deimossy/order-processing-system/internal/user/config"
	"github.com/deimossy/order-processing-system/internal/user/domain"
)

type UserRepo interface {
	SaveUser(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	DeleteByID(ctx context.Context, id string) error
}

type RefreshTokenRepo interface {
	SaveRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type UserService struct {
	logger           *slog.Logger
	cfg              config.Config
	userRepo         UserRepo
	refreshTokenRepo RefreshTokenRepo
}

func NewUserService(logger *slog.Logger, cfg config.Config, userRepo UserRepo, refreshTokenRepo RefreshTokenRepo) *UserService {
	return &UserService{
		logger:           logger,
		cfg:              cfg,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}
