package auth

import (
	"github.com/shopspring/decimal"
)

type User struct {
	Id        int64           `json:"id,omitempty"`
	Login     string          `json:"login"`
	Password  string          `json:"password,omitempty"`
	Currency  decimal.Decimal `json:"currency"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}
