package simplefin

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/round"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	Hostnames = []string{
		"bridge.simplefin.org",
		"beta-bridge.simplefin.org",
	}
)

type Client interface {
	GetAccounts(ctx context.Context) error
}

type Transaction struct {
	ID          string         `json:"id"`
	Posted      int64          `json:"posted"`
	Amount      json.Number    `json:"amount"`
	Description string         `json:"description"`
	Pending     bool           `json:"pending"`
	Extra       map[string]any `json:"extra"`
}

type Organization struct {
	Domain  string  `json:"domain"`
	SFinURL string  `json:"sfin-url"`
	Name    string  `json:"name"`
	URL     *string `json:"url"`
	ID      *string `json:"id"`
}

type Account struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Currency         string        `json:"currency"`
	Balance          json.Number   `json:"balance"`
	AvailableBalance json.Number   `json:"available-balance"`
	BalanceDate      int64         `json:"balance-date"`
	Transactions     []Transaction `json:"transactions"`
}

var (
	_ Client = &simplefinClient{}
)

type simplefinClient struct {
	log                *logrus.Entry
	clock              clock.Clock
	client             *http.Client
	username, password string
}

func NewSimpleFINClient(log *logrus.Entry, username, password string) Client {
	return &simplefinClient{
		log: log,
		client: &http.Client{
			Timeout: 60 * time.Second,
			Transport: round.NewObservabilityRoundTripper(http.DefaultTransport,
				func(
					ctx context.Context,
					request *http.Request,
					response *http.Response,
					err error,
				) {
					requestLog := log.WithContext(ctx).WithFields(logrus.Fields{
						"simplefin_method": request.Method,
						"simplefin_url":    request.URL.String(),
					})
					var statusCode int
					if response != nil {
						statusCode = response.StatusCode
						requestLog = requestLog.WithField("simplefin_statusCode", statusCode)

						var responseData map[string]any
						buffer := bytes.NewBuffer(nil)
						tee := io.TeeReader(response.Body, buffer)
						if err := json.NewDecoder(tee).Decode(&responseData); err != nil {
							log.WithError(err).Warn("failed to decode simplefin response as json for logging")
						}

						// Close the existing body before we replace it.
						response.Body.Close()
						// Then swap out the body
						response.Body = io.NopCloser(buffer)
					}

					// If you get a nil reference panic here during testing, its probably because you forgot to mock a certain endpoint.
					// Check to see if the error is a "no responder found" error.
					crumbs.HTTP(ctx,
						"SimpleFIN API Call",
						"simplefin",
						request.URL.String(),
						request.Method,
						statusCode,
						map[string]any{},
					)
					requestLog.Debug("SimpleFIN API call")
				}),
		},
		username: username,
		password: password,
	}
}

// GetAccounts implements Client.
func (s *simplefinClient) GetAccounts(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	url := url.URL{
		Scheme: "https",
		Host:   "bridge.simplefin.org",
		Path:   "/simplefine/accounts",
		User:   url.UserPassword(s.username, s.password),
	}
	url.Query().Add("start-date", s.clock.Now().Format("..."))
	url.Query().Add("end-date", s.clock.Now().Format("..."))
	url.RawQuery = url.Query().Encode()

	request, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to create SimpleFIN request")
	}

	response, err := s.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "failed to perform SimpleFIN request")
	}

	panic("unimplemented")
}

// ParseSimpleFINToken takes a base64 encoded token string from SimpleFIN. This
// string would be provided by an end user. This function parses it and verifies
// that it is a real token by decoding it and parsing it as a url. It then
// extracts the username and password from the URL after validating the domain
// name. An error is returned if the domain name is not a valid SimpleFIN
// domain.
func ParseSimpleFINToken(token string) (username, password string, err error) {
	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", "", errors.Wrap(err, "SimpleFIN token is not valid base64")
	}

	uri, err := url.ParseRequestURI(string(data))
	if err != nil {
		return "", "", errors.Wrap(err, "SimpleFIN token provided is not valid")
	}

	{ // Validate the hostname of the provided token
		validHostname := false
		providedHostname := uri.Hostname()
		for _, hostname := range Hostnames {
			if strings.EqualFold(providedHostname, hostname) {
				validHostname = true
				break
			}
		}
		if !validHostname {
			return "", "", errors.Errorf("SimpleFIN hostname is not valid: %s", providedHostname)
		}
	}

	if !strings.EqualFold(uri.Scheme, "https") {
		return "", "", errors.Errorf("SimpleFIN scheme is not valid: %s", uri.Scheme)
	}

	if !strings.EqualFold(uri.Scheme, "https") {
		return "", "", errors.Errorf("SimpleFIN scheme is not valid: %s", uri.Scheme)
	}

	password, ok := uri.User.Password()
	if !ok {
		return "", "", errors.New("SimpleFIN password is required")
	}

	username = uri.User.Username()
	if username == "" {
		return "", "", errors.New("SimpleFIN username is required")
	}

	return username, password, nil
}
