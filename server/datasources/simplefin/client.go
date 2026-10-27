package simplefin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/round"
	"github.com/pkg/errors"
)

// maxResponseBodySize caps how much of an upstream response body we will read.
// Same as Lunch Flow, this is to keep a hostile upstream from streaming an
// unbounded response at us.
const maxResponseBodySize = 10 * 1024 * 1024

// maxClaimResponseSize caps the claim response body, which should only ever be
// the access URL.
const maxClaimResponseSize = 4 * 1024

var (
	ErrInvalidSetupToken = errors.New("invalid SimpleFIN setup token")
	ErrSetupTokenClaimed = errors.New("SimpleFIN setup token has already been claimed")
	ErrHostNotAllowed    = errors.New("SimpleFIN host is not in the configured allowlist")
	ErrPaymentRequired   = errors.New("SimpleFIN requires payment")
	ErrForbidden         = errors.New("SimpleFIN access was denied")
)

type AccountSet struct {
	Errlist []Error `json:"errlist"`
	// Errors is deprecated in the protocol in favor of errlist, but the Bridge
	// still puts things like quota warnings in here.
	Errors      []string     `json:"errors"`
	Connections []Connection `json:"connections"`
	Accounts    []Account    `json:"accounts"`
}

type Error struct {
	Code         string  `json:"code"`
	Message      string  `json:"msg"`
	ConnectionId *string `json:"conn_id"`
	AccountId    *string `json:"account_id"`
}

type Connection struct {
	ConnectionId string `json:"conn_id"`
	Name         string `json:"name"`
}

type Account struct {
	Id               string        `json:"id"`
	Name             string        `json:"name"`
	ConnectionId     string        `json:"conn_id"`
	Currency         string        `json:"currency"`
	Balance          json.Number   `json:"balance"`
	AvailableBalance *json.Number  `json:"available-balance"`
	BalanceDate      int64         `json:"balance-date"`
	Transactions     []Transaction `json:"transactions"`
}

type Transaction struct {
	Id           string      `json:"id"`
	Posted       int64       `json:"posted"`
	Amount       json.Number `json:"amount"`
	Description  string      `json:"description"`
	TransactedAt *int64      `json:"transacted_at"`
	Pending      bool        `json:"pending"`
}

type GetAccountsOptions struct {
	// StartDate will restrict the transactions returned to ones posted on or
	// after this date. If it is zero then it is not sent and SimpleFIN picks.
	StartDate time.Time
	// BalancesOnly will tell SimpleFIN to not return any transactions at all.
	BalancesOnly bool
}

type SimpleFinClient interface {
	GetAccounts(ctx context.Context, options GetAccountsOptions) (*AccountSet, error)
}

type simpleFinClient struct {
	log        *slog.Logger
	baseUrl    url.URL
	username   string
	password   string
	httpClient *http.Client
}

