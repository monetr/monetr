package platypus

import (
	"context"
	"encoding/json"
	"github.com/MicahParks/keyfunc"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
	"time"
)

type WebhookVerificationKey struct {
	// The alg member identifies the cryptographic algorithm family used with the key.
	Alg string `json:"alg"`
	// The crv member identifies the cryptographic curve used with the key.
	Crv string `json:"crv"`
	// The kid (Key ID) member can be used to match a specific key. This can be used, for instance, to choose among a set of keys within the JWK during key rollover.
	Kid string `json:"kid"`
	// The kty (key type) parameter identifies the cryptographic algorithm family used with the key, such as RSA or EC.
	Kty string `json:"kty"`
	// The use (public key use) parameter identifies the intended use of the public key.
	Use string `json:"use"`
	// The x member contains the x coordinate for the elliptic curve point.
	X string `json:"x"`
	// The y member contains the y coordinate for the elliptic curve point.
	Y         string `json:"y"`
	CreatedAt int32  `json:"created_at"`
	ExpiredAt *int32 `json:"expired_at"`
}

func NewWebhookVerificationKeyFromPlaid(input plaid.JWKPublicKey) (WebhookVerificationKey, error) {
	return WebhookVerificationKey{
		Alg:       input.GetAlg(),
		Crv:       input.GetCrv(),
		Kid:       input.GetKid(),
		Kty:       input.GetKty(),
		Use:       input.GetUse(),
		X:         input.GetX(),
		Y:         input.GetY(),
		CreatedAt: input.GetCreatedAt(),
		ExpiredAt: myownsanity.Int32P(input.GetExpiredAt()),
	}, nil
}

type WebhookVerification interface {
	GetVerificationKey(ctx context.Context, keyId string) (*keyfunc.JWKS, error)
	Close() error
}

var (
	_ WebhookVerification = &memoryWebhookVerification{}
)

func NewInMemoryWebhookVerification(log *logrus.Entry, plaid Platypus, cleanupInterval time.Duration) WebhookVerification {
	verification := &memoryWebhookVerification{
		closed:        0,
		log:           log,
		plaid:         plaid,
		lock:          sync.Mutex{},
		cache:         map[string]*keyCacheItem{},
		cleanupTicker: time.NewTicker(cleanupInterval),
		closer:        make(chan chan error, 1),
	}
	go verification.cacheWorker() // Start the background worker.

	return verification
}

type keyCacheItem struct {
	expiration  time.Time
	keyFunction *keyfunc.JWKS
}

type memoryWebhookVerification struct {
	closed        uint32
	log           *logrus.Entry
	plaid         Platypus
	lock          sync.Mutex
	cache         map[string]*keyCacheItem
	cleanupTicker *time.Ticker
	closer        chan chan error
}

func (m *memoryWebhookVerification) GetVerificationKey(ctx context.Context, keyId string) (*keyfunc.JWKS, error) {
	if atomic.LoadUint32(&m.closed) > 0 {
		return nil, errors.New("webhook verification is closed")
	}

	span := sentry.StartSpan(ctx, "GetVerificationKey [InMemory]")
	defer span.Finish()

	log := m.log.WithField("keyId", keyId).WithContext(ctx)

	m.lock.Lock()
	defer m.lock.Unlock()

	item, ok := m.cache[keyId]
	if ok {
		if item.expiration.After(time.Now()) {
			log.Trace("jwk function already present in cache, returning")
			return item.keyFunction, nil
		}

		log.Trace("jwk function present in cache, but is expired; the cached function will be removed and a new one will be retrieved")
		delete(m.cache, keyId)
	}

	log.Trace("retrieving jwk from plaid")

	result, err := m.plaid.GetWebhookVerificationKey(span.Context(), keyId)
	if err != nil {
		return nil, err
	}

	var expiration time.Time
	if result.ExpiredAt != nil {
		expiration = time.Unix(int64(*result.ExpiredAt), 0)
	} else {
		// Making a huge assumption here, and this might end up causing problems later on. Maybe we should also add a
		// check here to make sure that items that are close to expiration even here should not be cached?
		expiration = time.Unix(int64(result.CreatedAt), 0).Add(30 * time.Minute)
	}

	var keys = struct {
		Keys []WebhookVerificationKey `json:"keys"`
	}{
		Keys: []WebhookVerificationKey{
			*result,
		},
	}

	encodedKeys, err := json.Marshal(keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert plaid verification key to json")
	}

	var jwksJSON json.RawMessage = encodedKeys

	jwksFunc, err := keyfunc.New(jwksJSON)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create key function")
	}

	m.cache[keyId] = &keyCacheItem{
		expiration:  expiration,
		keyFunction: jwksFunc,
	}

	return jwksFunc, nil
}

func (m *memoryWebhookVerification) cacheWorker() {
	for {
		select {
		case _ = <-m.cleanupTicker.C:
			m.cleanup()
		case promise := <-m.closer:
			m.log.Debug("closing jwk cache, stopping background worker")
			promise <- nil
			return
		}
	}
}

func (m *memoryWebhookVerification) cleanup() {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(m.cache) == 0 {
		m.log.Debug("no items in Plaid jwk cache, nothing to cleanup")
		return
	}

	m.log.Debug("cleaning up Plaid jwk cache")

	itemsToRemove := make([]string, 0, len(m.cache))
	for key, item := range m.cache {
		// If the item expiration is not in the future, then we need to add it to our list to be removed.
		if !item.expiration.After(time.Now()) {
			itemsToRemove = append(itemsToRemove, key)
		}
	}

	if len(itemsToRemove) == 0 {
		m.log.Debug("no items have expired in cache")
		return
	}

	m.log.Debugf("found %d expired item(s); cleaning them up", len(itemsToRemove))

	for _, key := range itemsToRemove {
		delete(m.cache, key)
	}

	return
}

func (m *memoryWebhookVerification) Close() error {
	if ok := atomic.CompareAndSwapUint32(&m.closed, 0, 1); !ok {
		return errors.New("webhook verification is already closed")
	}

	promise := make(chan error)
	m.closer <- promise

	return <-promise
}
