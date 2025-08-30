package postgres

import (
	"context"
	"errors"
	"github.com/deimossy/order-processing-system/internal/user/config"
	"github.com/deimossy/order-processing-system/internal/user/domain"
	"github.com/deimossy/order-processing-system/internal/user/usecase"
	"github.com/deimossy/order-processing-system/pkg/retry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrCodeUniqueViolation = "23505"
)

type PgUserRepo struct {
	db  *sqlx.DB
	cfg config.Config
}

var _ usecase.UserRepo = (*PgUserRepo)(nil)

func NewPgUserRepo(db *sqlx.DB, cfg config.Config) *PgUserRepo {
	return &PgUserRepo{
		db:  db,
		cfg: cfg,
	}
}

func (pg *PgUserRepo) SaveUser(ctx context.Context, user *domain.User) error {
	err := retry.Do(ctx, pg.cfg.PgMaxRetries, pg.cfg.PgBackoff, func() error {
		timeout, cancel := context.WithTimeout(ctx, pg.cfg.PgQueryTimeout)
		defer cancel()

		_, err := pg.db.NamedExecContext(timeout, saveUserQuery, user)
		if err != nil {
			return checkUnique(err)
		}

		return nil
	})

	return err
}

func (pg *PgUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}

	err := retry.Do(ctx, pg.cfg.PgMaxRetries, pg.cfg.PgBackoff, func() error {
		timeout, cancel := context.WithTimeout(ctx, pg.cfg.PgQueryTimeout)
		defer cancel()

		err := pg.db.GetContext(timeout, user, getUserByEmailQuery, email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrUserNotFound
			}
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PgUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}

	err := retry.Do(ctx, pg.cfg.PgMaxRetries, pg.cfg.PgBackoff, func() error {
		timeout, cancel := context.WithTimeout(ctx, pg.cfg.PgQueryTimeout)
		defer cancel()

		err := pg.db.GetContext(timeout, user, getUserByIDQuery, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrUserNotFound
			}
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PgUserRepo) DeleteByID(ctx context.Context, id string) error {
	err := retry.Do(ctx, pg.cfg.PgMaxRetries, pg.cfg.PgBackoff, func() error {
		timeout, cancel := context.WithTimeout(ctx, pg.cfg.PgQueryTimeout)
		defer cancel()

		r, err := pg.db.ExecContext(timeout, deleteUserByIDQuery, id)
		if err != nil {
			return err
		}

		rows, err := r.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return ErrUserNotFound
		}

		return nil
	})

	return err
}

func checkUnique(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == ErrCodeUniqueViolation {
		return ErrUserAlreadyExists
	}

	return err
}
