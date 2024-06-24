package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/gophermart/internal/gophermart/auth"
	"github.com/kaa-it/gophermart/internal/gophermart/http/rest"
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

type CreateRequest struct {
	Login    string
	Password string
}

type LoginRequest struct {
	Login    string
	Password string
}

type TokenRequest struct {
	RefreshToken string
}

func NewHandler(a auth.Service, l Logger) *Handler {
	return &Handler{a, l}
}

func (h *Handler) Route() *chi.Mux {
	mux := chi.NewMux()

	mux.Post("/register", h.l.RequestLogger(h.register))
	mux.Post("/login", h.l.RequestLogger(h.login))
	mux.Post("/token", h.l.RequestLogger(h.token))

	return mux
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := dec.Decode(&req); err != nil {
		h.l.Error(fmt.Sprintf("failed decoding body: %v", err))
		rest.DisplayAppError(w, http.StatusBadRequest, "failed decoding body")
		return
	}

	user := auth.User{Login: req.Login, Password: req.Password}

	credentials, err := h.a.CreateUser(r.Context(), user)
	if err != nil {
		h.l.Error(fmt.Sprintf("failed create user: %v", err))
		if errors.Is(err, auth.ErrUserValidation) {
			rest.DisplayAppError(w, http.StatusBadRequest, "failed create user")
			return
		}

		if errors.Is(err, auth.ErrInvalidUser) {
			rest.DisplayAppError(w, http.StatusConflict, "failed create user")
			return
		}

		rest.DisplayAppError(w, http.StatusInternalServerError, "failed create user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(credentials); err != nil {
		h.l.Error(fmt.Sprintf("failed encoding credentials: %v", err))
		return
	}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := dec.Decode(&req); err != nil {
		h.l.Error(fmt.Sprintf("failed decoding body: %v", err))
		rest.DisplayAppError(w, http.StatusBadRequest, "failed decoding body")
		return
	}

	user := auth.User{Login: req.Login, Password: req.Password}

	credentials, err := h.a.Login(r.Context(), user)
	if err != nil {
		h.l.Error(fmt.Sprintf("failed login: %v", err))
		if errors.Is(err, auth.ErrUserValidation) {
			rest.DisplayAppError(w, http.StatusBadRequest, "failed login")
		}

		if errors.Is(err, auth.ErrUserNotFound) || errors.Is(err, auth.ErrUnauthorized) {
			rest.DisplayAppError(w, http.StatusUnauthorized, "failed login")
			return
		}

		rest.DisplayAppError(w, http.StatusInternalServerError, "failed login")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(credentials); err != nil {
		h.l.Error(fmt.Sprintf("failed encoding credentials: %v", err))
		return
	}
}

func (h *Handler) token(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := dec.Decode(&req); err != nil {
		h.l.Error(fmt.Sprintf("failed decoding body: %v", err))
		rest.DisplayAppError(w, http.StatusBadRequest, "failed decoding body")
		return
	}

	fmt.Printf("req: %v/n", req)

	credentials, err := h.a.Token(r.Context(), req.RefreshToken)
	if err != nil {
		h.l.Error(fmt.Sprintf("failed refresh tokens: %v", err))
		if errors.Is(err, auth.ErrUnauthorized) {
			rest.DisplayAppError(w, http.StatusUnauthorized, "failed refresh tokens")
			return
		}

		rest.DisplayAppError(w, http.StatusInternalServerError, "failed refresh tokens")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(credentials); err != nil {
		h.l.Error(fmt.Sprintf("failed encoding credentials: %v", err))
		return
	}
}
