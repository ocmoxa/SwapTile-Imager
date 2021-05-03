package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/validate"

	"github.com/go-playground/validator/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"github.com/rs/zerolog"
)

// Core is the application main API.
type Core struct {
	cfg           config.Core
	repoImageMeta repository.ImageMetaRepository
	fileStorage   storage.FileStorage
	validate      *validator.Validate

	healthCheckers []imager.Healther

	buffersPool *sync.Pool
}

// Essentials of the Core.
type Essentials struct {
	KVP *redis.Pool
	repository.ImageMetaRepository
	storage.FileStorage
	*validator.Validate
}

// NewCore creates the application main API.
func NewCore(es Essentials, cfg config.Core) *Core {
	return &Core{
		repoImageMeta: es.ImageMetaRepository,
		fileStorage:   es.FileStorage,
		validate:      es.Validate,
		cfg:           cfg,
		buffersPool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		healthCheckers: []imager.Healther{
			es.FileStorage,
			redisHealthChecker{KVP: es.KVP},
		},
	}
}

// UploadImage saves original image to storage.
func (c Core) UploadImage(
	ctx context.Context,
	im imager.ImageMeta,
	r io.Reader,
) (newIM imager.ImageMeta, err error) {
	if im.ID == "" {
		im.ID = uuid.NewString()
	}

	l := zerolog.Ctx(ctx)
	l.Debug().
		Str("image_id", im.ID).
		Interface("image", im).
		Msg("uploading image")

	if im.Size > c.cfg.MaxImageSize {
		err = imerrors.Error(
			fmt.Sprintf("max allowed image size is %d", c.cfg.MaxImageSize),
		)

		return im, imerrors.NewOversizeError(err)
	}

	if err = c.validate.Struct(&im); err != nil {
		err = fmt.Errorf("validating image meta: %w", err)

		return im, imerrors.NewUnprocessableEntity(err)
	}

	err = validate.ContentType(im.MIMEType, c.cfg.ImageContentTypes)
	if err != nil {
		err = fmt.Errorf("validating content-type: %w", err)

		return im, imerrors.NewMediaTypeError(err)
	}

	found, err := c.repoImageMeta.Exists(ctx, im.ID)
	switch {
	case err != nil:
		return im, fmt.Errorf("checking id: %w", err)
	case found:
		return im, imerrors.NewConflictError(imerrors.Error("id already found"))
	}

	r = io.LimitReader(r, c.cfg.MaxImageSize)

	if err = c.fileStorage.Upload(ctx, im, r); err != nil {
		return im, fmt.Errorf("uploading file: %w", err)
	}

	err = c.repoImageMeta.Insert(ctx, im)
	if err != nil {
		err = fmt.Errorf("inseting meta info: %w", err)

		if derr := c.fileStorage.Delete(ctx, im.ID); derr != nil {
			derr = fmt.Errorf("deleting file: %w", derr)
			err = imerrors.ErrorPair(err, derr)

			l.Err(err).Interface("image", im).
				Msg("failed to rollback file upload")
		}

		return im, fmt.Errorf("inserting meta: %w", err)
	}

	return im, nil
}

// DeleteImage from a database and a storage.
func (c Core) DeleteImage(ctx context.Context, id string) (err error) {
	if err = c.validate.Var(id, "image_id"); err != nil {
		err = fmt.Errorf("validating image_id: %w", err)

		return imerrors.NewUnprocessableEntity(err)
	}

	found, err := c.repoImageMeta.Exists(ctx, id)
	switch {
	case err != nil:
		return fmt.Errorf("exists: %w", err)
	case !found:
		return imerrors.NewNotFoundError(imerrors.Error("image not found"))
	}

	err = c.fileStorage.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting image from file storage: %w", err)
	}

	err = c.repoImageMeta.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting image from database: %w", err)
	}

	return nil
}

// GetImage downloads image, resizes it and returns its body.
func (c Core) GetImage(
	ctx context.Context,
	id string,
	size imager.ImageSize,
) (f storage.File, err error) {
	cacheID := id + ":" + string(size)

	l := zerolog.Ctx(ctx)
	l.Debug().
		Str("image_id", id).
		Str("cache_id", cacheID).
		Str("image_size", string(size)).
		Msg("getting image")

	err = validate.ImageSize(size, c.cfg.SupportedImageSizes)
	if err != nil {
		err = fmt.Errorf("size: %w", err)

		return storage.File{}, imerrors.NewUnprocessableEntity(err)
	}

	if err = c.validate.Var(id, "image_id"); err != nil {
		err = fmt.Errorf("validating image_id: %w", err)

		return storage.File{}, imerrors.NewUnprocessableEntity(err)
	}

	f, err = c.fileStorage.Get(ctx, id)
	if err != nil {
		return storage.File{}, fmt.Errorf("getting image from storage: %w", err)
	}

	buf := c.buffersPool.Get().(*bytes.Buffer)
	defer c.buffersPool.Put(buf)
	buf.Reset()

	_, err = buf.ReadFrom(f)
	if err != nil {
		return storage.File{}, fmt.Errorf("reading image: %w", err)
	}

	imgData := buf.Bytes()

	width, height := size.Size()
	img := bimg.NewImage(imgData)

	resizedImgData, err := img.ResizeAndCrop(width, height)
	if err != nil {
		return storage.File{}, fmt.Errorf("resizing image: %w", err)
	}

	f.ReadCloser = io.NopCloser(bytes.NewReader(resizedImgData))

	return f, nil
}

// ListCategories returns all known categories.
func (c Core) ListCategories(
	ctx context.Context,
) (data []string, err error) {
	return c.repoImageMeta.Categories(ctx)
}

// ShuffleImages swaps random images in the category up to depth times.
func (c Core) ShuffleImages(
	ctx context.Context,
	category string,
	depth int,
) (err error) {
	if err = c.validate.Var(category, "category"); err != nil {
		err = fmt.Errorf("validating category: %w", err)

		return imerrors.NewUnprocessableEntity(err)
	}

	if err = c.validate.Var(depth, "min=1,max=1000"); err != nil {
		err = fmt.Errorf("validating depth: %w", err)

		return imerrors.NewUnprocessableEntity(err)
	}

	return c.repoImageMeta.Shuffle(ctx, category, depth)
}

// ListImages returns a list of images by the category and pagination.
func (c Core) ListImages(
	ctx context.Context,
	category string,
	pagination repository.Pagination,
) (im []imager.RawImageMetaJSON, err error) {
	if err = c.validate.Var(category, "category"); err != nil {
		err = fmt.Errorf("validating category: %w", err)

		return nil, imerrors.NewUnprocessableEntity(err)
	}

	if err = c.validate.Struct(&pagination); err != nil {
		err = fmt.Errorf("validating pagination: %w", err)

		return nil, imerrors.NewUnprocessableEntity(err)
	}

	return c.repoImageMeta.List(ctx, category, pagination)
}

// Health checks health of Redis and S3.
func (c Core) Health(ctx context.Context) (err error) {
	for _, h := range c.healthCheckers {
		err = h.Health(ctx)
		if err != nil {
			return fmt.Errorf("health: %w", err)
		}
	}

	return nil
}
