package teller

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
}

type AuthenticatedClient interface {
	GetAccounts(ctx context.Context) ([]Account, error)
	DeleteAccount(ctx context.Context, id string) error
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

	baseTransport := &http.Transport{
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}
	if configuration.Certificate != "" && configuration.PrivateKey != "" {
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
	}

	base.client = &http.Client{
		Transport: round.NewObservabilityRoundTripper(
			baseTransport,
			base.tellerRoundTripper,
		),
		Jar:     nil,
		Timeout: 60 * time.Second,
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

func (c *clientBase) newRequest(ctx context.Context, method string, path string, body io.Reader) *http.Request {
	request, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("https://%s%s", APIHostname, path),
		body,
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

	return nil
}

func (c *clientBase) GetInstitutions(ctx context.Context) ([]Institution, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	institutions := make([]Institution, 0)
	request := c.newRequest(span.Context(), "GET", "/institutions", nil)
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
