package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/validate"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

const (
	outputImageFormat      = imaging.JPEG
	outputImageContentType = "image/jpeg"
)

type Core struct {
	cfg           config.Core
	repoImageMeta repository.ImageMetaRepository
	fileStorage   storage.FileStorage
	fileCache     storage.FileCache
	validate      *validator.Validate
}

type Essentials struct {
	repository.ImageMetaRepository
	storage.FileStorage
	storage.FileCache
	*validator.Validate
}

func NewCore(es Essentials, cfg config.Core) *Core {
	return &Core{
		repoImageMeta: es.ImageMetaRepository,
		fileStorage:   es.FileStorage,
		fileCache:     es.FileCache,
		validate:      es.Validate,
		cfg:           cfg,
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

	if c.fileCache != nil {
		f, err = c.fileCache.Get(cacheID)
		switch {
		case err == nil:
			return f, nil
		case errors.As(err, &imerrors.NotFoundError{}):
			break
		default:
			return storage.File{}, fmt.Errorf("getting file from cache: %w", err)
		}
	}

	f, err = c.fileStorage.Get(ctx, id)
	if err != nil {
		return storage.File{}, fmt.Errorf("getting image from storage: %w", err)
	}

	img, err := imaging.Decode(f)
	if err != nil {
		return storage.File{}, fmt.Errorf("decoding image: %w", err)
	}

	width, height := size.Size()
	img = imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

	var buf bytes.Buffer
	err = imaging.Encode(&buf, img, outputImageFormat)
	if err != nil {
		return storage.File{}, fmt.Errorf("encoding image: %w", err)
	}

	f.ContentType = outputImageContentType
	f.ReadCloser = ioutil.NopCloser(&buf)

	if c.fileCache != nil {
		if err = c.fileCache.Put(cacheID, f); err != nil {
			return storage.File{}, fmt.Errorf("putting file to cache: %w", err)
		}
	}

	return
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
