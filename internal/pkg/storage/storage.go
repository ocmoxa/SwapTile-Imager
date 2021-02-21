package storage

import (
	"context"
	"io"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
)

type FileStorage interface {
	Get(ctx context.Context, id string) (f File, err error)
	Upload(ctx context.Context, im imager.ImageMeta, r io.Reader) (err error)
	Delete(ctx context.Context, id string) (err error)
}

type File struct {
	io.ReadCloser

	ContentType string
}

// FileCache caches images permanently.
type FileCache interface {
	// Put file to cache.
	Put(id string, f File) (err error)
	// Get file from cache. It returns NotFoundError if file not found.
	Get(id string) (f File, err error)
	// ClearExpired removes old unused entries from the cache.
	ClearExpired()
}
