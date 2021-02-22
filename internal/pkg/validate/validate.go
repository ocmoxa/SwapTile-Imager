package validate

import (
	"fmt"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"

	"github.com/go-playground/validator/v10"
)

// New creates validator.Validate and registers custom parsers.
func New() *validator.Validate {
	v := validator.New()

	v.RegisterAlias("image_id", "uuid")
	v.RegisterAlias("category", "alphanum")

	return v
}

// ValidateImageSize checks that image size is supported.
func ValidateImageSize(
	size imager.ImageSize,
	supportedSizes []imager.ImageSize,
) (err error) {
	for _, s := range supportedSizes {
		if s == size {
			return nil
		}
	}

	return fmt.Errorf("expected one of %v", supportedSizes)
}

// ValidateImageSize checks that content-type is supported.
func ValidateContentType(
	contentType string,
	supportedContentTypes []string,
) (err error) {
	for _, ct := range supportedContentTypes {
		if ct == contentType {
			return nil
		}
	}

	return fmt.Errorf("expected one of %v", supportedContentTypes)
}
