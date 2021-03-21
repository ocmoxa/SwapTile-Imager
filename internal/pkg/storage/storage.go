package storage

import (
	"context"
	"io"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
)

// FileStorage stores images.
type FileStorage interface {
	// Get images from the storage.
	Get(ctx context.Context, id string) (f File, err error)
	// Upload image to the storage.
	Upload(ctx context.Context, im imager.ImageMeta, r io.Reader) (err error)
	// Delete image in the storage.
	Delete(ctx context.Context, id string) (err error)

	imager.Healther
}

// File is an image from the storage. It stores content type. The file
// should be closed after usage.
type File struct {
	io.ReadCloser

	ContentType string
}
