// Imager contains domain application models and interfaces.
package imager

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
}
