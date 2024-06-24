package user

import (
	"github.com/kaa-it/gophermart/internal/gophermart/http/rest"
	"github.com/kaa-it/gophermart/pkg/auth"
	"io"
	"net/http"
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

	w.WriteHeader(http.StatusOK)

	w.Write(b)
}
