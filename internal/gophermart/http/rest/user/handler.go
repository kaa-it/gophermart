package user

import (
	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/gophermart/internal/gophermart/auth"
	authUtils "github.com/kaa-it/gophermart/pkg/auth"
	"net/http"
)

type Logger interface {
	RequestLogger(h http.HandlerFunc) http.HandlerFunc
	Error(args ...any)
}

type Handler struct {
	a auth.Service
	l Logger
}

func NewHandler(a auth.Service, l Logger) *Handler {
	return &Handler{a, l}
}

func (h *Handler) Route() *chi.Mux {
	mux := chi.NewMux()

	mux.Post("/register", h.l.RequestLogger(h.register))
	mux.Post("/login", h.l.RequestLogger(h.login))
	mux.Post("/token", h.l.RequestLogger(h.token))

	mux.Post("/orders", h.l.RequestLogger(authUtils.GetHandlerWithJwt(h.uploadOrder)))

	return mux
}
