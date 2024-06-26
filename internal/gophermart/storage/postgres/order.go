package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kaa-it/gophermart/internal/gophermart/orders"
	"github.com/shopspring/decimal"
)

type order struct {
	number     string
	userID     int64
	status     string
	accrual    decimal.NullDecimal
	uploadedAt time.Time
}

func (s *Storage) UploadOrder(ctx context.Context, orderNumber string, userID int64) error {
	_, err := s.dbpool.Exec(
		ctx,
		"INSERT INTO orders (number, user_id, status, uploaded_at) VALUES (@number, @user_id, @status, @uploaded_at)",
		pgx.NamedArgs{
			"number":      orderNumber,
			"user_id":     userID,
			"status":      orders.OrderStatusNew,
			"uploaded_at": time.Now(),
		},
	)

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !(errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation) {
		return err
	}

	order, err := s.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return err
	}

	if order.UserID == userID {
		return orders.ErrAlreadyUploadedBySameUser
	}

	return orders.ErrAlreadyUploadedByOtherUser
}

func (s *Storage) GetOrderByNumber(ctx context.Context, orderNumber string) (*orders.Order, error) {
	var res order

	err := s.dbpool.QueryRow(
		ctx,
		"SELECT * FROM orders WHERE number = @number",
		pgx.NamedArgs{
			"number": orderNumber,
		},
	).Scan(&res.number, &res.userID, &res.status, &res.accrual, &res.uploadedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orders.ErrOrderNotFound
		}

		return nil, err
	}

	o := &orders.Order{
		Number:     res.number,
		UserID:     res.userID,
		Status:     res.status,
		UploadedAt: res.uploadedAt,
	}

	if res.accrual.Valid {
		o.Accrual = &res.accrual.Decimal
	}

	return o, nil
}

func (s *Storage) GetOrders(ctx context.Context, userID int64) ([]orders.Order, error) {
	rows, err := s.dbpool.Query(
		ctx,
		"SELECT * FROM orders WHERE user_id = @user_id ORDER BY uploaded_at DESC",
		pgx.NamedArgs{
			"user_id": userID,
		},
	)
	if err != nil {
		return nil, err
	}

	var userOrders []orders.Order
	for rows.Next() {
		dbOrder := order{}
		err := rows.Scan(&dbOrder.number, &dbOrder.userID, &dbOrder.status, &dbOrder.accrual, &dbOrder.uploadedAt)
		if err != nil {
			return nil, err
		}

		userOrder := orders.Order{
			Number:     dbOrder.number,
			UserID:     dbOrder.userID,
			Status:     dbOrder.status,
			UploadedAt: dbOrder.uploadedAt,
		}

		if dbOrder.accrual.Valid {
			userOrder.Accrual = &dbOrder.accrual.Decimal
		}

		userOrders = append(userOrders, userOrder)
	}

	return userOrders, nil
}

func (s *Storage) UpdateOrderStatus(ctx context.Context, orderNumber string, status string) error {
	_, err := s.dbpool.Exec(
		ctx,
		"UPDATE orders SET status = @status WHERE number = @number",
		pgx.NamedArgs{
			"number": orderNumber,
			"status": status,
		},
	)

	return err
}

func (s *Storage) UpdateOrderAccrual(ctx context.Context, orderNumber string, userID int64, accrual decimal.Decimal) error {
	transaction, err := s.dbpool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer transaction.Rollback(ctx)

	_, err = transaction.Exec(
		ctx,
		"UPDATE orders SET accrual = @accrual, status = @status WHERE number = @number",
		pgx.NamedArgs{
			"number":  orderNumber,
			"accrual": accrual,
			"status":  "PROCESSED",
		},
	)

	if err != nil {
		return err
	}

	_, err = transaction.Exec(
		ctx,
		"UPDATE users SET current = users.current + @accrual WHERE id = @user_id",
		pgx.NamedArgs{
			"user_id": userID,
			"accrual": accrual,
		},
	)

	if err != nil {
		return err
	}

	return transaction.Commit(ctx)
}

func (s *Storage) GetOrdersPage(ctx context.Context, limit int64, offset int64) ([]orders.Order, error) {
	rows, err := s.dbpool.Query(
		ctx,
		"SELECT * FROM orders WHERE (status != 'INVALID') AND (status != 'PROCESSED') ORDER BY uploaded_at LIMIT @max OFFSET @begin",
		pgx.NamedArgs{
			"max":   limit,
			"begin": offset,
		},
	)
	if err != nil {
		return nil, err
	}

	var userOrders []orders.Order
	for rows.Next() {
		dbOrder := order{}
		err := rows.Scan(&dbOrder.number, &dbOrder.userID, &dbOrder.status, &dbOrder.accrual, &dbOrder.uploadedAt)
		if err != nil {
			return nil, err
		}

		userOrder := orders.Order{
			Number:     dbOrder.number,
			UserID:     dbOrder.userID,
			Status:     dbOrder.status,
			UploadedAt: dbOrder.uploadedAt,
		}

		if dbOrder.accrual.Valid {
			userOrder.Accrual = &dbOrder.accrual.Decimal
		}

		userOrders = append(userOrders, userOrder)
	}

	return userOrders, nil
}
