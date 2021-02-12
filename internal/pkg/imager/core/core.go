package core

import (
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"
)

type Core struct {
	repoImageMeta repository.ImageMetaRepository
}

type Essentials struct {
	repository.ImageMetaRepository
}

func NewCore(es Essentials) *Core {
	return &Core{
		repoImageMeta: es.ImageMetaRepository,
	}
}

func (c Core) UploadImage(meta imager.ImageMeta, data []byte) (err error) {
	return nil
}
