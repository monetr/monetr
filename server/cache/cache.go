package cache

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/pkg/errors"
)

type RedisController struct {
	mini *miniredis.Miniredis
	pool *redis.Pool
}

func NewRedisCache(
	log *slog.Logger,
	conf config.Redis,
) (*RedisController, error) {
	controller := &RedisController{}
	var redisAddress string
	var err error
	if conf.Enabled {
		redisAddress = net.JoinHostPort(conf.Address, strconv.Itoa(conf.Port))
		log.DebugContext(context.Background(), fmt.Sprintf("connecting to redis at: %s", redisAddress))
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
			return redis.Dial(
				"tcp",
				redisAddress,
				redis.DialUsername(conf.Username),
				redis.DialPassword(conf.Password),
				redis.DialDatabase(conf.Database),
			)
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

func (r *RedisController) Pool() *redis.Pool {
	return r.pool
}

func (r *RedisController) Close() error {
	err := r.pool.Close()
	if r.mini != nil {
		r.mini.Close()
	}
	return errors.Wrap(err, "failed to close pool gracefully")
}
