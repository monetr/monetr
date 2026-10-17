package cache

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"log/slog"

	"github.com/alicebob/miniredis/v2"
	"github.com/getsentry/sentry-go"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/certs"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	_ Cache = &valkeyCache{}
)

type ValkeyController struct {
	mini         *miniredis.Miniredis
	pool         *redis.Pool
	certificates certs.Source
}

func NewValkeyController(
	log *slog.Logger,
	conf config.Redis,
) (*ValkeyController, error) {
	controller := &ValkeyController{}
	var redisAddress string
	var err error
	dialOptions := []redis.DialOption{
		redis.DialUsername(conf.Username),
		redis.DialPassword(conf.Password),
		redis.DialDatabase(conf.Database),
	}
	if conf.Enabled {
		redisAddress = net.JoinHostPort(conf.Address, strconv.Itoa(conf.Port))
		log.DebugContext(context.Background(), fmt.Sprintf("connecting to redis at: %s", redisAddress))

		// If no CA certificate is provided then the server certificate is
		// verified against the system certificate authorities instead.
		if conf.TLS ||
			conf.CACertificatePath != "" ||
			conf.CertificatePath != "" ||
			conf.KeyPath != "" {
			controller.certificates, err = certs.NewFileSource(log, certs.Options{
				CACertificatePath:  conf.CACertificatePath,
				CertificatePath:    conf.CertificatePath,
				KeyPath:            conf.KeyPath,
				ServerName:         conf.Address,
				InsecureSkipVerify: conf.InsecureSkipVerify,
			})
			if err != nil {
				log.ErrorContext(context.Background(), "failed to load redis TLS certificates", "err", err)
				return nil, errors.Wrap(err, "failed to load redis TLS certificates")
			}
			// The certificates are watched for the lifetime of the controller, new
			// connections will pick up rotated certificates without swapping the
			// TLS config.
			if err = controller.certificates.Start(); err != nil {
				return nil, errors.Wrap(err, "failed to watch redis TLS certificates")
			}

			dialOptions = append(dialOptions,
				redis.DialUseTLS(true),
				redis.DialTLSConfig(controller.certificates.ClientConfig()),
			)
		}
	} else {
		controller.mini, err = miniredis.Run()
		if err != nil {
			return nil, errors.Wrap(err, "failed to run miniredis")
		}

		if conf.Username != "" && conf.Password != "" {
			controller.mini.RequireUserAuth(conf.Username, conf.Password)
		} else if conf.Password != "" {
			controller.mini.RequireAuth(conf.Password)
		}

		// Store our "embedded" redis address for use below.
		redisAddress = controller.mini.Server().Addr().String()
		log.InfoContext(context.Background(), "no redis config was provided, using miniredis!")
	}

	// Setup the redis pool for running jobs.
	controller.pool = &redis.Pool{
		MaxIdle:   10,
		MaxActive: 50,
		Dial: func() (redis.Conn, error) {
			// TODO (elliotcourant) Eventually support other networks besides
			//  tcp? Can redis even run on a unix socket?
			return redis.Dial("tcp", redisAddress, dialOptions...)
		},
	}

	// This will try to ping redis to make sure its up and running.
	if err = waitForRedis(log, 10, controller.pool); err != nil {
		log.ErrorContext(context.Background(), "failed to wait for redis to be available", "err", err)
		return nil, err
	}

	log.DebugContext(context.Background(), "successfully setup redis pool")

	return controller, nil
}

