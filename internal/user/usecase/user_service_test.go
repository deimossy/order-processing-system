package usecase_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/deimossy/order-processing-system/internal/user/auth"
	"github.com/deimossy/order-processing-system/internal/user/config"
	"github.com/deimossy/order-processing-system/internal/user/domain"
	"github.com/deimossy/order-processing-system/internal/user/usecase"
	"github.com/deimossy/order-processing-system/internal/user/usecase/mocks"
	errs "github.com/deimossy/order-processing-system/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*usecase.UserService, *mocks.MockRefreshTokenRepo, *mocks.MockUserRepo) {
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockRefreshTokenRepo := mocks.NewMockRefreshTokenRepo(t)

	mockRepositories := mocks.NewMockRepositories(t)
	mockRepositories.On("UserRepo").Return(mockUserRepo).Maybe()
	mockRepositories.On("RefreshTokenRepo").Return(mockRefreshTokenRepo).Maybe()

	mockUOW := mocks.NewMockUnitOfWork(t)

	mockUOW.
		On("WithTx", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(ctx context.Context, repos usecase.Repositories) error) error {
			return fn(ctx, mockRepositories) // для исполнения функций замоканными репами
		}).
		Maybe()

	mockUOW.
		On("WithoutTx", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(ctx context.Context, repos usecase.Repositories) error) error {
			return fn(ctx, mockRepositories) // для исполнения функций замоканными репами
		}).
		Maybe()

	cfg := config.Config{
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour,
		BcryptCost:      10,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	userService := usecase.NewUserService(cfg, mockUOW, privateKey)

	return userService, mockRefreshTokenRepo, mockUserRepo
}

func TestUserService_Register_Success(t *testing.T) {
	srv, mockRefreshTokenRepo, mockUserRepo := setup(t)

	ctx := context.Background()
	email := "test@example.com"
	password := "strongpassword123"

	mockUserRepo.
		On("GetByEmail", ctx, email).
		Return(nil, errs.ErrNotFound).
		Once()

	mockUserRepo.
		On("Save", ctx, mock.AnythingOfType("*domain.User")).
		Return(nil).
		Once()

	mockRefreshTokenRepo.
		On("Save", ctx, mock.AnythingOfType("*domain.RefreshToken")).
		Return(nil).
		Once()

	tokenPair, err := srv.Register(ctx, email, password)

	require.NoError(t, err)
	require.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.RefreshToken)
}

func TestUserService_Register_Failure(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"
	password := "strongpassword123"

	dbErr := errors.New("db error")

	tCases := []struct {
		name      string
		mockSetup func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo)
		expErr    error
	}{
		{
			name: "user_already_registered_1",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByEmail(ctx, email).Return(&domain.User{}, nil).Once()
			},
			expErr: errs.ErrUserAlreadyExists,
		},
		{
			name: "db_error_on_get_user_by_email",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByEmail(ctx, email).Return(nil, dbErr).Once()
			},
			expErr: dbErr,
		},
		{
			name: "user_already_registered_2",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByEmail(ctx, email).Return(nil, errs.ErrNotFound).Once()
				userRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.User")).Return(errs.ErrAlreadyExists).Once()
			},
			expErr: errs.ErrUserAlreadyExists,
		},
		{
			name: "db_error_on_user_save",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByEmail(ctx, email).Return(nil, errs.ErrNotFound).Once()
				userRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.User")).Return(dbErr).Once()
			},
			expErr: dbErr,
		},
		{
			name: "db_error_on_refresh_token_save",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByEmail(ctx, email).Return(nil, errs.ErrNotFound).Once()
				userRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.User")).Return(nil).Once()
				refreshTokenRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(dbErr).Once()
			},
			expErr: dbErr,
		},
	}

	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			srv, mockRefreshTokenRepo, mockUserRepo := setup(t)

			tc.mockSetup(mockUserRepo, mockRefreshTokenRepo)

			tokenPair, err := srv.Register(ctx, email, password)

			require.True(t, errors.Is(err, tc.expErr), "expected error %v but got %v", tc.expErr, err)
			require.Nil(t, tokenPair)
		})
	}
}

func TestUserService_Register_GenerateAccessTokenFailure(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockTokenRepo := mocks.NewMockRefreshTokenRepo(t)
	mockRepos := mocks.NewMockRepositories(t)
	mockRepos.On("UserRepo").Return(mockUserRepo).Maybe()
	mockRepos.On("RefreshTokenRepo").Return(mockTokenRepo).Maybe()
	mockUOW := mocks.NewMockUnitOfWork(t)
	mockUOW.On("WithTx", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(context.Context, usecase.Repositories) error) error {
			mockUserRepo.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(nil, errs.ErrNotFound)
			mockUserRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)
			mockTokenRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)
			return fn(ctx, mockRepos)
		})

	cfg := config.Config{}

	userService := usecase.NewUserService(cfg, mockUOW, nil)

	tokenPair, err := userService.Register(context.Background(), "test@example.com", "password")

	require.Error(t, err)
	assert.Nil(t, tokenPair)
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	srv, _, mockUserRepo := setup(t)

	ctx := context.Background()

	mockUserRepo.
		EXPECT().
		GetByID(ctx, mock.AnythingOfType("string")).
		Return(&domain.User{}, nil).
		Once()

	user, err := srv.GetUserByID(ctx, uuid.New().String())

	require.NoError(t, err)
	require.NotNil(t, user)
}

