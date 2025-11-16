package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Host     string 
	Port     int   
	Username string 
	Password string 
	DBName   string 
	SSLMode  string 
	Timeout  int64
}

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgresWithConfig(ctx context.Context, cfg PostgresConfig) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Postgres: %w", err)
	}

	return &Postgres{Pool: pool}, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
