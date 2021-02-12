package imredis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/improto"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"

	"github.com/gomodule/redigo/redis"
	"google.golang.org/protobuf/proto"
)

// ImageMetaRepository implements storage.ImageMetaRepository.
type ImageMetaRepository struct {
	kvp *redis.Pool
}

// NewImageMetaRepository initializes a redis storage that implements
// storage.ImageMetaRepository interface.
func NewImageMetaRepository(kvp *redis.Pool) *ImageMetaRepository {
	return &ImageMetaRepository{
		kvp: kvp,
	}
}

func (r ImageMetaRepository) key(category string) string {
	return keyPrefixImageIDs + ":" + category
}

func (r ImageMetaRepository) List(
	ctx context.Context,
	category string,
	pagination repository.Pagination,
) (im []repository.IndexedImageMeta, err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	imStrs, err := redis.ByteSlices(kv.Do(
		"LRANGE",
		r.key(category),                      // Key.
		pagination.Offset,                    // Start, inclusive.
		pagination.Offset+pagination.Limit-1, // Stop, inclusive.
	))
	if err != nil {
		return nil, fmt.Errorf("doing lrange: %w", err)
	}

	im = make([]repository.IndexedImageMeta, len(imStrs))
	for i, imData := range imStrs {
		imProto := &improto.ImageMeta{}

		if err = proto.Unmarshal(imData, imProto); err != nil {
			return nil, fmt.Errorf("unmarshalling text: %w", err)
		}

		im[i] = repository.IndexedImageMeta{
			Index:     i + pagination.Offset,
			ImageMeta: improto.FromImageMeta(imProto),
		}
	}

	return im, nil
}

func (r ImageMetaRepository) Insert(
	ctx context.Context,
	category string,
	im imager.ImageMeta,
) (err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	imProto := improto.ToImageMetaProto(im)

	imData, err := proto.Marshal(imProto)
	if err != nil {
		return fmt.Errorf("marshalling text: %w", err)
	}

	_, err = kv.Do(
		"RPUSH",
		r.key(category), // Key.
		imData,          // Element.
	)
	if err != nil {
		return fmt.Errorf("doing rpush: %w", err)
	}

	return err
}

func (r ImageMetaRepository) Delete(
	ctx context.Context,
	category string,
	index int,
) (err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	element, err := redis.Bytes(kv.Do(
		"LINDEX",
		r.key(category), // Key.
		index,           // Index.
	))
	switch {
	case err != nil:
		return fmt.Errorf("doing lindex: %w", err)
	case errors.Is(err, redis.ErrNil):
		return nil
	}

	_, err = kv.Do(
		"LREM",
		r.key(category), // Key.
		0,               // Count. Remove all elements equal to element.
		element,         // Element.
	)
	if err != nil {
		return fmt.Errorf("doing lrem: %w", err)
	}

	return err
}

func (r ImageMetaRepository) Categories(
	ctx context.Context,
) (categories []string, err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	// It is assumed that there will be few categories, up to 100.
	// Warning: KEYS can block Redis for a significant amount of time
	// if the number of categories becomes large.
	categories, err = redis.Strings(kv.Do("KEYS", keyPrefixImageIDs+"*"))
	if err != nil {
		return nil, fmt.Errorf("doing keys: %w", err)
	}

	for i, c := range categories {
		categories[i] = strings.TrimPrefix(c, keyPrefixImageIDs)
	}

	return categories, nil
}
