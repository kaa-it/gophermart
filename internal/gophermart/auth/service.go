package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/kaa-it/gophermart/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserValidation = errors.New("validation error")
	ErrInvalidUser    = errors.New("invalid user")
	ErrUserNotFound   = errors.New("user not found")
	ErrUnauthorized   = errors.New("unauthorized")
)

type Service interface {
	CreateUser(ctx context.Context, user User) (*Credentials, error)
	Login(ctx context.Context, user User) (*Credentials, error)
	Token(ctx context.Context, refreshToken string) (*Credentials, error)
}

type Repository interface {
	CreateUser(ctx context.Context, user User, refreshToken string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*User, error)
	CreateSession(ctx context.Context, userID int64, refreshToken string) error
	RemoveExpiredSessions(ctx context.Context) error
	GetUserIDBySessionToken(ctx context.Context, refreshToken string) (int64, error)
	UpdateSession(ctx context.Context, refreshToken, newRefreshToken string) error
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

	userID, err := s.r.CreateUser(ctx, user, refreshToken)
	if err != nil {
		return nil, err
	}

	accessToken := auth.CreateAccessToken(userID)

	return &Credentials{accessToken, refreshToken}, nil
}

func (s *service) Login(ctx context.Context, user User) (*Credentials, error) {
	if err := s.validateUser(user); err != nil {
		return nil, ErrUserValidation
	}

	dbUser, err := s.r.GetUserByLogin(ctx, user.Login)
	if err != nil {
		return nil, err
	}

	fmt.Println("test1")

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		return nil, ErrUnauthorized
	}

	refreshToken := auth.CreateRefreshToken()

	fmt.Println("test2")

	err = s.r.CreateSession(ctx, dbUser.ID, refreshToken)
	if err != nil {
		return nil, err
	}

	fmt.Println("test3")

	accessToken := auth.CreateAccessToken(dbUser.ID)

	return &Credentials{accessToken, refreshToken}, nil
}

func (s *service) Token(ctx context.Context, refreshToken string) (*Credentials, error) {
	if err := s.r.RemoveExpiredSessions(ctx); err != nil {
		return nil, err
	}

	fmt.Println("test1")

	userID, err := s.r.GetUserIDBySessionToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	fmt.Println("test2")

	newRefreshToken := auth.CreateRefreshToken()

	err = s.r.UpdateSession(ctx, refreshToken, newRefreshToken)
	if err != nil {
		return nil, err
	}

	accessToken := auth.CreateAccessToken(userID)

	return &Credentials{accessToken, newRefreshToken}, nil
}

func (s *service) validateUser(user User) error {
	if user.Login == "" || user.Password == "" {
		return ErrUserValidation
	}

	return nil
}