func TestUserService_GetUserByID_Failure(t *testing.T) {
	ctx := context.Background()

	dbErr := errors.New("db error")

	tCases := []struct {
		name      string
		mockSetup func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo)
		expErr    error
	}{
		{
			name: "user_not_found",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByID(ctx, mock.AnythingOfType("string")).Return(nil, errs.ErrNotFound).Once()
			},
			expErr: errs.ErrUserNotFound,
		},
		{
			name: "db_error_on_get_by_id",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByID(ctx, mock.AnythingOfType("string")).Return(nil, dbErr).Once()
			},
			expErr: dbErr,
		},
	}

	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			srv, mockRefreshTokenRepo, mockUserRepo := setup(t)

			tc.mockSetup(mockUserRepo, mockRefreshTokenRepo)

			user, err := srv.GetUserByID(ctx, uuid.New().String())

			require.True(t, errors.Is(err, tc.expErr), "expected error %v but got %v", tc.expErr, err)
			require.Nil(t, user)
		})
	}
}

func TestUserService_Login_Success(t *testing.T) {
	srv, mockRefreshTokenRepo, mockUserRepo := setup(t)

	ctx := context.Background()
	email := "test@example.com"
	password := "strongpassword123"
	hashPassword, _ := auth.HashPassword(password, 4)

	mockUserRepo.
		EXPECT().
		GetByEmail(ctx, email).
		Return(&domain.User{PasswordHash: hashPassword}, nil).
		Once()

	mockRefreshTokenRepo.
		EXPECT().
		RevokeAllByUserId(ctx, mock.AnythingOfType("string")).
		Return(nil).
		Once()

	mockRefreshTokenRepo.
		EXPECT().
		Save(ctx, mock.AnythingOfType("*domain.RefreshToken")).
		Return(nil).
		Once()

	tokenPair, err := srv.Login(ctx, email, password)

	require.NoError(t, err)
	require.NotEmpty(t, tokenPair)
}

func TestUserService_Login_Failure(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"
	password := "strongpassword123"
	hashPassword, _ := auth.HashPassword(password, 4)

	dbErr := errors.New("db error")

	tCases := []struct {
		name      string
		mockSetup func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo)
		expErr    error
	}{
		{
			name: "user_not_found",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByEmail(ctx, mock.AnythingOfType("string")).Return(nil, errs.ErrNotFound).Once()
			},
			expErr: errs.ErrUserNotFound,
		},
		{
			name: "db_error_on_get_by_email",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByEmail(ctx, mock.AnythingOfType("string")).Return(nil, dbErr).Once()
			},
			expErr: dbErr,
		},
		{
			name: "db_error_on_revoke_all_by_user_id",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByEmail(ctx, mock.AnythingOfType("string")).Return(&domain.User{PasswordHash: hashPassword}, nil).Once()
				refreshTokenRepo.EXPECT().RevokeAllByUserId(ctx, mock.AnythingOfType("string")).Return(dbErr).Once()
			},
			expErr: dbErr,
		},
		{
			name: "db_error_on_refresh_token_save",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				userRepo.EXPECT().GetByEmail(ctx, mock.AnythingOfType("string")).Return(&domain.User{PasswordHash: hashPassword}, nil).Once()
				refreshTokenRepo.EXPECT().RevokeAllByUserId(ctx, mock.AnythingOfType("string")).Return(nil).Once()
				refreshTokenRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(dbErr).Once()
			},
			expErr: dbErr,
		},
		{
			name: "invalid_pass",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				incorrectPassword := "incorrect_pass"
				incorrectHashPassword, _ := auth.HashPassword(incorrectPassword, 4)
				userRepo.EXPECT().GetByEmail(ctx, mock.AnythingOfType("string")).Return(&domain.User{PasswordHash: incorrectHashPassword}, nil).Once()
			},
			expErr: errs.ErrPasswordMismatch,
		},
	}

	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			srv, mockRefreshTokenRepo, mockUserRepo := setup(t)

			tc.mockSetup(mockUserRepo, mockRefreshTokenRepo)

			tokenPair, err := srv.Login(ctx, email, password)

			require.True(t, errors.Is(err, tc.expErr), "expected error %v but got %v", tc.expErr, err)
			require.Nil(t, tokenPair)
		})
	}
}

