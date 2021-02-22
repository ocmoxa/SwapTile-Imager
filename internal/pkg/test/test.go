// Package test contains general test helpers
package test

import (
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"

	"github.com/gomodule/redigo/redis"
)

func AssertErrNil(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

// LoadConfig loads config from the environment variables.
func LoadConfig(t *testing.T) config.Config {
	t.Helper()

	cfg, err := config.Load("")
	AssertErrNil(t, err)

	return cfg
}

// InitKVP initialized *redis.Pool. It skips test if env variable is not
// set.
func InitKVP(t *testing.T) *redis.Pool {
	t.Helper()

	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(LoadConfig(t).Redis.Endpoint)
		},
	}
}

// DisposeKVP closes *redis.Pool. Call it in defer.
func DisposeKVP(t *testing.T, kvp *redis.Pool) {
	t.Helper()

	if kvp == nil {
		return
	}

	err := kvp.Close()
	AssertErrNil(t, err)
}
