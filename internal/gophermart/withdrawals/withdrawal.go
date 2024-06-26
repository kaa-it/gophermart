package withdrawals

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type Withdrawal struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}

func (w Withdrawal) MarshalJSON() ([]byte, error) {
	type WithdrawalAlias Withdrawal

	alias := struct {
		WithdrawalAlias
		Sum float64 `json:"sum"`
	}{
		WithdrawalAlias: WithdrawalAlias(w),
		Sum:             w.Sum.InexactFloat64(),
	}

	return json.Marshal(alias)
}