func TestUserService_Login_GenerateAccessToken_Failure(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockTokenRepo := mocks.NewMockRefreshTokenRepo(t)
	mockRepos := mocks.NewMockRepositories(t)
	mockRepos.On("UserRepo").Return(mockUserRepo).Maybe()
	mockRepos.On("RefreshTokenRepo").Return(mockTokenRepo).Maybe()
	mockUOW := mocks.NewMockUnitOfWork(t)
	hashPass, _ := auth.HashPassword("password", 4)
	mockUOW.On("WithTx", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(context.Context, usecase.Repositories) error) error {
			mockUserRepo.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(&domain.User{PasswordHash: hashPass}, nil).Once()
			mockTokenRepo.EXPECT().RevokeAllByUserId(mock.Anything, mock.Anything).Return(nil).Once()
			mockTokenRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)
			return fn(ctx, mockRepos)
		})

	cfg := config.Config{}

	userService := usecase.NewUserService(cfg, mockUOW, nil)

	tokenPair, err := userService.Login(context.Background(), "test@example.com", "password")

	require.Error(t, err)
	assert.Nil(t, tokenPair)
}

func TestUserService_Logout_Success(t *testing.T) {
	srv, mockRefreshTokenRepo, _ := setup(t)

	plainRefreshToken := "refreshToken"
	ctx := context.Background()

	mockRefreshTokenRepo.
		EXPECT().
		RevokeByTokenHash(ctx, mock.AnythingOfType("string")).
		Return(nil).
		Once()

	err := srv.Logout(ctx, plainRefreshToken)

	require.NoError(t, err)
}

func TestUserService_Logout_Failure(t *testing.T) {
	refreshToken := "refreshToken"
	ctx := context.Background()

	dbErr := errors.New("db error")

	tCases := []struct {
		name      string
		mockSetup func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo)
		expErr    error
	}{
		{
			name: "refresh_token_not_found",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().RevokeByTokenHash(ctx, mock.AnythingOfType("string")).Return(errs.ErrNotFound).Once()
			},
			expErr: errs.ErrRefreshTokenNotFound,
		},
		{
			name: "db_error_on_revoke_by_token_hash",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().RevokeByTokenHash(ctx, mock.AnythingOfType("string")).Return(dbErr).Once()
			},
			expErr: dbErr,
		},
	}

	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			srv, mockRefreshTokenRepo, mockUserRepo := setup(t)

			tc.mockSetup(mockUserRepo, mockRefreshTokenRepo)

			err := srv.Logout(ctx, refreshToken)

			require.True(t, errors.Is(err, tc.expErr), "expected error %v but got %v", tc.expErr, err)
		})
	}
}

func TestUserService_Refresh_Success(t *testing.T) {
	srv, mockRefreshTokenRepo, mockUserRepo := setup(t)

	plainRefreshToken := "refreshToken"
	ctx := context.Background()

	mockRefreshTokenRepo.
		EXPECT().
		GetByTokenHash(ctx, mock.AnythingOfType("string")).
		Return(&domain.RefreshToken{
			RevokedAt: nil,
			ExpiresAt: time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC),
		}, nil).
		Once()

	mockUserRepo.
		EXPECT().
		GetByID(ctx, mock.AnythingOfType("string")).
		Return(&domain.User{}, nil).
		Once()

	mockRefreshTokenRepo.
		EXPECT().
		Save(ctx, mock.AnythingOfType("*domain.RefreshToken")).
		Return(nil).
		Once()

	mockRefreshTokenRepo.
		EXPECT().
		RevokeByTokenHash(ctx, mock.AnythingOfType("string")).
		Return(nil).
		Once()

	tokenPair, err := srv.Refresh(ctx, plainRefreshToken)

	require.NoError(t, err)
	require.NotEmpty(t, tokenPair)
}

