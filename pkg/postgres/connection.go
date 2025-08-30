package postgres

import (
	"context"
	"log"

	"github.com/deimossy/order-processing-system/internal/user/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPgClient(ctx context.Context, cfg config.Config) *sqlx.DB {
	db, err := sqlx.Open("pgx", cfg.PgDsn)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(cfg.PgPoolMaxOpenConns)
	db.SetMaxIdleConns(cfg.PgPoolMaxIdleConns)
	db.SetConnMaxLifetime(cfg.PgPoolConnMaxLifetime)

	timeout, cancel := context.WithTimeout(ctx, cfg.PgPingTimeout)
	defer cancel()

	if err = db.PingContext(timeout); err != nil {
		log.Fatal(err)
	}

	return db
}
