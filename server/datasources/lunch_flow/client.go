package lunch_flow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/round"
	"github.com/pkg/errors"
)

const DateFormat = "2006-01-02"

// maxResponseBodySize caps how much of an upstream response body we will read.
// This is defense against a hostile or compromised upstream streaming an
// unbounded response to exhaust memory. 10mb is generous for realistic account
// and transaction payloads while bounding worst case allocation.
const maxResponseBodySize = 10 * 1024 * 1024

type LunchFlowAccountId = json.Number

type Account struct {
	Id              LunchFlowAccountId `json:"id"`
	Name            string             `json:"name"`
	InstitutionName string             `json:"institution_name"`
	InstitutionLogo *string            `json:"institution_logo"`
	Provider        string             `json:"provider"`
	Currency        string             `json:"currency"`
	Status          string             `json:"status"`
}

type Balance struct {
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
}

type Transaction struct {
	Id          string             `json:"id"`
	AccountId   LunchFlowAccountId `json:"accountId"`
	Amount      json.Number        `json:"amount"`
	Currency    string             `json:"currency"`
	Date        string             `json:"date"`
	Merchant    string             `json:"merchant"`
	Description string             `json:"description"`
}

type LunchFlowClient interface {
	GetAccounts(ctx context.Context) ([]Account, error)
	GetBalance(ctx context.Context, accountId LunchFlowAccountId) (*Balance, error)
	GetTransactions(ctx context.Context, accountId LunchFlowAccountId) ([]Transaction, error)
}

type lunchFlowClient struct {
	accessToken string
	apiUrl      url.URL
	log         *slog.Logger
	httpClient  *http.Client
}

func NewLunchFlowClient(
	log *slog.Logger,
	apiUrl string,
	accessToken string,
	configuration config.LunchFlow,
) (LunchFlowClient, error) {
	if !configuration.Enabled {
		log.Error("lunch flow is not enabled on this server but the client is being instantiated!",
			"bug", true,
		)
		return nil, errors.New("Lunch Flow is not enabled on this server")
	}

	if !configuration.IsAllowedApiUrl(apiUrl) {
		log.Warn("rejected Lunch Flow API URL that is not in the configured allowlist, please update your configuration if this url is valid!",
			"apiUrl", apiUrl,
		)
		return nil, errors.New("Lunch Flow API URL is not in the configured allowlist")
	}

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
				requestLog := log.With(
					"lunch_flow_method", request.Method,
					"lunch_flow_url", request.URL.String(),
				)
				var statusCode int
				if response != nil {
					statusCode = response.StatusCode
					requestLog = requestLog.With("lunch_flow_statusCode", statusCode)
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
				requestLog.DebugContext(ctx, "Lunch Flow API call")
			}),
	}

	return &lunchFlowClient{
		accessToken: accessToken,
		apiUrl:      *parsedUrl,
		log:         log,
		httpClient:  httpClient,
	}, nil
}

func (l *lunchFlowClient) doRequest(ctx context.Context, relativePath string, result any) error {
	// This should copy the url object since we arent taking the pointer. This way
	// we can modify the URL object with our new path and proceed with the actual
	// request.
	url := l.apiUrl
	url.Path = path.Join(url.Path, relativePath)
	requestUrl := url.String()
	request, err := http.NewRequestWithContext(ctx, "GET", requestUrl, http.NoBody)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}
	request.Header.Add("x-api-key", l.accessToken)

	response, err := l.httpClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "failed to send HTTP request")
	}
	defer response.Body.Close()

	body := io.LimitReader(response.Body, maxResponseBodySize)
	if response.StatusCode != http.StatusOK {
		bodyStr, _ := io.ReadAll(body)
		return errors.Errorf("Lunch Flow request failed %s [%d]: %s", requestUrl, response.StatusCode, string(bodyStr))
	}

	if err := json.NewDecoder(body).Decode(result); err != nil {
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
		"/accounts",
		&result,
	); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve accounts from Lunch Flow")
	}

	return result.Accounts, nil
}

func (l *lunchFlowClient) GetBalance(ctx context.Context, accountId LunchFlowAccountId) (*Balance, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result struct {
		Balance Balance `json:"balance"`
	}
	if err := l.doRequest(
		span.Context(),
		fmt.Sprintf("/accounts/%s/balance", accountId),
		&result,
	); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve account balance from Lunch Flow")
	}

	return &result.Balance, nil
}

func (l *lunchFlowClient) GetTransactions(ctx context.Context, accountId LunchFlowAccountId) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result struct {
		Transactions []Transaction `json:"transactions"`
		Total        int64         `json:"total"`
	}
	if err := l.doRequest(
		span.Context(),
		fmt.Sprintf("/accounts/%s/transactions", accountId),
		&result,
	); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve account transactions from Lunch Flow")
	}

	return result.Transactions, nil
}

// ParseDate takes the date string from a Lunch Flow transaction and converts it
// to a timestamp in the user's timezone. It will return an empty time if there
// is an error parsing the date.
func ParseDate(input string, timezone *time.Location) (time.Time, error) {
	date, err := time.ParseInLocation(
		"2006-01-02",
		input,
		timezone,
	)
	return date, errors.WithStack(err)
}
