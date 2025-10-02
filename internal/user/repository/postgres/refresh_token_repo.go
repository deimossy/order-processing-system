package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/deimossy/order-processing-system/pkg/postgres"
	"time"

	"github.com/deimossy/order-processing-system/internal/user/domain"
	errs "github.com/deimossy/order-processing-system/pkg/errors"
)

type PgRefreshTokenRepo struct {
	db           postgres.DBTX
	queryTimeout time.Duration
}

func NewPgRefreshTokenRepo(db postgres.DBTX, queryTimeout time.Duration) *PgRefreshTokenRepo {
	return &PgRefreshTokenRepo{
		db:           db,
		queryTimeout: queryTimeout,
	}
}

func (pg *PgRefreshTokenRepo) Save(ctx context.Context, token *domain.RefreshToken) error {
	timeout, cancel := context.WithTimeout(ctx, pg.queryTimeout)
	defer cancel()

	_, err := pg.db.NamedExecContext(timeout, saveRefreshTokenQuery, token)
	if err != nil {
		return err
	}

	return err
}

func (pg *PgRefreshTokenRepo) GetByTokenHash(ctx context.Context, token string) (*domain.RefreshToken, error) {
	refreshToken := &domain.RefreshToken{}

	timeout, cancel := context.WithTimeout(ctx, pg.queryTimeout)
	defer cancel()

	err := pg.db.GetContext(timeout, refreshToken, getRefreshTokenByTokenHashQuery, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return refreshToken, nil
}

func (pg *PgRefreshTokenRepo) RevokeByTokenHash(ctx context.Context, token string) error {
	timeout, cancel := context.WithTimeout(ctx, pg.queryTimeout)
	defer cancel()

	r, err := pg.db.ExecContext(timeout, revokeRefreshTokenByTokenHashQuery, time.Now().UTC(), token)
	if err != nil {
		return err
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errs.ErrNotFound
	}

	return err
}
