package usecase

import (
	"context"
	"crypto/rsa"
	"errors"
	"github.com/deimossy/order-processing-system/internal/user/auth"
	"github.com/deimossy/order-processing-system/internal/user/config"
	"github.com/deimossy/order-processing-system/internal/user/domain"
	errs "github.com/deimossy/order-processing-system/pkg/errors"
	"time"
)

type RefreshTokenRepo interface {
	Save(ctx context.Context, token *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, token string) (*domain.RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, token string) error
}

type UserRepo interface {
	Save(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	DeleteByID(ctx context.Context, id string) error
}

type Repositories interface {
	UserRepo() UserRepo
	RefreshTokenRepo() RefreshTokenRepo
}

type UnitOfWork interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
	WithoutTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type UserService struct {
	cfg        config.Config
	uow        UnitOfWork
	privateKey *rsa.PrivateKey
}

func NewUserService(cfg config.Config, uow UnitOfWork, key *rsa.PrivateKey) *UserService {
	return &UserService{
		cfg:        cfg,
		uow:        uow,
		privateKey: key,
	}
}

func (s *UserService) Register(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	var (
		user  *domain.User
		token *domain.TokenPair
	)

	err := s.uow.WithTx(ctx, func(ctx context.Context, repos Repositories) error {
		// user check
		_, err := repos.UserRepo().GetByEmail(ctx, email)
		if err == nil {
			return errs.ErrUserAlreadyExists
		}
		// infra err
		if !errors.Is(err, errs.ErrNotFound) {
			return err
		}

		// hashing pass
		hashedPassword, err := auth.HashPassword(password, s.cfg.BcryptCost)
		if err != nil {
			return err
		}

		// user struct
		user = &domain.User{
			Email:        email,
			PasswordHash: hashedPassword,
			CreatedAt:    time.Now(),
		}

		// saving user
		err = repos.UserRepo().Save(ctx, user)
		if err != nil {
			if errors.Is(err, errs.ErrAlreadyExists) {
				return errs.ErrUserAlreadyExists
			}
			return err
		}

		// generate refresh token
		refreshToken, plain, err := auth.GenerateRefreshToken(user.ID, s.cfg.RefreshTokenTTL)
		if err != nil {
			return err
		}

		token = &domain.TokenPair{
			RefreshToken: plain,
		}

		return repos.RefreshTokenRepo().Save(ctx, refreshToken)
	})
	if err != nil {
		return nil, err
	}

	// generate access token
	accessToken, exp, err := auth.GenerateAccessToken(user.ID, user.Email, s.privateKey, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	token.UserID = user.ID
	token.AccessToken = accessToken
	token.ExpiresAt = exp

	return token, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	return nil, nil
}

func (s *UserService) Logout(ctx context.Context, token string) error {
	return nil
}

func (s *UserService) RefreshToken(ctx context.Context, token string) (*domain.TokenPair, error) {
	return nil, nil
}
