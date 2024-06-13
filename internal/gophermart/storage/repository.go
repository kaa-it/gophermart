package storage

import (
	"context"
	"errors"
	"fmt"
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

	dbpool, err := pgxpool.New(context.Background(), config.DSN)
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

	return err
}
