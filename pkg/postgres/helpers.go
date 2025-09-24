package postgres

import (
	"errors"
	"fmt"

	errs "github.com/deimossy/order-processing-system/pkg/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	CodeUniqueViolation     = "23505"
	CodeForeignKeyViolation = "23503"
)

func CheckUnique(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == CodeUniqueViolation {
		return fmt.Errorf("%w: %s must be unique", errs.ErrAlreadyExists, pgErr.ConstraintName)
	}

	return err
}

func CheckForeignKey(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == CodeForeignKeyViolation {
		return fmt.Errorf("%w: %s violeted", errs.ErrNotFound, pgErr.ConstraintName)
	}

	return err
}
