package core

import (
	"context"
	"fmt"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"

	"github.com/gomodule/redigo/redis"
)

type redisHealthChecker struct {
	KVP *redis.Pool
}

// Health pings redis.
func (c redisHealthChecker) Health(context.Context) (err error) {
	kv := c.KVP.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	_, err = kv.Do("PING")
	if err != nil {
		return fmt.Errorf("redis: doing ping: %w", err)
	}

	return nil
}
