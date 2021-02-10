// Package test contains general test helpers
package test

import (
	"os"
	"testing"

	"github.com/gomodule/redigo/redis"
)

func AssertErrNil(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

// InitKVP initialized *redis.Pool. It skips test if env variable is not
// set.
func InitKVP(t *testing.T) *redis.Pool {
	const envVar = "TEST_IMAGE_REDIS"

	t.Helper()

	addr := os.Getenv(envVar)
	if addr == "" {
		t.Skip("env variable", envVar, "not set")

		return nil
	}

	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(addr)
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
