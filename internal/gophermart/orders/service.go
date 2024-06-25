package orders

import (
	"context"
	"errors"

	"github.com/EClaesson/go-luhn"
)

var (
	ErrInvalidOrderFormat         = errors.New("invalid order format")
	ErrAlreadyUploadedBySameUser  = errors.New("already uploaded by same user")
	ErrAlreadyUploadedByOtherUser = errors.New("already uploaded by other user")
	ErrOrderNotFound              = errors.New("order not found")
)

type Service interface {
	UploadOrder(ctx context.Context, orderNumber string, userID int64) error
	GetOrders(ctx context.Context, userID int64) ([]Order, error)
}

type Repository interface {
	UploadOrder(ctx context.Context, orderNumber string, userID int64) error
	GetOrders(ctx context.Context, userID int64) ([]Order, error)
}

type service struct {
	r Repository
}

func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) UploadOrder(ctx context.Context, orderNumber string, userID int64) error {

	if err := s.validateOrderNumber(orderNumber); err != nil {
		return err
	}

	return s.r.UploadOrder(ctx, orderNumber, userID)
}

func (s *service) GetOrders(ctx context.Context, userID int64) ([]Order, error) {
	return s.r.GetOrders(ctx, userID)
}

func (s *service) validateOrderNumber(number string) error {
	valid, err := luhn.IsValid(number)
	if err != nil || !valid {
		return ErrInvalidOrderFormat
	}

	return nil
}
