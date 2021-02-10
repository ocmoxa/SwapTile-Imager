package imredis

import (
	"context"
	"errors"

	"github.com/gomodule/redigo/redis"
)

// CategoryStorage implements storage.CategoryStorage.
type CategoryStorage struct {
	kvp *redis.Pool
}

// NewCategorytorage initializes a redis storage that implements
// storage.CategoryStorage interface.
func NewCategorytorage(kvp *redis.Pool) *CategoryStorage {
	return &CategoryStorage{
		kvp: kvp,
	}
}

func (*CategoryStorage) List(ctx context.Context) (names []string, err error) {
	return nil, errors.New("unimplemented")
}

func (*CategoryStorage) Delete(ctx context.Context, name string) (err error) {
	return errors.New("unimplemented")
}

func (*CategoryStorage) Upsert(ctx context.Context, name string) (err error) {
	return errors.New("unimplemented")
}
