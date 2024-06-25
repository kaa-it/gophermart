package withdrawals

import (
	"time"

	"github.com/shopspring/decimal"
)

type Withdrawal struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}
