package config

import (
	"log/slog"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerGRPCPort        string        `env:"USER_SERVICE_GRPC_PORT"`
	RedisAddr             string        `env:"USER_SERVICE_REDIS_ADDR"`
	RedisPingTimeout      time.Duration `env:"USER_SERVICE_REDIS_PING_TIMEOUT"`
	PgDsn                 string        `env:"POSTGRES_DSN"`
	PgPoolConnMaxLifetime time.Duration `env:"POSTGRES_POOL_CONN_MAX_LIFETIME"`
	PgPoolMaxIdleConns    int           `env:"POSTGRES_POOL_MAX_IDLE_CONNS"`
	PgPoolMaxOpenConns    int           `env:"POSTGRES_POOL_MAX_OPEN_CONNS"`
	PgPingTimeout         time.Duration `env:"POSTGRES_PING_TIMEOUT"`
	PgMaxRetries          int           `env:"POSTGRES_MAX_RETRIES"`
	PgBackoff             time.Duration `env:"POSTGRES_BACKOFF"`
	PgMaxBackoff          time.Duration `env:"POSTGRES_MAX_BACKOFF"`
	PgQueryTimeout        time.Duration `env:"POSTGRES_QUERY_TIMEOUT"`
	KafkaBackoff          time.Duration `env:"KAFKA_BACKOFF"`
	KafkaMaxWait          time.Duration `env:"KAFKA_MAX_WAIT"`
	BcryptCost            int           `env:"BCRYPT_COST"`
	RSAPrivateKeyPath     string        `env:"RSA_PRIVATE_KEY_PATH"`
	RSAPublicKeyPath      string        `env:"RSA_PUBLIC_KEY_PATH"`
}

func NewConfig(logger *slog.Logger) Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		logger.Warn("fail to parse config, used default values",
			slog.String("error", err.Error()),
		)
	}

	logger.Debug("env variables",
		slog.Any("config", cfg),
	)

	return cfg
}
