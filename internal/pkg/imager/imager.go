// Package imager contains domain application models and interfaces.
package imager

import (
	"strconv"
	"strings"
)

// ImageMeta contains the main information about an image.
type ImageMeta struct {
	// ID of the image.
	ID string `json:"id" validate:"required,image_id"`
	// Author of the image.
	Author string `json:"author" validate:"required"`
	// WEBSource is a url to the website with original image.
	WEBSource string `json:"source" validate:"required"`
	// MIMEType is an image media type.
	MIMEType string `json:"mimetype" validate:"required"`
	// Category of the image.
	Category string `json:"category" validate:"required,category"`
	// Size of file in bytes.
	Size int64 `json:"-" validate:"required,gt=0"`
}

// ImageSize is a size defined as a string: WIDTHxHEIGHT.
type ImageSize string

// Size parses image size and returns width and height. If image size
// is invalid both values will be zero.
func (is ImageSize) Size() (width int, height int) {
	const tokensWidthHeightCount = 2

	sizeTokens := strings.Split(string(is), "x")
	if len(sizeTokens) != tokensWidthHeightCount {
		return 0, 0
	}

	var err error
	width, err = strconv.Atoi(sizeTokens[0])
	switch {
	case err != nil:
		fallthrough
	case width <= 0:
		return 0, 0
	}

	height, err = strconv.Atoi(sizeTokens[1])
	switch {
	case err != nil:
		fallthrough
	case height <= 0:
		return 0, 0
	}

	return width, height
}
