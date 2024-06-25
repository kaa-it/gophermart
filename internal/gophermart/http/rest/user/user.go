package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kaa-it/gophermart/internal/gophermart/auth"
	"github.com/kaa-it/gophermart/internal/gophermart/http/rest"
	authUtils "github.com/kaa-it/gophermart/pkg/auth"
	"net/http"
)

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

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
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
	w.Header().Set("Authorization", credentials.AccessToken)
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
	w.Header().Set("Authorization", credentials.AccessToken)
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
	w.Header().Set("Authorization", credentials.AccessToken)
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(credentials); err != nil {
		h.l.Error(fmt.Sprintf("failed encoding credentials: %v", err))
		return
	}
}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	userID := authUtils.GetUserIDByToken(r)

	if userID == nil {
		h.l.Error("failed to get user id from token")
		rest.DisplayAppError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := r.Context()

	user, err := h.a.GetUserByID(ctx, *userID)
	if err != nil {
		h.l.Error(fmt.Sprintf("failed get user: %v", err))
		rest.DisplayAppError(w, http.StatusInternalServerError, "failed get user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	user.Current.Float64()

	balance := BalanceResponse{Current: user.Current.InexactFloat64(), Withdrawn: user.Withdrawn.InexactFloat64()}

	enc := json.NewEncoder(w)
	if err := enc.Encode(balance); err != nil {
		h.l.Error(fmt.Sprintf("failed encoding credentials: %v", err))
		return
	}
}
