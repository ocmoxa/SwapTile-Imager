// Package test contains general test helpers
package test

import (
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"

	"github.com/gomodule/redigo/redis"
)

func AssertErrNil(tb testing.TB, err error) {
	tb.Helper()

	if err != nil {
		tb.Fatal(err)
	}
}

// LoadConfig loads config from the environment variables.
func LoadConfig(tb testing.TB) config.Config {
	tb.Helper()

	cfg, err := config.Load("")
	AssertErrNil(tb, err)

	return cfg
}

// InitKVP initialized *redis.Pool. It skips test if env variable is not
// set.
func InitKVP(tb testing.TB) *redis.Pool {
	tb.Helper()

	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(LoadConfig(tb).Redis.Endpoint)
		},
	}
}

// DisposeKVP closes *redis.Pool. Call it in defer.
func DisposeKVP(tb testing.TB, kvp *redis.Pool) {
	tb.Helper()

	if kvp == nil {
		return
	}

	err := kvp.Close()
	AssertErrNil(tb, err)
}
