package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	_ MonetrClient = &monetrHttpClient{}
)

type monetrHttpClient struct {
	log      *logrus.Entry
	endpoint string
	token    string
	client   *http.Client
}

func NewMonetrHTTPClient(log *logrus.Entry, endpoint, token string) MonetrClient {
	return &monetrHttpClient{
		log:      log,
		endpoint: endpoint,
		token:    token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (m *monetrHttpClient) GetTransactions(ctx context.Context, bankAccountId uint64, count, offset int64) ([]models.Transaction, error) {
	result := make([]models.Transaction, 0)
	if err := m.request(ctx, fmt.Sprintf("/api/bank_accounts/%d/transactions", bankAccountId), url.Values{
		"count": []string{
			strconv.FormatInt(count, 10),
		},
		"offset": []string{
			strconv.FormatInt(offset, 10),
		},
	}, &result); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve transactions")
	}

	return result, nil
}

func (m *monetrHttpClient) GetSpending(ctx context.Context, bankAccountId uint64) ([]models.Spending, error) {
	result := make([]models.Spending, 0)
	if err := m.request(ctx, fmt.Sprintf("/api/bank_accounts/%d/spending", bankAccountId), nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve spending")
	}

	return result, nil
}

func (m *monetrHttpClient) GetFundingSchedules(ctx context.Context, bankAccountId uint64) ([]models.FundingSchedule, error) {
	result := make([]models.FundingSchedule, 0)
	if err := m.request(ctx, fmt.Sprintf("/api/bank_accounts/%d/funding_schedules", bankAccountId), nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve funding schedules")
	}

	return result, nil
}

func (m *monetrHttpClient) GetBankAccounts(ctx context.Context) ([]models.BankAccount, error) {
	result := make([]models.BankAccount, 0)
	if err := m.request(ctx, "/api/bank_accounts", nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve bank accounts")
	}

	return result, nil
}

func (m *monetrHttpClient) GetLinks(ctx context.Context) ([]models.Link, error) {
	result := make([]models.Link, 0)
	if err := m.request(ctx, "/api/links", nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve links")
	}

	return result, nil
}

func (m *monetrHttpClient) GetMe(ctx context.Context) (*models.User, error) {
	var result struct {
		User *models.User `json:"user"`
	}
	if err := m.request(ctx, "/api/users/me", nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve user details")
	}

	return result.User, nil
}

func (m *monetrHttpClient) request(ctx context.Context, path string, query url.Values, result interface{}) error {
	uri, err := url.Parse(m.endpoint)
	if err != nil {
		return errors.Wrap(err, "failed to parse monetr endpoint")
	}
	uri.Path = path
	if query != nil {
		uri.RawQuery = query.Encode()
	}

	log := m.log.WithFields(logrus.Fields{
		"server": uri.Hostname(),
		"path":   path,
		"query":  query,
	})

	request, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to create monetr request")
	}
	request.AddCookie(&http.Cookie{
		Name:     "M-Token",
		Value:    m.token,
		Domain:   uri.Hostname(),
		Expires:  time.Now().Add(30 * time.Second),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	start := time.Now()
	response, err := m.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "failed to make request to monetr")
	}
	end := time.Since(start)
	defer response.Body.Close()

	log.WithField("rtt", end).Trace("retrieved data from monetr")

	if response.StatusCode != 200 {
		var responseError struct {
			Error string `json:"error"`
		}
		if err = json.NewDecoder(response.Body).Decode(&responseError); err != nil {
			return errors.Wrap(err, "failed to decode error response body")
		}

		return errors.Errorf("request failure [%d]: %s", response.StatusCode, responseError.Error)
	}

	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		return errors.Wrap(err, "failed to decode response body")
	}

	return nil
}
