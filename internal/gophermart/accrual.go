package gophermart

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/go-resty/resty/v2"
	"github.com/kaa-it/gophermart/internal/gophermart/storage/postgres"
	"github.com/kaa-it/gophermart/pkg/logger"
)

const (
	_requestTimeout         = 5 * time.Second
	_retryCount             = 3
	_retryWaitTime          = time.Second
	_retryMaxWaitTime       = 5 * time.Second
	_retryDelay             = 2 * time.Second
	_idleDelay              = time.Second
	_pageSize               = 10
	_accrualStatusProcessed = "PROCESSED"
)

type AccrualResponse struct {
	Order   string          `json:"order"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

type Accrual struct {
	log     *logger.Logger
	storage *postgres.Storage
	client  *resty.Client
	address string
	cancel  context.CancelFunc
}

func NewAccrual(log *logger.Logger, storage *postgres.Storage, address string) *Accrual {
	client := resty.New()

	client.SetTimeout(_requestTimeout)
	client.SetRetryCount(_retryCount)
	client.SetRetryWaitTime(_retryWaitTime)
	client.SetRetryMaxWaitTime(_retryMaxWaitTime)
	client.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
		return _retryDelay, nil
	})

	return &Accrual{
		log,
		storage,
		client,
		address,
		nil,
	}
}

func (a *Accrual) Run() {

	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	var offset int64

	for {
		orders, err := a.storage.GetOrdersPage(ctx, _pageSize, offset)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			a.log.Error(fmt.Sprintf("Error getting orders: %v", err))
			continue
		}

		orderCount := len(orders)

		if orderCount == 0 {
			<-time.After(_idleDelay)
			continue
		}

		for _, order := range orders {
			resp, err := a.SendRequest(order.Number)
			if err != nil {
				a.log.Error(fmt.Sprintf("Error sending order to accrual: %v", err))
				continue
			}

			if resp.Status != _accrualStatusProcessed {
				err = a.storage.UpdateOrderStatus(ctx, order.Number, order.Status)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						break
					}
					a.log.Error(fmt.Sprintf("Error updating order status: %v", err))
					continue
				}
			} else {
				err = a.storage.UpdateOrderAccrual(ctx, order.Number, order.UserID, resp.Accrual)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						break
					}
					a.log.Error(fmt.Sprintf("Error updating order accrual: %v", err))
					continue
				}
			}
		}

		if orderCount == _pageSize {
			offset += int64(orderCount)
		} else {
			offset = 0
		}
	}
}

func (a *Accrual) Shutdown() {
	if a.cancel != nil {
		a.cancel()
	}
}

func (a *Accrual) SendRequest(orderNumber string) (*AccrualResponse, error) {
	req := a.client.R()
	req.Method = http.MethodGet

	url := fmt.Sprintf("%s/api/orders/%s", a.address, orderNumber)

	response := &AccrualResponse{}

	req.SetResult(response)

	_, err := req.Get(url)
	if err != nil {
		return nil, err
	}

	return response, nil
}
