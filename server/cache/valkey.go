package cache

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/certs"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/pkg/errors"
	"github.com/valkey-io/valkey-go"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	_ Cache = &valkeyCache{}
)

// waitForValkey creates the client, retrying until valkey is reachable.
// valkey-go dials and handshakes as part of creating the client, so a failure
// here means valkey is not available (yet).
func waitForValkey(
	ctx context.Context,
	log *slog.Logger,
	maxAttempts int,
	options valkey.ClientOption,
) (valkey.Client, error) {
	for i := range maxAttempts {
		log.Log(ctx, logging.LevelTrace, "connecting to valkey")
		client, err := valkey.NewClient(options)
		if err != nil {
			log.ErrorContext(ctx, fmt.Sprintf("failed to connect to valkey, attempt: %d", i+1), "err", err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		return client, nil
	}

	return nil, errors.Errorf("failed to connect to valkey after %d attempt(s)", maxAttempts)
}

var (
	ErrBlankKey = errors.New("key is blank")
)

const (
	valkeyConnectTimeout   = 5 * time.Second
	valkeyReadWriteTimeout = 3 * time.Second
)

type valkeyCache struct {
	log          *slog.Logger
	client       valkey.Client
	certificates certs.Source
}

func NewValkeyCache(
	ctx context.Context,
	log *slog.Logger,
	conf config.Valkey,
) (Cache, error) {
	cache := &valkeyCache{
		log: log,
	}
	var err error
	address := net.JoinHostPort(conf.Address, strconv.Itoa(conf.Port))
	options := valkey.ClientOption{
		InitAddress: []string{address},
		Username:    conf.Username,
		Password:    conf.Password,
		SelectDB:    conf.Database,
		// Without these a valkey that stops responding would hang whatever request
		// or job is waiting on the cache. Nothing here issues blocking commands, so
		// every call should come back well within these.
		Dialer: net.Dialer{
			Timeout: valkeyConnectTimeout,
		},
		ConnWriteTimeout: valkeyReadWriteTimeout,
		// We only ever talk to a single valkey node, don't let the client try to
		// discover a cluster topology.
		ForceSingleClient: true,
		// Client side caching relies on CLIENT TRACKING and nothing here uses
		// DoCache anyway.
		DisableCache: true,
	}

	log.DebugContext(ctx, "connecting to valkey", "address", address)

	// If no CA certificate is provided then the server certificate is verified
	// against the system certificate authorities instead.
	if conf.TLS ||
		conf.CACertificatePath != "" ||
		conf.CertificatePath != "" ||
		conf.KeyPath != "" {
		cache.certificates, err = certs.NewFileSource(log, certs.Options{
			CACertificatePath:  conf.CACertificatePath,
			CertificatePath:    conf.CertificatePath,
			KeyPath:            conf.KeyPath,
			ServerName:         conf.Address,
			InsecureSkipVerify: conf.InsecureSkipVerify,
		})
		if err != nil {
			log.ErrorContext(ctx, "failed to load valkey TLS certificates", "err", err)
			return nil, errors.Wrap(err, "failed to load valkey TLS certificates")
		}
		// The certificates are watched for the lifetime of the controller, new
		// connections will pick up rotated certificates without swapping the TLS
		// config.
		if err = cache.certificates.Start(); err != nil {
			return nil, errors.Wrap(err, "failed to watch valkey TLS certificates")
		}

		options.TLSConfig = cache.certificates.ClientConfig()
	}

	cache.client, err = waitForValkey(ctx, log, 10, options)
	if err != nil {
		log.ErrorContext(ctx, "failed to wait for valkey to be available", "err", err)
		if cache.certificates != nil {
			cache.certificates.Stop()
		}
		return nil, err
	}

	log.DebugContext(ctx, "successfully setup valkey client")

	return cache, nil
}

func (r *valkeyCache) Set(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}
	span := sentry.StartSpan(ctx, "cache.put")
	defer span.Finish()
	span.Description = key
	span.SetData("db.system", "redis")
	span.SetData("cache.key", []string{key})
	span.SetData("cache.item_size", len(value))

	start := time.Now()
	if err := r.client.Do(
		span.Context(),
		r.client.B().Set().Key(key).Value(valkey.BinaryString(value)).Build(),
	).Error(); err != nil {
		span.SetData("cache.success", false)
		recordOperation("set", "error", start)
		return errors.Wrap(err, "failed to store item in cache")
	}

	span.SetData("cache.success", true)
	recordOperation("set", "ok", start)
	return nil
}

func (r *valkeyCache) SetTTL(ctx context.Context, key string, value []byte, lifetime time.Duration) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}
	span := sentry.StartSpan(ctx, "cache.put")
	defer span.Finish()
	span.Description = key
	span.SetData("db.system", "redis")
	span.SetData("cache.key", []string{key})
	span.SetData("cache.item_size", len(value))
	span.SetData("cache.ttl", int64(lifetime.Seconds()))

	start := time.Now()
	if err := r.client.Do(
		span.Context(),
		r.client.B().Set().
			Key(key).
			Value(valkey.BinaryString(value)).
			ExatTimestamp(time.Now().Add(lifetime).Unix()).
			Build(),
	).Error(); err != nil {
		span.SetData("cache.success", false)
		recordOperation("set", "error", start)
		return errors.Wrap(err, "failed to store item in cache")
	}

	span.SetData("cache.success", true)
	recordOperation("set", "ok", start)
	return nil
}

