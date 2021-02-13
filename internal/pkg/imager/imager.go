// Imager contains domain application models and interfaces.
package imager

import (
	"strconv"
	"strings"
)

// ImageMeta contains the main information about an image.
type ImageMeta struct {
	// ID of the image.
	ID string `json:"id"`
	// Author of the image.
	Author string `json:"author"`
	// WEBSource is a url to the website with original image.
	WEBSource string `json:"source"`
	// MIMEType is an image media type.
	MIMEType string `json:"mimetype"`
	// Size of file in bytes.
	Size int64 `json:"size"`
	// Category of the image.
	Category string `json:"category"`
}

// ImageSize is a size defined as a string: WIDTHxHEIGHT
type ImageSize string

// Width parses image size and returns width and height. If image size
// is invalid both values will be zero.
func (is ImageSize) Size() (width int, height int) {
	sizeTokens := strings.Split(string(is), "x")
	if len(sizeTokens) != 2 {
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
