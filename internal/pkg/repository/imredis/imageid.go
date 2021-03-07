package imredis

import (
	"context"
	"fmt"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"

	"github.com/gomodule/redigo/redis"
)

// ImageIDRepository implements storage.ImageIDRepository.
type ImageIDRepository struct {
	kvp *redis.Pool
}

// NewImageIDRepository initializes a redis storage that implements
// storage.ImageIDRepository interface.
func NewImageIDRepository(kvp *redis.Pool) *ImageIDRepository {
	return &ImageIDRepository{
		kvp: kvp,
	}
}

// Set upserts image ID.
func (r ImageIDRepository) Set(ctx context.Context, id string) (ok bool, err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	ok, err = redis.Bool(kv.Do("SADD", keyImageID, id))
	if err != nil {
		return false, fmt.Errorf("doing sadd: %w", err)
	}

	return ok, nil
}

// Delete image ID.
func (r ImageIDRepository) Delete(ctx context.Context, id string) (err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	_, err = kv.Do("SREM", keyImageID, id)
	if err != nil {
		return fmt.Errorf("doing srem: %w", err)
	}

	return nil
}
