package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/deimossy/order-processing-system/internal/user/config"
	"github.com/deimossy/order-processing-system/internal/user/domain"
	"github.com/deimossy/order-processing-system/internal/user/usecase"
	errs "github.com/deimossy/order-processing-system/pkg/errors"
	"github.com/deimossy/order-processing-system/pkg/retry"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

type PgRefreshTokenRepo struct {
	db  *sqlx.DB
	cfg config.Config
}

var _ usecase.RefreshTokenRepo = (*PgRefreshTokenRepo)(nil)

func NewPgRefreshTokenRepo(db *sqlx.DB, cfg config.Config) *PgRefreshTokenRepo {
	return &PgRefreshTokenRepo{
		db:  db,
		cfg: cfg,
	}
}

func (pg *PgRefreshTokenRepo) Save(ctx context.Context, token *domain.RefreshToken) error {
	err := retry.Do(ctx, pg.cfg.PgMaxRetries, pg.cfg.PgBackoff, pg.cfg.PgMaxBackoff, func() error {
		timeout, cancel := context.WithTimeout(ctx, pg.cfg.PgQueryTimeout)
		defer cancel()

		_, err := pg.db.NamedExecContext(timeout, saveRefreshTokenQuery, token)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (pg *PgRefreshTokenRepo) GetByTokenHash(ctx context.Context, token string) (*domain.RefreshToken, error) {
	refreshToken := &domain.RefreshToken{}

	err := retry.Do(ctx, pg.cfg.PgMaxRetries, pg.cfg.PgBackoff, pg.cfg.PgMaxBackoff, func() error {
		timeout, cancel := context.WithTimeout(ctx, pg.cfg.PgQueryTimeout)
		defer cancel()

		err := pg.db.GetContext(timeout, refreshToken, getRefreshTokenByTokenHashQuery, token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.ErrNotFound
			}
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return refreshToken, nil
}

func (pg *PgRefreshTokenRepo) RevokeByTokenHash(ctx context.Context, token string) error {
	err := retry.Do(ctx, pg.cfg.PgMaxRetries, pg.cfg.PgBackoff, pg.cfg.PgMaxBackoff, func() error {
		timeout, cancel := context.WithTimeout(ctx, pg.cfg.PgQueryTimeout)
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

		return nil
	})

	return err
}
