package postgres

import (
	"fmt"
	"time"
)

type Config struct {
	PgProto          string        `env:"POSTGRES_PROTO" envDefault:"tcp"`
	PgURL            string        `env:"POSTGRES_URL" envDefault:""`
	PgAddress        string        `env:"POSTGRES_ADDRESS" envDefault:"localhost"`
	PgPort           string        `env:"POSTGRES_PORT" envDefault:""`
	PgDB             string        `env:"POSTGRES_DB" envDefault:""`
	PgUser           string        `env:"POSTGRES_USER" envDefault:""`
	PgPassword       string        `env:"POSTGRES_PASS" envDefault:""`
	PgSSLMode        string        `env:"POSTGRES_SSL_MODE" envDefault:"verify-full"`
	PgMigrationsPath string        `env:"POSTGRES_MIGRATIONSPATH" envDefault:"file:migrations"`
	PgTimeout        time.Duration `env:"POSTGRES_TIMEOUT" envDefault:"5s"`
	PGMaxIdleConns   int           `env:"POSTGRES_MAX_IDLE_CONNS" envDefault:"10"`
	PGMaxOpenConns   int           `env:"POSTGRES_MAX_OPEN_CONNS" envDefault:"15"`
}

func (c *Config) GetPGMigrationsPath() string { return c.PgMigrationsPath }
func (c *Config) GetPGMaxIdleConns() int      { return c.PGMaxIdleConns }
func (c *Config) GetPGMaxOpenConns() int      { return c.PGMaxOpenConns }
func (c *Config) GetPGTimeout() time.Duration { return c.PgTimeout }

func (c *Config) GetPGAddress() string {
	return c.PgAddress + ":" + c.PgPort
}

func (c *Config) GetPGDSN() string {
	pgURL := c.PgURL
	if pgURL == "" {
		pgURL = fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=%s",
			c.PgUser,
			c.PgPassword,
			c.GetPGAddress(),
			c.PgDB,
			c.PgSSLMode,
		)
	}
	return pgURL
}