func waitForRedis(log *slog.Logger, maxAttempts int, pool *redis.Pool) error {
	for i := range maxAttempts {
		log.Log(context.Background(), logging.LevelTrace, "pinging redis")
		result, err := pool.Get().Do("PING")
		if err != nil {
			log.ErrorContext(context.Background(), fmt.Sprintf("failed to ping redis, attempt: %d", i+1), "err", err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		log.DebugContext(context.Background(), fmt.Sprintf("response from redis: %v", result))
		return nil
	}

	return errors.Errorf("failed to connect to redis after %d attempt(s)", maxAttempts)
}

func (r *ValkeyController) Pool() *redis.Pool {
	return r.pool
}

func (r *ValkeyController) Close() error {
	err := r.pool.Close()
	if r.mini != nil {
		r.mini.Close()
	}
	if r.certificates != nil {
		r.certificates.Stop()
	}
	return errors.Wrap(err, "failed to close pool gracefully")
}

var (
	ErrBlankKey = errors.New("key is blank")
)

type valkeyCache struct {
	log          *slog.Logger
	client       *redis.Pool
	certificates certs.Source
}

// NewValkeyCacheOld
// Deprecated: Use [NewValkeyCache] instead!
func NewValkeyCacheOld(log *slog.Logger, client *redis.Pool) Cache {
	return &valkeyCache{
		log:    log,
		client: client,
	}
}

func NewValkeyCache(
	ctx context.Context,
	log *slog.Logger,
	conf config.Valkey,
) (Cache, error) {
	cache := &valkeyCache{
		log: log,
	}
	var address string
	var err error
	dialOptions := []redis.DialOption{
		redis.DialUsername(conf.Username),
		redis.DialPassword(conf.Password),
		redis.DialDatabase(conf.Database),
	}

	address = net.JoinHostPort(conf.Address, strconv.Itoa(conf.Port))
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

		dialOptions = append(dialOptions,
			redis.DialUseTLS(true),
			redis.DialTLSConfig(cache.certificates.ClientConfig()),
		)
	}

	// Setup the redis pool for running jobs.
	cache.client = &redis.Pool{
		MaxIdle:   10,
		MaxActive: 50,
		Dial: func() (redis.Conn, error) {
			// TODO (elliotcourant) Eventually support other networks besides tcp? Can
			// valkey even run on a unix socket?
			return redis.Dial("tcp", address, dialOptions...)
		},
	}

	// This will try to ping valkey to make sure its up and running.
	if err = waitForRedis(log, 10, cache.client); err != nil {
		log.ErrorContext(ctx, "failed to wait for valkey to be available", "err", err)
		return nil, err
	}

	if err := prometheus.Register(NewPoolCollector(cache.client)); err != nil {
		var already prometheus.AlreadyRegisteredError
		if !errors.As(err, &already) {
			return nil, errors.Wrap(err, "failed to register valkey pool metrics")
		}
		log.WarnContext(ctx,
			"valkey pool metrics collector could not be registered, there might be missing metrics!",
			"err", err,
		)
	}

	log.DebugContext(ctx, "successfully setup valkey pool")

	return cache, nil
}

func (r *valkeyCache) do(ctx context.Context, commandName string, args ...any) (any, error) {
	conn, err := r.client.GetContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve connection from pool")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			r.log.WarnContext(ctx, "failed to close/release redis connection", "err", err)
		}
	}()

	return conn.Do(commandName, args...)
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
	if _, err := r.do(span.Context(), "SET", key, value); err != nil {
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
	if _, err := r.do(
		span.Context(),
		"SET", key, value, "EXAT", time.Now().Add(lifetime).Unix(),
	); err != nil {
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
	result, err := r.do(span.Context(), "GET", key)
	if err != nil {
		span.SetData("cache.success", false)
		span.Status = sentry.SpanStatusInternalError
		recordOperation("get", "error", start)
		return nil, errors.Wrap(err, "failed to retrieve item from cache")
	}
	span.SetData("cache.success", true)

	switch raw := result.(type) {
	case nil:
		span.SetData("cache.hit", false)
		span.Status = sentry.SpanStatusNotFound
		recordOperation("get", "miss", start)
		return nil, nil
	case string:
		span.SetData("cache.hit", true)
		span.SetData("cache.item_size", len(raw))
		recordOperation("get", "hit", start)
		return []byte(raw), nil
	case []byte:
		span.SetData("cache.hit", true)
		span.SetData("cache.item_size", len(raw))
		recordOperation("get", "hit", start)
		return raw, nil
	default:
		span.Status = sentry.SpanStatusUnimplemented
		panic(fmt.Sprintf("unsupported cache value type: %T", raw))
	}
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
	result, err := r.do(
		span.Context(),
		"EVAL", compareAndSwapScript, 1, key, expected, next, ttl.Milliseconds(),
	)
	if err != nil {
		span.SetData("cache.success", false)
		span.Status = sentry.SpanStatusInternalError
		recordOperation("compare_and_swap", "error", start)
		return false, errors.Wrap(err, "failed to compare and swap item in cache")
	}
	span.SetData("cache.success", true)

	swapped, err := redis.Int64(result, nil)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		recordOperation("compare_and_swap", "error", start)
		return false, errors.Wrap(err, "failed to interpret compare and swap result")
	}

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
	if _, err := r.do(span.Context(), "DEL", key); err != nil {
		span.SetData("cache.success", false)
		recordOperation("delete", "error", start)
		return errors.Wrap(err, "failed to delete item from cache")
	}

	span.SetData("cache.success", true)
	recordOperation("delete", "ok", start)
	return nil
}

func (r *valkeyCache) Close() error {
	err := r.client.Close()
	if r.certificates != nil {
		r.certificates.Stop()
	}
	return errors.Wrap(err, "failed to close pool gracefully")
}