func (r *valkeyCache) SetEz(ctx context.Context, key string, object any) error {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = ""

	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return r.Set(span.Context(), key, data)
}

func (r *valkeyCache) SetEzTTL(ctx context.Context, key string, object any, lifetime time.Duration) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return r.SetTTL(span.Context(), key, data, lifetime)
}

func (r *valkeyCache) Get(ctx context.Context, key string) ([]byte, error) {
	span := sentry.StartSpan(ctx, "cache.get")
	defer span.Finish()

	if key == "" {
		span.Status = sentry.SpanStatusInvalidArgument
		return nil, errors.WithStack(ErrBlankKey)
	}

	span.Status = sentry.SpanStatusOK
	span.Description = key
	span.SetData("cache.key", []string{key})
	span.SetData("db.system", "redis")

	start := time.Now()
	result, err := r.client.Do(
		span.Context(),
		r.client.B().Get().Key(key).Build(),
	).AsBytes()
	if valkey.IsValkeyNil(err) {
		span.SetData("cache.success", true)
		span.SetData("cache.hit", false)
		span.Status = sentry.SpanStatusNotFound
		recordOperation("get", "miss", start)
		return nil, nil
	} else if err != nil {
		span.SetData("cache.success", false)
		span.Status = sentry.SpanStatusInternalError
		recordOperation("get", "error", start)
		return nil, errors.Wrap(err, "failed to retrieve item from cache")
	}
	span.SetData("cache.success", true)
	span.SetData("cache.hit", true)
	span.SetData("cache.item_size", len(result))
	recordOperation("get", "hit", start)
	return result, nil
}

func (r *valkeyCache) GetEz(ctx context.Context, key string, output any) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Status = sentry.SpanStatusOK

	data, err := r.Get(span.Context(), key)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err = msgpack.Unmarshal(data, output); err != nil {
		span.Status = sentry.SpanStatusDataLoss
		return errors.Wrap(err, "failed to unmarshal from cache")
	}

	return nil
}

// Lua for CompareAndSwap, run atomically by redis (no read-then-write race).
// KEYS[1] key, ARGV[1] expected, ARGV[2] next, ARGV[3] ttl ms. Returns 1 on
// swap, 0 otherwise (no write on a mismatch or missing key).
const compareAndSwapScript = `
local current = redis.call('GET', KEYS[1])
if current == false then
	return 0
end
if current == ARGV[1] then
	local ttl = tonumber(ARGV[3])
	if ttl and ttl > 0 then
		redis.call('SET', KEYS[1], ARGV[2], 'PX', ttl)
	else
		redis.call('SET', KEYS[1], ARGV[2], 'KEEPTTL')
	end
	return 1
end
return 0
`

func (r *valkeyCache) CompareAndSwap(ctx context.Context, key string, expected, next []byte, ttl time.Duration) (bool, error) {
	if key == "" {
		return false, errors.WithStack(ErrBlankKey)
	}

	span := sentry.StartSpan(ctx, "cache.compare_and_swap")
	defer span.Finish()
	span.Description = key
	span.Status = sentry.SpanStatusOK
	span.SetData("db.system", "redis")
	span.SetData("cache.key", []string{key})

	start := time.Now()
	swapped, err := r.client.Do(
		span.Context(),
		r.client.B().Eval().
			Script(compareAndSwapScript).
			Numkeys(1).
			Key(key).
			Arg(
				valkey.BinaryString(expected),
				valkey.BinaryString(next),
				strconv.FormatInt(ttl.Milliseconds(), 10),
			).
			Build(),
	).AsInt64()
	if err != nil {
		span.SetData("cache.success", false)
		span.Status = sentry.SpanStatusInternalError
		recordOperation("compare_and_swap", "error", start)
		return false, errors.Wrap(err, "failed to compare and swap item in cache")
	}
	span.SetData("cache.success", true)

	span.SetData("cache.swapped", swapped == 1)
	if swapped == 1 {
		recordOperation("compare_and_swap", "swapped", start)
	} else {
		recordOperation("compare_and_swap", "not_swapped", start)
	}
	return swapped == 1, nil
}

func (r *valkeyCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}

	span := sentry.StartSpan(ctx, "cache.remove")
	defer span.Finish()
	span.Description = key
	span.Status = sentry.SpanStatusOK
	span.SetData("db.system", "redis")
	span.SetData("cache.key", []string{key})

	start := time.Now()
	if err := r.client.Do(
		span.Context(),
		r.client.B().Del().Key(key).Build(),
	).Error(); err != nil {
		span.SetData("cache.success", false)
		recordOperation("delete", "error", start)
		return errors.Wrap(err, "failed to delete item from cache")
	}

	span.SetData("cache.success", true)
	recordOperation("delete", "ok", start)
	return nil
}

func (r *valkeyCache) Close() error {
	r.client.Close()
	if r.certificates != nil {
		r.certificates.Stop()
	}
	return nil
}
