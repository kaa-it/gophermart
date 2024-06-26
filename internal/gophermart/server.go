package gophermart

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/gophermart/internal/gophermart/auth"
	"github.com/kaa-it/gophermart/internal/gophermart/http/rest/user"
	"github.com/kaa-it/gophermart/internal/gophermart/orders"
	"github.com/kaa-it/gophermart/internal/gophermart/storage/postgres"
	"github.com/kaa-it/gophermart/internal/gophermart/withdrawals"
	"github.com/kaa-it/gophermart/pkg/logger"
)

type Server struct {
	log     *logger.Logger
	storage *postgres.Storage
	server  *http.Server
	address string
}

func NewServer(log *logger.Logger, storage *postgres.Storage, address string) *Server {
	return &Server{
		log,
		storage,
		nil,
		address,
	}
}

func (s *Server) Run() {
	s.log.Info("starting server")

	userService := auth.NewService(s.storage)
	ordersService := orders.NewService(s.storage)
	withdrawalsService := withdrawals.NewService(s.storage)

	userHandler := user.NewHandler(userService, ordersService, withdrawalsService, s.log)

	r := chi.NewRouter()

	r.Mount("/api/user", userHandler.Route())

	s.server = &http.Server{
		Addr:    s.address,
		Handler: r,
	}

	err := s.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err.Error())
	}

	s.log.Info("server stopped")
}

func (s *Server) Shutdown() {
	if s.server == nil {
		return
	}

	if err := s.server.Shutdown(context.Background()); err != nil {
		s.log.Error(err.Error())
	}
}
