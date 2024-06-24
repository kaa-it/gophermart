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

	var userID int64

	err = transaction.QueryRow(
		ctx,
		"INSERT INTO users (login, password) VALUES (@login, @password) RETURNING id",
		pgx.NamedArgs{
			"login":    user.Login,
			"password": user.Password,
		},
	).Scan(&userID)

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
			"user_id":       userID,
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

	return userID, nil
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*auth.User, error) {
	var res user

	err := s.dbpool.QueryRow(
		ctx,
		"SELECT * FROM users WHERE login = @login",
		pgx.NamedArgs{
			"login": login,
		},
	).Scan(&res.id, &res.login, &res.password, &res.currency, &res.withdrawn)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}

		return nil, err
	}

	u := &auth.User{
		ID:        res.id,
		Login:     res.login,
		Password:  res.password,
		Currency:  res.currency,
		Withdrawn: res.withdrawn,
	}

	return u, nil
}

func (s *Storage) CreateSession(ctx context.Context, userID int64, refreshToken string) error {
	_, err := s.dbpool.Exec(
		ctx,
		"INSERT INTO sessions (user_id, refresh_token, expired) VALUES (@user_id, @refresh_token, @expired)",
		pgx.NamedArgs{
			"user_id":       userID,
			"refresh_token": refreshToken,
			"expired":       time.Now().Add(_refreshTokenLifetime),
		},
	)

	return err
}

func (s *Storage) RemoveExpiredSessions(ctx context.Context) error {
	_, err := s.dbpool.Exec(
		ctx,
		"DELETE FROM sessions WHERE expired < NOW()",
	)

	return err
}

func (s *Storage) GetUserIDBySessionToken(ctx context.Context, refreshToken string) (int64, error) {
	var userID int64

	err := s.dbpool.QueryRow(
		ctx,
		"SELECT user_id FROM sessions WHERE refresh_token = @refresh_token",
		pgx.NamedArgs{
			"refresh_token": refreshToken,
		},
	).Scan(&userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, auth.ErrUnauthorized
		}

		return 0, err
	}

	return userID, nil
}

func (s *Storage) UpdateSession(ctx context.Context, refreshToken, newRefreshToken string) error {
	_, err := s.dbpool.Exec(
		ctx,
		"UPDATE sessions SET refresh_token = @new_refresh_token, expired = @expired WHERE refresh_token = @refresh_token",
		pgx.NamedArgs{
			"refresh_token":     refreshToken,
			"new_refresh_token": newRefreshToken,
			"expired":           time.Now().Add(_refreshTokenLifetime),
		},
	)

	return err
}
