package imredis

import (
	"context"
	"errors"

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

func (r ImageIDRepository) List(
	ctx context.Context,
	category string,
	offset int,
	count int,
) (ids []string, err error) {
	return nil, errors.New("unimplemented")
}

func (r ImageIDRepository) Insert(ctx context.Context, id string) (err error) {
	return errors.New("unimplemented")
}

func (r ImageIDRepository) Delete(ctx context.Context, id string) (err error) {
	return errors.New("unimplemented")
}
