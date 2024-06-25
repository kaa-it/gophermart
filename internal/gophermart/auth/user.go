package auth

import (
	"github.com/shopspring/decimal"
)

type User struct {
	ID        int64           `json:"id,omitempty"`
	Login     string          `json:"login"`
	Password  string          `json:"password,omitempty"`
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}
