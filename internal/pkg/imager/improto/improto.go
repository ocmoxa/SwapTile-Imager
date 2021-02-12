// Package improto contains generated proto messages and helpers to them.
package improto

import "github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"

// ToImageMetaProto converts imager.ImageMeta to ImageMeta.
func ToImageMetaProto(im imager.ImageMeta) *ImageMeta {
	return &ImageMeta{
		Id:        im.ID,
		Author:    im.Author,
		WebSource: im.WEBSource,
		MimeType:  im.MIMEType,
	}
}

// FromImageMeta converts *improto.ImageMeta to imager.ImageMeta. The
// argument im can be nil.
func FromImageMeta(im *ImageMeta) imager.ImageMeta {
	return imager.ImageMeta{
		ID:        im.GetId(),
		Author:    im.GetAuthor(),
		WEBSource: im.GetWebSource(),
		MIMEType:  im.GetMimeType(),
	}
}
