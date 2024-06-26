package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/kaa-it/gophermart/internal/gophermart/http/rest"
	"github.com/kaa-it/gophermart/internal/gophermart/withdrawals"
	"github.com/kaa-it/gophermart/pkg/auth"
	"github.com/shopspring/decimal"
)

type WithdrawRequest struct {
	Order string          `json:"order"`
	Sum   decimal.Decimal `json:"sum"`
}

func (h *Handler) withdraw(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDByToken(r)

	if userID == nil {
		h.l.Error("failed to get user id from token")
		rest.DisplayAppError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req WithdrawRequest

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := dec.Decode(&req); err != nil {
		h.l.Error(fmt.Sprintf("failed decoding body: %v", err))
		rest.DisplayAppError(w, http.StatusBadRequest, "failed decoding body")
		return
	}

	ctx := r.Context()

	if err := h.wd.Withdraw(ctx, *userID, req.Order, req.Sum); err != nil {
		h.l.Error(fmt.Sprintf("failed withdraw: %v", err))

		if errors.Is(err, withdrawals.ErrNotEnoughFunds) {
			rest.DisplayAppError(w, http.StatusPaymentRequired, "not enough funds")
			return
		}

		if errors.Is(err, withdrawals.ErrInvalidOrderFormat) {
			rest.DisplayAppError(w, http.StatusUnprocessableEntity, "invalid order format")
			return
		}

		rest.DisplayAppError(w, http.StatusInternalServerError, err.Error())
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDByToken(r)

	if userID == nil {
		h.l.Error("failed to get user id from token")
		rest.DisplayAppError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := r.Context()

	userWithdrawals, err := h.wd.GetWithdrawals(ctx, *userID)
	if err != nil {
		h.l.Error(fmt.Sprintf("failed to get withdrawals for user: %v", err))
		rest.DisplayAppError(w, http.StatusInternalServerError, "failed to get withdrawals")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(userWithdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(userWithdrawals); err != nil {
		h.l.Error(fmt.Sprintf("failed encoding orders: %v", err))
		return
	}
}
