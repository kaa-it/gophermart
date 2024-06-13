package gophermart

import (
	"encoding/json"
	"net/http"
)

type appError struct {
	Error string `json:"error"`
}

func DisplayAppError(w http.ResponseWriter, httpStatus int, message string) {
	err := appError{
		Error: message,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)

	_ = json.NewEncoder(w).Encode(err)
}
