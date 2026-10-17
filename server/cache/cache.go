package cache

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/pkg/errors"
)

type Cache interface {
	Set(ctx context.Context, key string, value []byte) error
	SetTTL(ctx context.Context, key string, value []byte, lifetime time.Duration) error
	SetEz(ctx context.Context, key string, object any) error
	SetEzTTL(ctx context.Context, key string, object any, lifetime time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	GetEz(ctx context.Context, key string, output any) error
	Delete(ctx context.Context, key string) error
	// CompareAndSwap atomically sets key to next only if its current value equals
	// expected; a missing key or mismatch writes nothing and returns false. On a
	// swap the expiry is refreshed to ttl, or kept (KEEPTTL) when ttl <= 0.
	CompareAndSwap(ctx context.Context, key string, expected, next []byte, ttl time.Duration) (swapped bool, err error)
	Close() error
}

func GetCache(
	ctx context.Context,
	log *slog.Logger,
	clock clock.Clock,
	configuration config.Configuration,
) (Cache, error) {
	log.Log(ctx, logging.LevelTrace, "setting up caching interface")

	// Migration from the old config to the new, if the user does not have kind
	// configured right now then we will show a warning telling them to configure
	// it. In the mean time we will try to infer it from the existing (deprecated)
	// redis config
	kind := configuration.Cache.Kind
	if kind == "" {
		log.WarnContext(ctx,
			"Config warning! cache.kind is unset, monetr will fallback to the redis configuration for now but in a future update will default to an in memory cache if none is configured",
			"documentation", "https://monetr.app/documentation/configure/cache",
		)
		if configuration.Redis.Enabled {
			// Use the deprecated value to signal that it is the old config
			kind = "redis"
			// If redis is enabled then we will use the configuration from the old
			// config for now. As we get further along and move to deprecate the redis
			// configuration block, this will stop working and instead user's will
			// encounter an error or we will fall back to in memory caching
			configuration.Cache.Valkey.Address = configuration.Redis.Address
			configuration.Cache.Valkey.Port = configuration.Redis.Port
			configuration.Cache.Valkey.Database = configuration.Redis.Database
			configuration.Cache.Valkey.Username = configuration.Redis.Username
			configuration.Cache.Valkey.Password = configuration.Redis.Password
			configuration.Cache.Valkey.TLS = configuration.Redis.TLS
			configuration.Cache.Valkey.InsecureSkipVerify = configuration.Redis.InsecureSkipVerify
			configuration.Cache.Valkey.CACertificatePath = configuration.Redis.CACertificatePath
			configuration.Cache.Valkey.KeyPath = configuration.Redis.KeyPath
			configuration.Cache.Valkey.CertificatePath = configuration.Redis.CertificatePath
		} else {
			kind = "memory"
		}
	}

	switch strings.ToLower(configuration.Cache.Kind) {
	case "redis":
		log.WarnContext(ctx,
			"Config warning! redis.enabled is true and the config has not been migrated to use cache.valkey instead!",
			"documentation", "https://monetr.app/documentation/configure/cache",
		)
		fallthrough
	case "valkey":
		return NewValkeyCache(ctx, log, configuration.Cache.Valkey)
	case "memory":
		return NewLRUCache(log, clock, configuration.Cache.Memory), nil
	default:
		return nil, errors.Errorf("invalid cache kind configured: %s", kind)
	}
}
