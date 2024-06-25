package postgres

import (
	"context"
	"errors"
	"fmt"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DSN string
}

type Storage struct {
	config *Config
	dbpool *pgxpool.Pool
}

func NewStorage(config *Config) (*Storage, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}

	pgConfig, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	pgConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &Storage{
		config,
		dbpool,
	}, nil
}

func (s *Storage) Close() {
	s.dbpool.Close()
}

func (s *Storage) Initialize(ctx context.Context) error {
	_, err := s.dbpool.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS users
          (
			id        SERIAL PRIMARY KEY,
			login     TEXT NOT NULL UNIQUE,
			password  TEXT NOT NULL,
			current   DECIMAL DEFAULT 0.0,
			withdrawn DECIMAL DEFAULT 0.0
          );`,
	)

	if err != nil {
		return err
	}

	_, err = s.dbpool.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS sessions
		(
			id            SERIAL PRIMARY KEY,
			user_id       INT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
			refresh_token TEXT      NOT NULL,
			expired       TIMESTAMP NOT NULL
		);`,
	)

	if err != nil {
		return err
	}

	_, err = s.dbpool.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS orders
        (
		    number TEXT PRIMARY KEY,
		    user_id       INT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
		    status TEXT NOT NULL,
		    accrual DECIMAL,
		    uploaded_at TIMESTAMP NOT NULL
	    );`,
	)

	if err != nil {
		return err
	}

	_, err = s.dbpool.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS withdrawals
		(
			number TEXT PRIMARY KEY,
			user_id       INT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
			sum DECIMAL NOT NULL,
			processed_at TIMESTAMP NOT NULL
		);`,
	)

	return err
}
