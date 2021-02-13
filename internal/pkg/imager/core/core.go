package core

import (
	"context"
	"fmt"
	"io"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage"

	"github.com/rs/zerolog"
)

type Core struct {
	repoImageMeta repository.ImageMetaRepository
	fileStorage   storage.FileStorage
}

type Essentials struct {
	repository.ImageMetaRepository
	storage.FileStorage
}

func NewCore(es Essentials) *Core {
	return &Core{
		repoImageMeta: es.ImageMetaRepository,
		fileStorage:   es.FileStorage,
	}
}

func (c Core) UploadImage(
	ctx context.Context,
	im imager.ImageMeta,
	r io.Reader,
) (err error) {
	l := zerolog.Ctx(ctx)
	l.Debug().
		Str("image_id", im.ID).
		Interface("image", im).
		Msg("uploading image")

	// TODO: validate. Check size.

	if err = c.fileStorage.Upload(ctx, im, r); err != nil {
		return fmt.Errorf("uploading file: %w", err)
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

		return err
	}

	return nil
}

func (c Core) GetImage(
	ctx context.Context,
	id string,
	size imager.ImageSize,
) (r storage.File, err error) {
	// TODO: validate. Check size.

	l := zerolog.Ctx(ctx)
	l.Debug().
		Str("image_id", id).
		Str("image_size", string(size)).
		Msg("getting image")

	return c.fileStorage.Get(ctx, id)
}

func (c Core) Categories(
	ctx context.Context,
) (data []string, err error) {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("getting categories")

	// TODO: cache it.

	return c.repoImageMeta.Categories(ctx)
}

func (c Core) ListImages(
	ctx context.Context,
	category string,
	pagination repository.Pagination,
) (im []repository.IndexedImageMeta, err error) {
	// TODO: validate.

	return c.repoImageMeta.List(ctx, category, pagination)
}
