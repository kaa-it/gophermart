package user

import "context"

type Service interface {
}

type Repository interface {
}

type service struct {
	r Repository
}

func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) CreateUser(ctx context.Context, login string, password string) error {

}
