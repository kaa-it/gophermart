package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/kaa-it/gophermart/internal/gophermart/http/rest"
	"github.com/kaa-it/gophermart/internal/gophermart/orders"
	"github.com/kaa-it/gophermart/pkg/auth"
)

func (h *Handler) uploadOrder(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	userID := auth.GetUserIDByToken(r)

	if userID == nil {
		h.l.Error("failed to get user id from token")
		rest.DisplayAppError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := r.Context()

	if err := h.o.UploadOrder(ctx, string(b), *userID); err != nil {
		h.l.Error(fmt.Sprintf("failed to upload order: %v", err))

		if errors.Is(err, orders.ErrAlreadyUploadedBySameUser) {
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.Is(err, orders.ErrAlreadyUploadedByOtherUser) {
			rest.DisplayAppError(w, http.StatusConflict, "already uploaded by other user")
			return
		}

		if errors.Is(err, orders.ErrInvalidOrderFormat) {
			rest.DisplayAppError(w, http.StatusUnprocessableEntity, "invalid order format")
			return
		}

		rest.DisplayAppError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) getOrders(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDByToken(r)

	if userID == nil {
		h.l.Error("failed to get user id from token")
		rest.DisplayAppError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := r.Context()

	userOrders, err := h.o.GetOrders(ctx, *userID)
	if err != nil {
		h.l.Error(fmt.Sprintf("failed to get orders for user: %v", err))
		rest.DisplayAppError(w, http.StatusInternalServerError, "failed to get orders")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(userOrders) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(userOrders); err != nil {
		h.l.Error(fmt.Sprintf("failed encoding orders: %v", err))
		return
	}
}
