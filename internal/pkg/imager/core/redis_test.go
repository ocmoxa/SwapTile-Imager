// +build integration

package core

import (
	"context"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"
)

func TestRedisHealthChecker(t *testing.T) {
	kvp := test.InitKVP(t)
	defer test.DisposeKVP(t, kvp)

	ctx := context.Background()
	rhc := redisHealthChecker{KVP: kvp}
	err := rhc.Health(ctx)
	test.AssertErrNil(t, err)
}
