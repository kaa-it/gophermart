package auth

import (
	"context"
	"errors"
	"github.com/kaa-it/gophermart/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserValidation = errors.New("validation error")
var ErrInvalidUser = errors.New("invalid user")

type Service interface {
	CreateUser(ctx context.Context, user User) (*Credentials, error)
}

type Repository interface {
	CreateUser(ctx context.Context, user User, refreshToken string) (int64, error)
}

type service struct {
	r Repository
}

func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) CreateUser(ctx context.Context, user User) (*Credentials, error) {
	if err := s.validateUser(user); err != nil {
		return nil, ErrUserValidation
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user.Password = string(hash)

	refreshToken := auth.CreateRefreshToken()

	userId, err := s.r.CreateUser(ctx, user, refreshToken)
	if err != nil {
		return nil, err
	}

	accessToken := auth.CreateAccessToken(userId)

	return &Credentials{accessToken, refreshToken}, nil
}

//func (s *service) Login(ctx context.Context, login string, password string) (Credentials, error) {
//
//}

func (s *service) validateUser(user User) error {
	if user.Login == "" || user.Password == "" {
		return ErrUserValidation
	}

	return nil
}
