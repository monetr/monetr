package plaid_helper

import (
	"context"
	"encoding/json"
	"github.com/MicahParks/keyfunc"
	"github.com/getsentry/sentry-go"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"sync"
)

type WebhookVerification interface {
	GetVerificationKey(ctx context.Context, keyId string) (*keyfunc.JWKS, error)
	Close() error
}

var (
	_ WebhookVerification = &redisWebhookVerification{}
	_ WebhookVerification = &memoryWebhookVerification{}
)

func NewMemoryWebhookVerificationCache(log *logrus.Entry, client Client) WebhookVerification {
	return &memoryWebhookVerification{
		log:         log,
		plaidClient: client,
		lock:        sync.Mutex{},
		cache:       map[string]*keyfunc.JWKS{},
	}
}

type memoryWebhookVerification struct {
	log         *logrus.Entry
	plaidClient Client
	lock        sync.Mutex
	cache       map[string]*keyfunc.JWKS
}

func (m *memoryWebhookVerification) GetVerificationKey(ctx context.Context, keyId string) (*keyfunc.JWKS, error) {
	span := sentry.StartSpan(ctx, "GetVerificationKey")
	defer span.Finish()

	log := m.log.WithField("keyId", keyId).WithContext(ctx)

	m.lock.Lock()
	defer m.lock.Unlock()

	jwksFunc, ok := m.cache[keyId]
	if ok {
		log.Trace("jwk function already present in cache, returning")
		return jwksFunc, nil
	}

	log.Trace("jwk function missing in cache, retrieving from plaid")

	verificationResponse, err := m.plaidClient.GetWebhookVerificationKey(span.Context(), keyId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve public verification key")
	}

	var keys = struct {
		Keys []plaid.WebhookVerificationKey `json:"keys"`
	}{
		Keys: []plaid.WebhookVerificationKey{
			verificationResponse.Key,
		},
	}

	encodedKeys, err := json.Marshal(keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert plaid verification key to json")
	}

	var jwksJSON json.RawMessage = encodedKeys

	jwksFunc, err = keyfunc.New(jwksJSON)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create key function")
	}

	m.cache[keyId] = jwksFunc

	return jwksFunc, nil
}

func (m *memoryWebhookVerification) Close() error {
	return m.plaidClient.Close()
}

type redisWebhookVerification struct {
	log         *logrus.Entry
	plaidClient Client
	redisClient redis.Conn
}

// TODO Finish building out caching of public verification keys.
func (r *redisWebhookVerification) GetVerificationKey(ctx context.Context, keyId string) (*keyfunc.JWKS, error) {
	result, err := redis.String(r.redisClient.Do("GET", keyId))
	if err != nil {
		return nil, err
	}

	var jwksJSON json.RawMessage = []byte(result)
	jwkKeyFunc, err := keyfunc.New(jwksJSON)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create key function")
	}

	return jwkKeyFunc, nil
}

func (r *redisWebhookVerification) Close() error {
	if err := r.plaidClient.Close(); err != nil {
		r.log.WithError(err).Errorf("failed to close plaid client gracefully")
		return errors.Wrap(err, "failed to close plaid client")
	}
	if err := r.redisClient.Close(); err != nil {
		r.log.WithError(err).Errorf("failed to close redis client gracefully")
		return errors.Wrap(err, "failed to close redis client")
	}

	return nil
}
