package imredis

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"

	"github.com/gomodule/redigo/redis"
)

// CategoryNameAll is a special name of category that contains all images.
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

func (r ImageMetaRepository) keyImageID(category string) string {
	return keyPrefixImageID + category
}

// List of image meta.
func (r ImageMetaRepository) List(
	ctx context.Context,
	category string,
	pagination repository.Pagination,
) (im []imager.RawImageMetaJSON, err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	imageIDs, err := redis.Strings(kv.Do(
		"LRANGE",
		r.keyImageID(category),
		pagination.Offset,                    // Start, inclusive.
		pagination.Offset+pagination.Limit-1, // Stop, inclusive.
	))
	switch {
	case err != nil:
		return nil, fmt.Errorf("doing lrange: %w", err)
	case len(imageIDs) == 0:
		return nil, nil
	}

	metaArgs := make([]interface{}, 0, len(imageIDs)+1)
	metaArgs = append(metaArgs, keyImageMeta)
	for _, imgID := range imageIDs {
		metaArgs = append(metaArgs, imgID)
	}

	imBytes, err := redis.ByteSlices(kv.Do(
		"HMGET",
		metaArgs...,
	))
	if err != nil {
		return nil, fmt.Errorf("doing hmget: %w", err)
	}

	im = make([]imager.RawImageMetaJSON, len(imBytes))
	for i, v := range imBytes {
		im[i] = v
	}

	return im, nil
}

// Exists checks that image meta found.
func (r ImageMetaRepository) Exists(
	ctx context.Context,
	imageID string,
) (found bool, err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	found, err = redis.Bool(kv.Do(
		"HEXISTS",
		keyImageMeta,
		imageID,
	))
	if err != nil {
		return false, fmt.Errorf("doing hexists: %w", err)
	}

	return found, nil
}

// Insert an image metadata.
func (r ImageMetaRepository) Insert(
	ctx context.Context,
	im imager.ImageMeta,
) (err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	imData, err := im.RawJSON()
	if err != nil {
		return fmt.Errorf("encoding image meta: %w", err)
	}

	p := newPipeline(kv)
	p.Send("MULTI")
	p.Send(
		"RPUSH",
		r.keyImageID(im.Category),
		im.ID, // Element.
	)
	p.Send(
		"RPUSH",
		r.keyImageID(CategoryNameAll),
		im.ID, // Element.
	)
	p.Send(
		"HSET",
		keyImageMeta,
		im.ID,
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
	imageID string,
) (err error) {
	kv := r.kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	rawIM, err := redis.Bytes(kv.Do(
		"HGET",
		keyImageMeta,
		imageID,
	))
	if err != nil {
		return fmt.Errorf("doing hget: %w", err)
	}

	im, err := imager.RawImageMetaJSON(rawIM).ImageMeta()
	if err != nil {
		return fmt.Errorf("decoding image meta: %w", err)
	}

	p := newPipeline(kv)
	p.Send("MULTI")
	p.Send(
		"LREM",
		r.keyImageID(im.Category),
		0, // Count. Remove all elements equal to element.
		imageID,
	)
	p.Send(
		"LREM",
		r.keyImageID(CategoryNameAll),
		0, // Count. Remove all elements equal to element.
		imageID,
	)
	p.Send(
		"HDEL",
		keyImageMeta,
		imageID,
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

	key := r.keyImageID(category)

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

		imageID, err := redis.Bytes(kv.Do(
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
			imageID,
		)
		p.Send(
			"RPUSH",
			key,
			imageID,
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
	categories, err = redis.Strings(kv.Do("KEYS", r.keyImageID("*")))
	if err != nil {
		return nil, fmt.Errorf("doing keys: %w", err)
	}

	for i, c := range categories {
		categories[i] = strings.TrimPrefix(c, keyPrefixImageID)
	}

	return categories, nil
}
