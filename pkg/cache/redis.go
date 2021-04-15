package cache

import (
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type RedisController struct {
	mini *miniredis.Miniredis
	pool *redis.Pool
}

func NewRedisCache(log *logrus.Entry, conf config.Redis) (*RedisController, error) {
	controller := &RedisController{}
	var redisAddress string
	var err error
	if conf.Enabled {
		redisAddress = fmt.Sprintf("%s:%d", conf.Address, conf.Port)
	} else {
		controller.mini, err = miniredis.Run()
		if err != nil {
			return nil, errors.Wrap(err, "failed to run miniredis")
		}

		// Store our "embedded" redis address for use below.
		redisAddress = controller.mini.Server().Addr().String()
	}

	// Setup the redis pool for running jobs.
	controller.pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			// TODO (elliotcourant) Eventually support other networks besides
			//  tcp? Can redis even run on a unix socket?
			return redis.Dial("tcp", redisAddress)
		},
	}

	// This will try to ping redis to make sure its up and running.
	if err = waitForRedis(log, 10, controller.pool); err != nil {
		return nil, err
	}

	return controller, nil
}

func waitForRedis(log *logrus.Entry, maxAttempts int, pool *redis.Pool) error {
	for i := 0; i < maxAttempts; i++ {
		log.Trace("pinging redis")
		result, err := pool.Get().Do("PING")
		if err != nil {
			log.WithError(err).Errorf("failed to ping redis, attempt: %d", i+1)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		log.Tracef("response from redis: %v", result)
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
