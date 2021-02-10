package imredis

import (
	"context"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"

	"github.com/gomodule/redigo/redis"
)

// CategoryRepository implements storage.CategoryRepository.
type CategoryRepository struct {
	kvp *redis.Pool
}

// NewCategoryRepository initializes a redis storage that implements
// storage.CategoryRepository interface.
func NewCategoryRepository(kvp *redis.Pool) *CategoryRepository {
	return &CategoryRepository{
		kvp: kvp,
	}
}

func (r CategoryRepository) List(ctx context.Context) (names []string, err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	return redis.Strings(kv.Do("SMEMBERS", keyCategories))
}

func (r CategoryRepository) Delete(ctx context.Context, name string) (err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	_, err = kv.Do("SREM", keyCategories, name)
	return err
}

func (r CategoryRepository) Upsert(ctx context.Context, name string) (err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	_, err = kv.Do("SADD", keyCategories, name)
	return err
}
