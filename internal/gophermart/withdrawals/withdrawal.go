package withdrawals

import (
	"github.com/shopspring/decimal"
	"time"
)

type Withdrawal struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}
