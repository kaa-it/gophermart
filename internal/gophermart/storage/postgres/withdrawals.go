package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kaa-it/gophermart/internal/gophermart/withdrawals"
	"github.com/shopspring/decimal"
)

type withdrawal struct {
	number      string
	userID      int64
	sum         decimal.NullDecimal
	processedAt time.Time
}

func (s *Storage) Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error {
	transaction, err := s.dbpool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer transaction.Rollback(ctx)

	user, err := s.getUserByID(ctx, transaction, userID)
	if err != nil {
		return err
	}

	newCurrent := user.Current.Sub(sum)

	if newCurrent.Sign() < 0 {
		return withdrawals.ErrNotEnoughFunds
	}

	newWithdrawn := user.Withdrawn.Add(sum)

	_, err = transaction.Exec(
		ctx,
		"UPDATE users set current = @current, withdrawn = @withdrawn WHERE id = @id",
		pgx.NamedArgs{
			"current":   newCurrent,
			"withdrawn": newWithdrawn,
			"id":        userID,
		},
	)

	if err != nil {
		return err
	}

	_, err = transaction.Exec(
		ctx,
		"INSERT INTO withdrawals (number, user_id, sum, processed_at) VALUES (@number, @user_id, @sum, @processed_at)",
		pgx.NamedArgs{
			"number":       orderNumber,
			"user_id":      userID,
			"sum":          sum,
			"processed_at": time.Now(),
		},
	)

	if err != nil {
		return err
	}

	return transaction.Commit(ctx)
}

func (s *Storage) GetWithdrawals(ctx context.Context, userID int64) ([]withdrawals.Withdrawal, error) {
	rows, err := s.dbpool.Query(
		ctx,
		"SELECT * FROM withdrawals WHERE user_id = @user_id ORDER BY processed_at DESC",
		pgx.NamedArgs{
			"user_id": userID,
		},
	)
	if err != nil {
		return nil, err
	}

	var userWithdrawals []withdrawals.Withdrawal
	for rows.Next() {
		dbWithdrawal := withdrawal{}
		err := rows.Scan(&dbWithdrawal.number, &dbWithdrawal.userID, &dbWithdrawal.sum, &dbWithdrawal.processedAt)
		if err != nil {
			return nil, err
		}

		userWithdrawal := withdrawals.Withdrawal{
			Order:       dbWithdrawal.number,
			Sum:         dbWithdrawal.sum.Decimal,
			ProcessedAt: dbWithdrawal.processedAt,
		}

		userWithdrawals = append(userWithdrawals, userWithdrawal)
	}

	return userWithdrawals, nil
}
