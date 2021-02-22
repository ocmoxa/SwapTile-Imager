// Package repository holds models and interfaces of the database
// repositories.
package repository

import (
	"context"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
)

// ImageMetaRepository stores known image meta data.
type ImageMetaRepository interface {
	// List returns a list of image ids for the given category.
	List(ctx context.Context, category string, pagination Pagination) (im []IndexedImageMeta, err error)
	// Insert saves new image id to the category. The uniquness of the
	// id is not checked.
	Insert(ctx context.Context, im imager.ImageMeta) (err error)
	// Delete deletes image by id from the category.
	Delete(ctx context.Context, category string, index int) (err error)
	// Categories returns a list of known categories.
	Categories(ctx context.Context) (categories []string, err error)
}

// ImageIDRepository saves image ids to ensure uniqueness.
type ImageIDRepository interface {
	// Set saves id and returns if it is already used.
	Set(ctx context.Context, id string) (ok bool, err error)
	// Delete id.
	Delete(ctx context.Context, id string) (err error)
}

// Pagination holds query limits.
type Pagination struct {
	// Limit of rows in the result.
	Limit int `validate:"gte=1"`
	// Offset of rows in the result.
	Offset int `validate:"gte=0"`
}

// IndexedImageMeta adds index to image meta information.
type IndexedImageMeta struct {
	Index int `json:"index"`
	imager.ImageMeta
}
