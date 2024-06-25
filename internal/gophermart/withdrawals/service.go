package withdrawals

import (
	"context"
	"errors"

	"github.com/EClaesson/go-luhn"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidOrderFormat = errors.New("invalid order format")
	ErrNotEnoughFunds     = errors.New("not enough funds")
)

type Service interface {
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error
	GetWithdrawals(ctx context.Context, userID int64) ([]Withdrawal, error)
}

type Repository interface {
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error
	GetWithdrawals(ctx context.Context, userID int64) ([]Withdrawal, error)
}

type service struct {
	r Repository
}

func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error {
	if err := s.validateOrderNumber(orderNumber); err != nil {
		return err
	}

	return s.r.Withdraw(ctx, userID, orderNumber, sum)
}

func (s *service) GetWithdrawals(ctx context.Context, userID int64) ([]Withdrawal, error) {
	return s.r.GetWithdrawals(ctx, userID)
}

func (s *service) validateOrderNumber(number string) error {
	valid, err := luhn.IsValid(number)
	if err != nil || !valid {
		return ErrInvalidOrderFormat
	}

	return nil
}