// ClaimSetupToken takes the setup token the user got from their SimpleFIN
// server and exchanges it for an access URL. The access URL has the basic auth
// credentials in it and should be treated as a secret. Setup tokens can only be
// claimed once, if the token has already been claimed then
// [ErrSetupTokenClaimed] is returned.
func ClaimSetupToken(
	ctx context.Context,
	log *slog.Logger,
	setupToken string,
	configuration config.SimpleFin,
) (string, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	if !configuration.Enabled {
		log.Error("simplefin is not enabled on this server but a setup token is being claimed!",
			"bug", true,
		)
		return "", errors.New("SimpleFIN is not enabled on this server")
	}

	// The setup token is just the claim URL base64 encoded. Trim the padding off
	// and decode without it, this way it works whether or not whoever gave the
	// user their token included the padding.
	decoded, err := base64.RawStdEncoding.DecodeString(
		strings.TrimRight(strings.TrimSpace(setupToken), "="),
	)
	if err != nil {
		return "", errors.WithStack(ErrInvalidSetupToken)
	}

	claimUrl, err := url.Parse(strings.TrimSpace(string(decoded)))
	if err != nil || claimUrl.Host == "" {
		return "", errors.WithStack(ErrInvalidSetupToken)
	}

	// Never log the full claim URL, the token is in the path and if the claim
	// fails for some other reason then it could still be claimed by whoever can
	// read our logs.
	if !configuration.IsAllowedUrl(claimUrl.String()) {
		log.WarnContext(span.Context(), "rejected SimpleFIN setup token for a host that is not in the configured allowlist, please update your configuration if this host is valid!",
			"host", claimUrl.Host,
		)
		return "", errors.WithStack(ErrHostNotAllowed)
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(request *http.Request, _ []*http.Request) error {
			return errors.Errorf("blocked SimpleFIN redirect to %s", request.URL.Host)
		},
	}

	request, err := http.NewRequestWithContext(
		span.Context(),
		http.MethodPost,
		claimUrl.String(),
		http.NoBody,
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	response, err := httpClient.Do(request)
	if err != nil {
		// The url.Error has the whole claim URL in it, so only keep what's inside
		// of it for the same reason as above.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return "", errors.Wrap(err, "failed to send SimpleFIN claim request")
	}
	defer response.Body.Close()

	log.DebugContext(span.Context(), "SimpleFIN setup token claim",
		"host", claimUrl.Host,
		"statusCode", response.StatusCode,
	)

	body, err := io.ReadAll(io.LimitReader(response.Body, maxClaimResponseSize))
	if err != nil {
		return "", errors.Wrap(err, "failed to read SimpleFIN claim response")
	}

	switch response.StatusCode {
	case http.StatusOK:
		// Do nothing
	case http.StatusForbidden:
		return "", errors.WithStack(ErrSetupTokenClaimed)
	default:
		return "", errors.Errorf("SimpleFIN claim request failed [%d]: %s", response.StatusCode, string(body))
	}

	// Don't include the parse error here, it would have the credentials in it.
	accessUrl := strings.TrimSpace(string(body))
	parsedAccessUrl, err := url.Parse(accessUrl)
	if err != nil {
		return "", errors.New("SimpleFIN returned an invalid access URL")
	}

	// The access URL could point somewhere other than where we claimed the token,
	// so it needs to be checked against the allowlist too.
	if !configuration.IsAllowedUrl(accessUrl) {
		log.WarnContext(span.Context(),
			"rejected SimpleFIN access URL for a host that is not in the configured allowlist, please update your configuration if this host is valid!",
			"host", parsedAccessUrl.Host,
		)
		return "", errors.WithStack(ErrHostNotAllowed)
	}

	if _, ok := parsedAccessUrl.User.Password(); !ok {
		return "", errors.New("SimpleFIN access URL is missing credentials")
	}

	return accessUrl, nil
}

func NewSimpleFinClient(
	log *slog.Logger,
	accessUrl string,
	configuration config.SimpleFin,
) (SimpleFinClient, error) {
	if !configuration.Enabled {
		log.Error(
			"simplefin is not enabled on this server but the client is being instantiated!",
			"bug", true,
		)
		return nil, errors.New("SimpleFIN is not enabled on this server")
	}

	// Don't include the parse error here, it would have the credentials in it.
	parsedUrl, err := url.Parse(accessUrl)
	if err != nil {
		return nil, errors.New("SimpleFIN access URL is not valid")
	}

	if !configuration.IsAllowedUrl(accessUrl) {
		log.Warn(
			"rejected SimpleFIN access URL for a host that is not in the configured allowlist, please update your configuration if this host is valid!",
			"host", parsedUrl.Host,
		)
		return nil, errors.WithStack(ErrHostNotAllowed)
	}

	password, ok := parsedUrl.User.Password()
	if !ok {
		return nil, errors.New("SimpleFIN access URL is missing credentials")
	}
	username := parsedUrl.User.Username()

	// Take the credentials out of the URL and send them as a header instead. The
	// base URL gets logged and ends up in error messages, this way the password
	// is never in any of those places.
	baseUrl := *parsedUrl
	baseUrl.User = nil
	baseUrl.RawQuery = ""
	baseUrl.Fragment = ""

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		// The Bridge doesn't need redirects and we are sending credentials with
		// every request, so just don't follow any of them.
		CheckRedirect: func(request *http.Request, _ []*http.Request) error {
			log.Warn(
				"blocked SimpleFIN redirect",
				"host", baseUrl.Host,
				"redirectUrl", request.URL.Redacted(),
			)
			return errors.Errorf("blocked SimpleFIN redirect to %s", request.URL.Redacted())
		},
		Transport: round.NewObservabilityRoundTripper(http.DefaultTransport,
			func(
				ctx context.Context,
				request *http.Request,
				response *http.Response,
				_ error,
				_ time.Duration,
			) {
				requestLog := log.With(
					"simplefin_method", request.Method,
					"simplefin_url", request.URL.String(),
				)
				var statusCode int
				if response != nil {
					statusCode = response.StatusCode
					requestLog = requestLog.With("simplefin_statusCode", statusCode)
				}

				crumbs.HTTP(ctx,
					"SimpleFIN API Call",
					"simplefin",
					request.URL.String(),
					request.Method,
					statusCode,
					map[string]any{},
				)
				requestLog.DebugContext(ctx, "SimpleFIN API call")
			}),
	}

	return &simpleFinClient{
		log:        log,
		baseUrl:    baseUrl,
		username:   username,
		password:   password,
		httpClient: httpClient,
	}, nil
}

// GetAccounts retrieves every account, balance and transaction the access URL
// can see in a single request. SimpleFIN only allows about 24 of these per day
// before it starts warning and eventually disables the token, so callers need
// to be careful about how often this is called.
// See: https://beta-bridge.simplefin.org/info/developers
func (s *simpleFinClient) GetAccounts(
	ctx context.Context,
	options GetAccountsOptions,
) (*AccountSet, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	requestUrl := s.baseUrl
	requestUrl.Path = path.Join(requestUrl.Path, "/accounts")
	query := url.Values{}
	query.Set("version", "2")
	if !options.StartDate.IsZero() {
		query.Set("start-date", strconv.FormatInt(options.StartDate.Unix(), 10))
	}
	if options.BalancesOnly {
		query.Set("balances-only", "1")
	}
	requestUrl.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(
		span.Context(),
		http.MethodGet,
		requestUrl.String(),
		http.NoBody,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP request")
	}
	request.SetBasicAuth(s.username, s.password)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send HTTP request")
	}
	defer response.Body.Close()

	body := io.LimitReader(response.Body, maxResponseBodySize)
	switch response.StatusCode {
	case http.StatusOK:
		// Do nothing
	case http.StatusPaymentRequired:
		return nil, errors.WithStack(ErrPaymentRequired)
	case http.StatusForbidden:
		return nil, errors.WithStack(ErrForbidden)
	default:
		bodyStr, _ := io.ReadAll(body)
		return nil, errors.Errorf("SimpleFIN request failed %s [%d]: %s", requestUrl.String(), response.StatusCode, string(bodyStr))
	}

	var result AccountSet
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, errors.Wrapf(err, "failed to decode response for request %s [%d]", requestUrl.String(), response.StatusCode)
	}

	return &result, nil
}
