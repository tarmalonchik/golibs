package postgres

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

var registerOnce sync.Once

func New(cfg Config) (*Postgres, error) {
	if cfg.GetPGDSN() == "" {
		return nil, nil
	}

	registerOnce.Do(func() {
		sql.Register("instrumented-postgres", pq.Driver{})
	})

	dbSQL, err := sql.Open("instrumented-postgres", cfg.GetPGDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	dbSQL.SetMaxIdleConns(cfg.GetPGMaxIdleConns())
	dbSQL.SetMaxOpenConns(cfg.GetPGMaxOpenConns())
	dbSQL.SetConnMaxLifetime(cfg.GetPGTimeout())

	return &Postgres{db: dbSQL}, nil
}

func (p *Postgres) GetDB() *sql.DB {
	return p.db
}