func TestUserService_Refresh_Failure(t *testing.T) {
	refreshToken := "refreshToken"
	ctx := context.Background()

	dbErr := errors.New("db error")

	tCases := []struct {
		name      string
		mockSetup func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo)
		expErr    error
	}{
		{
			name: "refresh_token_not_found",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().GetByTokenHash(ctx, mock.AnythingOfType("string")).Return(nil, errs.ErrNotFound).Once()
			},
			expErr: errs.ErrRefreshTokenNotFound,
		},
		{
			name: "db_err_on_get_by_token_hash",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().GetByTokenHash(ctx, mock.AnythingOfType("string")).Return(nil, dbErr).Once()
			},
			expErr: dbErr,
		},
		{
			name: "user_not_found",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().GetByTokenHash(ctx, mock.AnythingOfType("string")).Return(&domain.RefreshToken{
					RevokedAt: nil,
					ExpiresAt: time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC),
				}, nil).Once()
				userRepo.EXPECT().GetByID(ctx, mock.AnythingOfType("string")).Return(nil, errs.ErrNotFound)
			},
			expErr: errs.ErrUserNotFound,
		},
		{
			name: "db_err_on_get_by_id",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().GetByTokenHash(ctx, mock.AnythingOfType("string")).Return(&domain.RefreshToken{
					RevokedAt: nil,
					ExpiresAt: time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC),
				}, nil).Once()
				userRepo.EXPECT().GetByID(ctx, mock.AnythingOfType("string")).Return(nil, dbErr)
			},
			expErr: dbErr,
		},
		{
			name: "db_err_on_refresh_token_save",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().GetByTokenHash(ctx, mock.AnythingOfType("string")).Return(&domain.RefreshToken{
					RevokedAt: nil,
					ExpiresAt: time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC),
				}, nil).Once()
				userRepo.EXPECT().GetByID(ctx, mock.AnythingOfType("string")).Return(&domain.User{}, nil)
				refreshTokenRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(dbErr)
			},
			expErr: dbErr,
		},
		{
			name: "refresh_token_not_found",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().GetByTokenHash(ctx, mock.AnythingOfType("string")).Return(&domain.RefreshToken{
					RevokedAt: nil,
					ExpiresAt: time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC),
				}, nil).Once()
				userRepo.EXPECT().GetByID(ctx, mock.AnythingOfType("string")).Return(&domain.User{}, nil)
				refreshTokenRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)
				refreshTokenRepo.EXPECT().RevokeByTokenHash(ctx, mock.AnythingOfType("string")).Return(errs.ErrNotFound)
			},
			expErr: errs.ErrRefreshTokenNotFound,
		}, {
			name: "db_err_on_revoke_by_token_hash",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().GetByTokenHash(ctx, mock.AnythingOfType("string")).Return(&domain.RefreshToken{
					RevokedAt: nil,
					ExpiresAt: time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC),
				}, nil).Once()
				userRepo.EXPECT().GetByID(ctx, mock.AnythingOfType("string")).Return(&domain.User{}, nil)
				refreshTokenRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)
				refreshTokenRepo.EXPECT().RevokeByTokenHash(ctx, mock.AnythingOfType("string")).Return(dbErr)
			},
			expErr: dbErr,
		},
		{
			name: "refresh_token_revoked",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().GetByTokenHash(ctx, mock.AnythingOfType("string")).Return(&domain.RefreshToken{
					RevokedAt: &time.Time{},
				}, nil).Once()
			},
			expErr: errs.ErrRefreshTokenRevoked,
		},
		{
			name: "refresh_token_expired",
			mockSetup: func(userRepo *mocks.MockUserRepo, refreshTokenRepo *mocks.MockRefreshTokenRepo) {
				refreshTokenRepo.EXPECT().GetByTokenHash(ctx, mock.AnythingOfType("string")).Return(&domain.RefreshToken{
					RevokedAt: nil,
					ExpiresAt: time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC),
				}, nil).Once()
			},
			expErr: errs.ErrRefreshTokenExpired,
		},
	}

	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			srv, mockRefreshToken, mockUser := setup(t)
			tc.mockSetup(mockUser, mockRefreshToken)

			tokenPair, err := srv.Refresh(ctx, refreshToken)

			require.True(t, errors.Is(err, tc.expErr), "expected error %v but got %v", tc.expErr, err)
			require.Nil(t, tokenPair)
		})
	}
}

func TestUserService_Refresh_GenerateAccessToken_Failure(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockTokenRepo := mocks.NewMockRefreshTokenRepo(t)
	mockRepos := mocks.NewMockRepositories(t)
	mockRepos.On("UserRepo").Return(mockUserRepo).Maybe()
	mockRepos.On("RefreshTokenRepo").Return(mockTokenRepo).Maybe()
	mockUOW := mocks.NewMockUnitOfWork(t)
	mockUOW.On("WithTx", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(context.Context, usecase.Repositories) error) error {
			mockTokenRepo.EXPECT().GetByTokenHash(ctx, mock.AnythingOfType("string")).Return(&domain.RefreshToken{
				RevokedAt: nil,
				ExpiresAt: time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC),
			}, nil).Once()
			mockUserRepo.EXPECT().GetByID(ctx, mock.AnythingOfType("string")).Return(&domain.User{}, nil).Once()
			mockTokenRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Once()
			mockTokenRepo.EXPECT().RevokeByTokenHash(ctx, mock.AnythingOfType("string")).Return(nil).Once()
			return fn(ctx, mockRepos)
		})

	cfg := config.Config{}

	userService := usecase.NewUserService(cfg, mockUOW, nil)

	tokenPair, err := userService.Refresh(context.Background(), "refreshToken")

	require.Error(t, err)
	assert.Nil(t, tokenPair)
}
