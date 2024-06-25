package gophermart

import (
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/go-resty/resty/v2"
	"github.com/kaa-it/gophermart/internal/gophermart/storage/postgres"
	"github.com/kaa-it/gophermart/pkg/logger"
)

const (
	_requestTimeout   = 5 * time.Second
	_retryCount       = 3
	_retryWaitTime    = time.Second
	_retryMaxWaitTime = 5 * time.Second
	_retryDelay       = 2 * time.Second
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
	}
}

func (a *Accrual) Run() {
	// TODO  Go through db
	// TODO: Make request per order
	// TODO: Change db according result of request

	//var offset int64
	//
	//orders, err := a.storage.GetOrdersPage(context.Background(), 10, offset)
	//if err != nil {
	//	// Skip
	//}
	//
	//for _, order := range orders {
	//	// Send request
	//}
}

func (a *Accrual) Shutdown() {

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
