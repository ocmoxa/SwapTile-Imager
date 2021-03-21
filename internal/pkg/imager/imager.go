// Package imager contains domain application models and interfaces.
package imager

import (
	"context"
	"encoding/json"
	"fmt"
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
	Category string `json:"category" validate:"required,category,ne=all"`
	// Size of file in bytes.
	Size int64 `json:"-" validate:"required,gt=0"`
}

// RawJSON converts ImageMeta to raw json representation.
func (im ImageMeta) RawJSON() (RawImageMetaJSON, error) {
	return json.Marshal(&im)
}

// RawImageMetaJSON holds JSON meta information about image meta. It is
// not decoded.
type RawImageMetaJSON []byte

// MarshalJSON implements JSON marshaller.
func (rawIM *RawImageMetaJSON) MarshalJSON() ([]byte, error) {
	return []byte(*rawIM), nil
}

// UnmarshalJSON implements JSON unmarshaller.
func (rawIM *RawImageMetaJSON) UnmarshalJSON(data []byte) error {
	*rawIM = make(RawImageMetaJSON, len(data))
	copy(*rawIM, data)

	return nil
}

// ImageMeta decodes raw image meta details into ImageMeta.
func (rawIM RawImageMetaJSON) ImageMeta() (im ImageMeta, err error) {
	if err = json.Unmarshal(rawIM, &im); err != nil {
		return ImageMeta{}, fmt.Errorf("decoding json meta: %w", err)
	}

	return im, nil
}

// String implements Stringer interface.
func (rawIM RawImageMetaJSON) String() string {
	return string(rawIM)
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

// Healther checks component status.
type Healther interface {
	// Healther checks component status.
	Health(ctx context.Context) (err error)
}
