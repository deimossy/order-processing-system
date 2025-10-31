package usecase

import (
	"context"
	"crypto/rsa"
	"errors"
	"time"

	"github.com/deimossy/order-processing-system/internal/user/auth"
	"github.com/deimossy/order-processing-system/internal/user/config"
	"github.com/deimossy/order-processing-system/internal/user/domain"
	errs "github.com/deimossy/order-processing-system/pkg/errors"
)

type RefreshTokenRepo interface {
	Save(ctx context.Context, token *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, token string) (*domain.RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, token string) error
	RevokeAllByUserId(ctx context.Context, userId string) error
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

func (s *UserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var user *domain.User

	err := s.uow.WithoutTx(ctx, func(ctx context.Context, repos Repositories) error {
		userDB, err := repos.UserRepo().GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return errs.ErrUserNotFound
			}
			return err
		}
		user = userDB
		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Register(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	var (
		token *domain.TokenPair
		user  *domain.User
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
			UserID:       user.ID,
		}

		return repos.RefreshTokenRepo().Save(ctx, refreshToken)
	})
	if err != nil {
		return nil, err
	}

	// generate access token
	accessToken, exp, err := auth.GenerateAccessToken(user.ID, email, s.privateKey, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	token.AccessToken = accessToken
	token.ExpiresAt = exp

	return token, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	var token *domain.TokenPair

	err := s.uow.WithTx(ctx, func(ctx context.Context, repos Repositories) error {
		// check user
		user, err := repos.UserRepo().GetByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return errs.ErrUserNotFound
			}
			return err
		}

		// validate password
		if !auth.CheckPasswordHash(password, user.PasswordHash) {
			return errs.ErrPasswordMismatch
		}

		// generate new refresh token
		refreshToken, plain, err := auth.GenerateRefreshToken(user.ID, s.cfg.RefreshTokenTTL)
		if err != nil {
			return err
		}

		token = &domain.TokenPair{
			RefreshToken: plain,
			UserID:       user.ID,
		}

		// revoke all old refresh tokens
		err = repos.RefreshTokenRepo().RevokeAllByUserId(ctx, user.ID)
		if err != nil && !errors.Is(err, errs.ErrNotFound) {
			return err
		}

		return repos.RefreshTokenRepo().Save(ctx, refreshToken)
	})

	if err != nil {
		return nil, err
	}

	// generate new access token
	accessToken, exp, err := auth.GenerateAccessToken(token.UserID, email, s.privateKey, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	token.AccessToken = accessToken
	token.ExpiresAt = exp

	return token, nil
}

func (s *UserService) Logout(ctx context.Context, plainRefreshToken string) error {
	hashedRefreshToken := auth.HashToken(plainRefreshToken)

	err := s.uow.WithoutTx(ctx, func(ctx context.Context, repos Repositories) error {
		// revoke all refresh tokens
		err := repos.RefreshTokenRepo().RevokeByTokenHash(ctx, hashedRefreshToken)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return errs.ErrRefreshTokenNotFound
			}
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) Refresh(ctx context.Context, plainRefreshToken string) (*domain.TokenPair, error) {
	hashedRefreshToken := auth.HashToken(plainRefreshToken)
	var (
		token *domain.TokenPair
		user  *domain.User
	)

	err := s.uow.WithTx(ctx, func(ctx context.Context, repos Repositories) error {
		// get refresh token
		refreshToken, err := repos.RefreshTokenRepo().GetByTokenHash(ctx, hashedRefreshToken)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return errs.ErrRefreshTokenNotFound
			}
			return err
		}

		// validating refresh token
		if refreshToken.RevokedAt != nil {
			return errs.ErrRefreshTokenRevoked
		}

		if refreshToken.ExpiresAt.Before(time.Now()) {
			return errs.ErrRefreshTokenExpired
		}

		// get user
		user, err = repos.UserRepo().GetByID(ctx, refreshToken.UserID)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return errs.ErrUserNotFound
			}
			return err
		}

		// generate new refresh token
		newRefreshToken, newPlainRefreshToken, err := auth.GenerateRefreshToken(user.ID, s.cfg.RefreshTokenTTL)
		if err != nil {
			return err
		}

		// saving new refresh token
		err = repos.RefreshTokenRepo().Save(ctx, newRefreshToken)
		if err != nil {
			return err
		}

		// revoke old refresh token
		err = repos.RefreshTokenRepo().RevokeByTokenHash(ctx, hashedRefreshToken)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return errs.ErrRefreshTokenNotFound
			}
			return err
		}

		token = &domain.TokenPair{
			RefreshToken: newPlainRefreshToken,
			UserID:       user.ID,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// generate access token
	accessToken, exp, err := auth.GenerateAccessToken(user.ID, user.Email, s.privateKey, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	token.AccessToken = accessToken
	token.ExpiresAt = exp

	return token, nil
}
