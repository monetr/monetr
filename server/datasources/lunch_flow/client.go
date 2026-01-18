package lunch_flow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/round"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const DefaultBaseURL = "https://lunchflow.com/"

type AccountId json.Number

type Account struct {
	Id              AccountId `json:"id"`
	Name            string    `json:"name"`
	InstitutionName string    `json:"institution_name"`
	InstitutionLogo *string   `json:"institution_logo"`
	Provider        string    `json:"provider"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
}

type Balance struct {
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
}

type Transaction struct {
	Id          string      `json:"id"`
	AccountId   AccountId   `json:"accountId"`
	Amount      json.Number `json:"amount"`
	Currency    string      `json:"currency"`
	Date        string      `json:"date"`
	Merchant    string      `json:"merchant"`
	Description string      `json:"description"`
}

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=client.go -package=mockgen -destination=../../internal/mockgen/lunch_flow_client.go LunchFlowClient
type LunchFlowClient interface {
	GetAccounts(ctx context.Context) ([]Account, error)
	GetBalance(ctx context.Context, accountId AccountId) (*Balance, error)
	GetTransactions(ctx context.Context, accountId AccountId) ([]Transaction, error)
}

type lunchFlowClient struct {
	accessToken string
	apiUrl      url.URL
	log         *logrus.Entry
	httpClient  *http.Client
}

func NewLunchFlowClient(
	log *logrus.Entry,
	apiUrl string,
	accessToken string,
) (LunchFlowClient, error) {
	parsedUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: round.NewObservabilityRoundTripper(http.DefaultTransport,
			func(
				ctx context.Context,
				request *http.Request,
				response *http.Response,
				err error,
			) {
				requestLog := log.WithContext(ctx).WithFields(logrus.Fields{
					"lunch_flow_method": request.Method,
					"lunch_flow_url":    request.URL.String(),
				})
				var statusCode int
				if response != nil {
					statusCode = response.StatusCode
					requestLog = requestLog.WithField("lunch_flow_statusCode", statusCode)
				}

				// If you get a nil reference panic here during testing, its probably
				// because you forgot to mock a certain endpoint. Check to see if the
				// error is a "no responder found" error.
				crumbs.HTTP(ctx,
					"Lunch Flow API Call",
					"lunch_flow",
					request.URL.String(),
					request.Method,
					statusCode,
					map[string]any{},
				)
				requestLog.Debug("Lunch Flow API call")
			}),
	}

	return &lunchFlowClient{
		accessToken: accessToken,
		apiUrl:      *parsedUrl,
		log:         log,
		httpClient:  httpClient,
	}, nil
}

func (l *lunchFlowClient) doRequest(ctx context.Context, path string, result any) error {
	url := l.apiUrl
	url.Path = path
	requestUrl := url.String()
	request, err := http.NewRequestWithContext(ctx, "GET", requestUrl, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", l.accessToken))

	response, err := l.httpClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "failed to send HTTP request")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.Errorf("lunch flow request failed %s [%d]", requestUrl, response.StatusCode)
	}

	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return errors.Wrapf(err, "failed to decode response for request %s [%d]", requestUrl, response.StatusCode)
	}

	return nil
}

func (l *lunchFlowClient) GetAccounts(ctx context.Context) ([]Account, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result struct {
		Accounts []Account `json:"accounts"`
		Total    int64     `json:"total"`
	}
	if err := l.doRequest(
		span.Context(),
		"/api/v1/accounts",
		&result,
	); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve accounts from lunch flow")
	}

	return result.Accounts, nil
}

func (l *lunchFlowClient) GetBalance(ctx context.Context, accountId AccountId) (*Balance, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result struct {
		Balance Balance `json:"balance"`
	}
	if err := l.doRequest(
		span.Context(),
		fmt.Sprintf("/api/v1/accounts/%s/balance", accountId),
		&result,
	); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve account balance from lunch flow")
	}

	return &result.Balance, nil
}

func (l *lunchFlowClient) GetTransactions(ctx context.Context, accountId AccountId) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result struct {
		Transactions []Transaction `json:"transactions"`
		Total        int64         `json:"total"`
	}
	if err := l.doRequest(
		span.Context(),
		fmt.Sprintf("/api/v1/accounts/%s/transactions", accountId),
		&result,
	); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve account transactions from lunch flow")
	}

	return result.Transactions, nil
}
