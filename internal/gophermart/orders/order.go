package orders

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	Number     string           `json:"number"`
	UserID     int64            `json:"-,"`
	Status     string           `json:"status"`
	Accrual    *decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time        `json:"uploaded_at"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	type OrderAlias Order

	var accrual *float64
	if o.Accrual != nil {
		accrualFloat := o.Accrual.InexactFloat64()
		accrual = &accrualFloat
	}

	alias := struct {
		OrderAlias
		Accrual *float64 `json:"accrual,omitempty"`
	}{
		OrderAlias: OrderAlias(o),
		Accrual:    accrual,
	}

	return json.Marshal(alias)
}
