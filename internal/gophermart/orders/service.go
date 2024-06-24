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
	UploadOrder(ctx context.Context, orderNumber string) error
}

type Repository interface {
}

type service struct {
	r Repository
}

func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) UploadOrder(ctx context.Context, orderNumber string) error {
	valid, err := luhn.IsValid(orderNumber)
	if err != nil {
		return err
	}

	if !valid {
		return ErrInvalidOrderFormat
	}

	return nil
}
