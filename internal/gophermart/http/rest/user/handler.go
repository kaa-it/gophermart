package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/gophermart/internal/gophermart/auth"
	"github.com/kaa-it/gophermart/internal/gophermart/orders"
	"github.com/kaa-it/gophermart/internal/gophermart/withdrawals"
	authUtils "github.com/kaa-it/gophermart/pkg/auth"
)

type Logger interface {
	RequestLogger(h http.HandlerFunc) http.HandlerFunc
	Error(args ...any)
}

type Handler struct {
	a  auth.Service
	o  orders.Service
	wd withdrawals.Service
	l  Logger
}

func NewHandler(a auth.Service, o orders.Service, wd withdrawals.Service, l Logger) *Handler {
	return &Handler{a, o, wd, l}
}

func (h *Handler) Route() *chi.Mux {
	mux := chi.NewMux()

	mux.Post("/register", h.l.RequestLogger(h.register))
	mux.Post("/login", h.l.RequestLogger(h.login))
	mux.Post("/token", h.l.RequestLogger(h.token))

	mux.Post("/orders", h.l.RequestLogger(authUtils.GetHandlerWithJwt(h.uploadOrder)))
	mux.Get("/orders", h.l.RequestLogger(authUtils.GetHandlerWithJwt(h.getOrders)))

	mux.Get("/balance", h.l.RequestLogger(authUtils.GetHandlerWithJwt(h.getBalance)))

	mux.Post("/balance/withdraw", h.l.RequestLogger(authUtils.GetHandlerWithJwt(h.withdraw)))
	mux.Get("/balance/withdrawals", h.l.RequestLogger(authUtils.GetHandlerWithJwt(h.getWithdrawals)))

	return mux
}
