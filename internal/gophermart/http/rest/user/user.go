package user

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/gophermart/internal/gophermart"
	"github.com/kaa-it/gophermart/internal/gophermart/service/user"
	"net/http"
)

type Logger interface {
	RequestLogger(h http.HandlerFunc) http.HandlerFunc
	Error(args ...interface{})
}

type Handler struct {
	u user.Service
	l Logger
}

type CreateRequest struct {
	login    string
	password string
}

func NewHandler(u user.Service, l Logger) *Handler {
	return &Handler{u, l}
}

func (h *Handler) Route() *chi.Mux {
	mux := chi.NewMux()

	mux.Post("/register", h.l.RequestLogger(h.register))
	mux.Post("/login", h.l.RequestLogger(h.login))

	return mux
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := dec.Decode(&req); err != nil {
		h.l.Error(fmt.Sprintf("failed decoding body: %v", err))
		gophermart.DisplayAppError(w, http.StatusBadRequest, "failed decoding body")
		return
	}

	h.u.СreateUser(r.Context(), req.login, req.password)

}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {

}
