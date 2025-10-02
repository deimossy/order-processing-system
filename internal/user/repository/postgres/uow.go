package postgres

import (
	"context"
	"database/sql"
	"github.com/deimossy/order-processing-system/internal/user/config"
	"github.com/deimossy/order-processing-system/internal/user/usecase"
	"github.com/deimossy/order-processing-system/pkg/retry"
	"github.com/jmoiron/sqlx"
)

type PgUnitOfWork struct {
	db  *sqlx.DB
	cfg config.Config
}

func NewPgUnitOfWork(db *sqlx.DB, cfg config.Config) *PgUnitOfWork {
	return &PgUnitOfWork{
		db:  db,
		cfg: cfg,
	}
}

func (u *PgUnitOfWork) WithoutTx(ctx context.Context, fn func(ctx context.Context, repos usecase.Repositories) error) error {
	return retry.Do(ctx, u.cfg.PgMaxRetries, u.cfg.PgBackoff, u.cfg.PgMaxBackoff, func() error {
		repos := NewPgRepos(
			NewPgUserRepo(u.db, u.cfg.PgQueryTimeout),
			NewPgRefreshTokenRepo(u.db, u.cfg.PgQueryTimeout),
		)
		return fn(ctx, repos)
	}, retry.ShouldRetryQuery)
}

func (u *PgUnitOfWork) WithTx(ctx context.Context, fn func(ctx context.Context, repos usecase.Repositories) error) error {
	return retry.Do(ctx, u.cfg.PgMaxRetries, u.cfg.PgBackoff, u.cfg.PgMaxBackoff, func() error {
		tx, err := u.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return err
		}
		defer func() {
			_ = tx.Rollback()
		}()

		reposWithTx := NewPgRepos(
			NewPgUserRepo(tx, u.cfg.PgQueryTimeout),
			NewPgRefreshTokenRepo(tx, u.cfg.PgQueryTimeout),
		)

		if err = fn(ctx, reposWithTx); err != nil {
			return err
		}

		return tx.Commit()
	}, retry.ShouldRetryTx)
}
