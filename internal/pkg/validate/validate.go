package validate

import (
	"fmt"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"

	"github.com/go-playground/validator/v10"
)

// New creates validator.Validate and registers custom parsers.
func New() (*validator.Validate, error) {
	v := validator.New()

	err := v.RegisterValidation("image_id", func(fl validator.FieldLevel) bool {
		err := v.Var(fl.Field().Interface(), "uuid")
		return err == nil
	})
	if err != nil {
		return nil, fmt.Errorf("registering image_id: %w", err)
	}

	err = v.RegisterValidation("category", func(fl validator.FieldLevel) bool {
		err := v.Var(fl.Field().Interface(), "alphanum")
		return err == nil
	})
	if err != nil {
		return nil, fmt.Errorf("registering category: %w", err)
	}

	return v, nil
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
