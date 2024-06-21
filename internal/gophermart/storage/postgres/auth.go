package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kaa-it/gophermart/internal/gophermart/auth"
	"github.com/shopspring/decimal"
	"time"
)

const (
	_refreshTokenLifetime = 2 * time.Hour
)

type user struct {
	id        int64
	login     string
	password  string
	currency  decimal.Decimal
	withdrawn decimal.Decimal
}

func (s *Storage) CreateUser(ctx context.Context, user auth.User, refreshToken string) (int64, error) {
	transaction, err := s.dbpool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}

	defer transaction.Rollback(ctx)

	var userId int64

	err = transaction.QueryRow(
		ctx,
		"INSERT INTO users (login, password) VALUES (@login, @password) RETURNING id",
		pgx.NamedArgs{
			"login":    user.Login,
			"password": user.Password,
		},
	).Scan(&userId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, auth.ErrInvalidUser
		}

		return 0, err
	}

	_, err = transaction.Exec(
		ctx,
		"INSERT INTO sessions (user_id, refresh_token, expired) VALUES (@user_id, @refresh_token, @expired)",
		pgx.NamedArgs{
			"user_id":       userId,
			"refresh_token": refreshToken,
			"expired":       time.Now().Add(_refreshTokenLifetime),
		},
	)

	if err != nil {
		return 0, err
	}

	if err := transaction.Commit(ctx); err != nil {
		return 0, err
	}

	return userId, nil
}
