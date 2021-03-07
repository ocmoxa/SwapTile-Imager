package imredis

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/improto"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"

	"github.com/gomodule/redigo/redis"
	"google.golang.org/protobuf/proto"
)

// CategoryNameAll is a name of category with all images.
const CategoryNameAll = "all"

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
	return keyPrefixImageMeta + category
}

// List of image meta.
func (r ImageMetaRepository) List(
	ctx context.Context,
	category string,
	pagination repository.Pagination,
) (im []repository.IndexedImageMeta, err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	imStrs, err := redis.ByteSlices(kv.Do(
		"LRANGE",
		r.key(category),
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
		im[i].Category = category
	}

	return im, nil
}

// Insert an image metadata.
func (r ImageMetaRepository) Insert(
	ctx context.Context,
	im imager.ImageMeta,
) (err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	imProto := improto.ToImageMetaProto(im)

	imData, err := proto.Marshal(imProto)
	if err != nil {
		return fmt.Errorf("marshalling text: %w", err)
	}

	p := newPipeline(kv)
	p.Send("MULTI")
	p.Send(
		"RPUSH",
		r.key(im.Category),
		imData, // Element.
	)
	p.Send(
		"RPUSH",
		r.key(CategoryNameAll),
		imData, // Element.
	)
	_, err = p.Do("EXEC")
	if err != nil {
		return fmt.Errorf("doing exec: %w", err)
	}

	return err
}

// Delete an image metadata by index in the category.
func (r ImageMetaRepository) Delete(
	ctx context.Context,
	category string,
	index int,
) (err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	element, err := redis.Bytes(kv.Do(
		"LINDEX",
		r.key(category),
		index,
	))
	switch {
	case err != nil:
		return fmt.Errorf("doing lindex: %w", err)
	case errors.Is(err, redis.ErrNil):
		return nil
	}

	p := newPipeline(kv)
	p.Send("MULTI")
	p.Send(
		"LREM",
		r.key(category),
		0, // Count. Remove all elements equal to element.
		element,
	)
	p.Send(
		"LREM",
		r.key(CategoryNameAll),
		0, // Count. Remove all elements equal to element.
		element,
	)
	_, err = p.Do("EXEC")
	if err != nil {
		return fmt.Errorf("doing exec: %w", err)
	}

	return err
}

// Shuffle image metadata in the category.
func (r ImageMetaRepository) Shuffle(
	ctx context.Context,
	category string,
	depth int,
) (err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	key := r.key(category)

	count, err := redis.Int(kv.Do("LLEN", key))
	switch {
	case err != nil:
		return fmt.Errorf("doing llen: %w", err)
	case count == 0, depth <= 0:
		return nil
	}

	for i := 0; i < depth; i++ {
		// nolint: gosec // It requires speed.
		elementIndex := rand.Intn(count)

		element, err := redis.Bytes(kv.Do(
			"LINDEX",
			key,
			elementIndex,
		))
		switch {
		case err == nil:
		case errors.Is(err, redis.ErrNil):
			// Count of elements changed, ignore.
			continue
		default:
			return fmt.Errorf("doing lindex: %w", err)
		}

		p := newPipeline(kv)
		p.Send("MULTI")
		p.Send(
			"LREM",
			key,
			1, // Count. Remove 1 element equal to element moving from head to tail.
			element,
		)
		p.Send(
			"RPUSH",
			key,
			element,
		)
		_, err = p.Do("EXEC")
		if err != nil {
			return fmt.Errorf("doing exec: %w", err)
		}
	}

	return nil
}

// Categories returns a list of known categories. This data is obtained
// from the keys of the images.
func (r ImageMetaRepository) Categories(
	ctx context.Context,
) (categories []string, err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	// It is assumed that there will be few categories, up to 100.
	// Warning: KEYS can block Redis for a significant amount of time
	// if the number of categories becomes large.
	categories, err = redis.Strings(kv.Do("KEYS", keyPrefixImageMeta+"*"))
	if err != nil {
		return nil, fmt.Errorf("doing keys: %w", err)
	}

	for i, c := range categories {
		categories[i] = strings.TrimPrefix(c, keyPrefixImageMeta)
	}

	return categories, nil
}
