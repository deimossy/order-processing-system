package usecase

import (
	"context"

	"github.com/deimossy/order-processing-system/internal/user/domain"
)

type RefreshTokenRepo interface {
	Save(ctx context.Context, token *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, token string) (*domain.RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, token string) error
}
