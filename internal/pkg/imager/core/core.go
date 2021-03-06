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
	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"github.com/rs/zerolog"
)

type Core struct {
	cfg           config.Core
	repoImageMeta repository.ImageMetaRepository
	repoImageID   repository.ImageIDRepository
	fileStorage   storage.FileStorage
	validate      *validator.Validate

	buffersPool *sync.Pool
}

type Essentials struct {
	repository.ImageMetaRepository
	repository.ImageIDRepository
	storage.FileStorage
	*validator.Validate
}

func NewCore(es Essentials, cfg config.Core) *Core {
	return &Core{
		repoImageMeta: es.ImageMetaRepository,
		repoImageID:   es.ImageIDRepository,
		fileStorage:   es.FileStorage,
		validate:      es.Validate,
		cfg:           cfg,
		buffersPool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

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
		err = fmt.Errorf("max allowed image size is %d", c.cfg.MaxImageSize)

		return im, imerrors.NewOversizeError(err)
	}

	if err = c.validate.Struct(&im); err != nil {
		err = fmt.Errorf("validating image meta: %w", err)

		return im, imerrors.NewUserError(err)
	}

	err = validate.ValidateContentType(im.MIMEType, c.cfg.ImageContentTypes)
	if err != nil {
		err = fmt.Errorf("validating content-type: %w", err)

		return im, imerrors.NewMediaTypeError(err)
	}

	ok, err := c.repoImageID.Set(ctx, im.ID)
	switch {
	case err != nil:
		return im, fmt.Errorf("saving id: %w", err)
	case !ok:
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

		return im, err
	}

	return im, nil
}

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

	err = validate.ValidateImageSize(size, c.cfg.SupportedImageSizes)
	if err != nil {
		err = fmt.Errorf("size: %w", err)

		return storage.File{}, imerrors.NewUserError(err)
	}

	if err = c.validate.Var(id, "image_id"); err != nil {
		err = fmt.Errorf("validating image_id: %w", err)

		return storage.File{}, imerrors.NewUserError(err)
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

func (c Core) ListCategories(
	ctx context.Context,
) (data []string, err error) {
	return c.repoImageMeta.Categories(ctx)
}

func (c Core) ListImages(
	ctx context.Context,
	category string,
	pagination repository.Pagination,
) (im []repository.IndexedImageMeta, err error) {
	if err = c.validate.Var(category, "category"); err != nil {
		err = fmt.Errorf("validating category: %w", err)

		return nil, imerrors.NewUserError(err)
	}

	if err = c.validate.Struct(&pagination); err != nil {
		err = fmt.Errorf("validating pagination: %w", err)

		return nil, imerrors.NewUserError(err)
	}

	return c.repoImageMeta.List(ctx, category, pagination)
}
