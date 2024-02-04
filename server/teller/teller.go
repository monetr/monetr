package teller

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/monetr/monetr/server/build"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/round"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const APIHostname = "api.teller.io"
const APIVersion = "2020-10-12"

type Client interface {
	GetHealth(ctx context.Context) error
	GetInstitutions(ctx context.Context) ([]Institution, error)
	GetAuthenticatedClient(accessToken string) AuthenticatedClient
}

type clientBase struct {
	log           *logrus.Entry
	client        *http.Client
	configuration config.Teller
}

func NewClient(log *logrus.Entry, configuration config.Teller) (Client, error) {
	base := &clientBase{
		log:           log,
		client:        nil,
		configuration: configuration,
	}

	var roundTripper = http.DefaultTransport
	if configuration.Certificate != "" && configuration.PrivateKey != "" {
		baseTransport := &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
		}
		cert, err := tls.LoadX509KeyPair(
			configuration.Certificate,
			configuration.PrivateKey,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load teller certificate and private key")
		}

		baseTransport.TLSClientConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		roundTripper = baseTransport
	}

	base.client = &http.Client{
		Transport: round.NewObservabilityRoundTripper(
			roundTripper,
			base.tellerRoundTripper,
		),
		Jar:     nil,
		Timeout: 1 * time.Second,
	}

	return base, nil
}

func (s *clientBase) tellerRoundTripper(
	ctx context.Context,
	request *http.Request,
	response *http.Response,
	err error,
) {
	var statusCode int
	var requestId string
	if response != nil {
		statusCode = response.StatusCode
		requestId = response.Header.Get("x-request-id")
	}
	// If you get a nil reference panic here during testing, its probably because you forgot to mock a certain endpoint.
	// Check to see if the error is a "no responder found" error.
	crumbs.HTTP(ctx,
		"Teller API Call",
		"teller",
		request.URL.String(),
		request.Method,
		statusCode,
		map[string]interface{}{
			"Request-Id": requestId,
		},
	)
}

func (c *clientBase) newUnauthenticatedRequest(ctx context.Context, method string, path string, body any) *http.Request {
	var reader io.Reader

	if body != nil {
		buffer := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buffer).Encode(body); err != nil {
			panic(fmt.Sprintf("failed to marshal request body: %+v", err))
		}
		reader = buffer
	}

	request, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("https://%s%s", APIHostname, path),
		reader,
	)
	if err != nil {
		panic(fmt.Sprintf("unable to create teller http request for %s %s", method, path))
	}

	request.Header.Add("Teller-Version", APIVersion)
	request.Header.Add("User-Agent", fmt.Sprintf("monetr-%s", build.Release))

	return request
}

func (c *clientBase) GetHealth(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	request := c.newUnauthenticatedRequest(span.Context(), "GET", "/institutions", nil)
	response, err := c.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "failed to request teller health")
	}

	if response.StatusCode == http.StatusOK {
		return nil
	}

	return errors.Errorf("failed teller health check: %d", response.StatusCode)
}

func (c *clientBase) GetInstitutions(ctx context.Context) ([]Institution, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	institutions := make([]Institution, 0)
	request := c.newUnauthenticatedRequest(span.Context(), "GET", "/institutions", nil)
	response, err := c.client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request teller institutions")
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&institutions); err != nil {
		return nil, errors.Wrap(err, "failed to decode json response")
	}

	return institutions, nil
}

func (c *clientBase) GetAuthenticatedClient(accessToken string) AuthenticatedClient {
	return &authenticatedClientBase{
		clientBase:  c,
		accessToken: accessToken,
	}
}

type AuthenticatedClient interface {
	GetAccounts(ctx context.Context) ([]Account, error)
	GetAccountBalance(ctx context.Context, id string) (*Balance, error)
	DeleteAccount(ctx context.Context, id string) error
	GetTransactions(ctx context.Context, accountId string, fromId *string, limit int64) ([]Transaction, error)
}

type authenticatedClientBase struct {
	*clientBase
	accessToken string
}

func (c *authenticatedClientBase) newAuthenticatedRequest(ctx context.Context, method string, path string, body any) *http.Request {
	request := c.newUnauthenticatedRequest(ctx, method, path, body)

	authentication := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:", c.accessToken)),
	)
	request.Header.Add("Authorization", fmt.Sprintf("Basic %s", authentication))

	return request
}

func (c *authenticatedClientBase) GetAccounts(ctx context.Context) ([]Account, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	items := make([]Account, 0)
	request := c.newAuthenticatedRequest(span.Context(), "GET", "/accounts", nil)
	response, err := c.client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request teller accounts")
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, errors.Errorf("failed to retrieve accounts: %d", response.StatusCode)
	}

	if err := json.NewDecoder(response.Body).Decode(&items); err != nil {
		return nil, errors.Wrap(err, "failed to decode json response")
	}

	return items, nil
}

func (c *authenticatedClientBase) DeleteAccount(ctx context.Context, id string) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	path := fmt.Sprintf("/accounts/%s", id)
	request := c.newAuthenticatedRequest(span.Context(), "DELETE", path, nil)
	response, err := c.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "failed to request delete teller account")
	}

	if response.StatusCode >= 400 {
		return errors.Errorf("failed to delete account: %d", response.StatusCode)
	}

	return nil
}

func (c *authenticatedClientBase) GetAccountBalance(
	ctx context.Context,
	id string,
) (*Balance, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result Balance
	path := fmt.Sprintf("/accounts/%s/balances", id)
	request := c.newAuthenticatedRequest(span.Context(), "GET", path, nil)
	response, err := c.client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request teller account balance")
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, errors.Errorf(
			"failed to retrieve Teller account balance: %d",
			response.StatusCode,
		)
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "failed to decode json response")
	}

	return &result, nil
}

func (c *authenticatedClientBase) GetTransactions(ctx context.Context, accountId string, fromId *string, limit int64) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	params := url.Values{
		"count": []string{
			strconv.FormatInt(limit, 10),
		},
	}
	if fromId != nil {
		params["from_id"] = []string{
			*fromId,
		}
	}
	path := fmt.Sprintf("/accounts/%s/transactions?%s", accountId, params.Encode())

	items := make([]Transaction, 0, limit)
	request := c.newAuthenticatedRequest(span.Context(), "GET", path, nil)
	response, err := c.client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request Teller transactions")
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, errors.Errorf(
			"failed to retrieve Teller transactions: %d",
			response.StatusCode,
		)
	}

	if err := json.NewDecoder(response.Body).Decode(&items); err != nil {
		return nil, errors.Wrap(err, "failed to decode json response")
	}

	return items, nil
}
