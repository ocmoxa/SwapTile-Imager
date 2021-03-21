// Package repository holds models and interfaces of the database
// repositories.
package repository

import (
	"context"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
)

// ImageMetaRepository stores known image meta data.
type ImageMetaRepository interface {
	// List returns a list of image details for the given category. The
	// result is not encoded.
	List(ctx context.Context, category string, pagination Pagination) (im []imager.RawImageMetaJSON, err error)
	// Exists checks that image meta found.
	Exists(ctx context.Context, imageID string) (found bool, err error)
	// Insert saves new image id to the category. The uniquness of the
	// id is not checked.
	Insert(ctx context.Context, im imager.ImageMeta) (err error)
	// Delete deletes image by id from the category.
	Delete(ctx context.Context, imageID string) (err error)
	// Categories returns a list of known categories.
	Categories(ctx context.Context) (categories []string, err error)
	// Shuffle swaps random images in the category. The depth should be
	// positive.
	Shuffle(ctx context.Context, category string, depth int) (err error)
}

// Pagination holds query limits.
type Pagination struct {
	// Limit of rows in the result.
	Limit int `validate:"gte=1,lte=200"`
	// Offset of rows in the result.
	Offset int `validate:"gte=0"`
}
