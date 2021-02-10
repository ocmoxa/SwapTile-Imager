package imredis

import (
	"context"
	"errors"

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
	return nil, errors.New("unimplemented")
}

func (r CategoryRepository) Delete(ctx context.Context, name string) (err error) {
	return errors.New("unimplemented")
}

func (r CategoryRepository) Upsert(ctx context.Context, name string) (err error) {
	return errors.New("unimplemented")
}
