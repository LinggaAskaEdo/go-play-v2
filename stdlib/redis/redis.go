package redis

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
)

type Options struct {
	Enabled            bool
	Address            []string
	Password           string
	MaxRetries         int
	MinRetryBackoff    time.Duration
	MaxRetryBackoff    time.Duration
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolSize           int
	MinIdleConns       int
	MaxConnAge         time.Duration
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
	MaxRedirects       int
	ReadOnly           bool
	RouteByLatency     bool
	RouteRandomly      bool
}

func Init(logger zerolog.Logger, opt Options) *redis.Client {
	if !opt.Enabled {
		return nil
	}

	univOptions := &redis.UniversalOptions{
		Addrs:              opt.Address,
		Password:           opt.Password,
		MaxRetries:         opt.MaxRetries,
		MinRetryBackoff:    opt.MinRetryBackoff,
		MaxRetryBackoff:    opt.MaxRetryBackoff,
		DialTimeout:        opt.DialTimeout,
		ReadTimeout:        opt.ReadTimeout,
		WriteTimeout:       opt.WriteTimeout,
		PoolSize:           opt.PoolSize,
		MinIdleConns:       opt.MinIdleConns,
		MaxConnAge:         opt.MaxConnAge,
		PoolTimeout:        opt.PoolTimeout,
		IdleTimeout:        opt.IdleTimeout,
		IdleCheckFrequency: opt.IdleCheckFrequency,
		MaxRedirects:       opt.MaxRedirects,
		ReadOnly:           opt.ReadOnly,
		RouteByLatency:     opt.RouteByLatency,
		RouteRandomly:      opt.RouteRandomly,
	}

	univClient := redis.NewUniversalClient(univOptions)
	redisClient := univClient.(*redis.Client)

	ctx := redisClient.Context()
	ping, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Panic().Err(err).Str("when", "init").Send()
	}

	logger.Debug().Str("redis_status", ping).Send()

	return redisClient
}
