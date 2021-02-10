package imredis

import (
	"context"
	"errors"

	"github.com/gomodule/redigo/redis"
)

// ImageIDStorage implements storage.ImageIDStorage.
type ImageIDStorage struct {
	kvp *redis.Pool
}

// NewImageIDStorage initializes a redis storage that implements
// storage.ImageIDStorage interface.
func NewImageIDStorage(kvp *redis.Pool) *ImageIDStorage {
	return &ImageIDStorage{
		kvp: kvp,
	}
}

func (s ImageIDStorage) List(
	ctx context.Context,
	category string,
	offset int,
	count int,
) (ids []string, err error) {
	return nil, errors.New("unimplemented")
}

func (s ImageIDStorage) Insert(ctx context.Context, id string) (err error) {
	return errors.New("unimplemented")
}

func (s ImageIDStorage) Delete(ctx context.Context, id string) (err error) {
	return errors.New("unimplemented")
}
