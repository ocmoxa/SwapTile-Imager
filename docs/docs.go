package docs

import "embed"

// nolint: gochecknoglobals // Go embed.
//go:embed swagger.yml ui
var embededFiles embed.FS

// Swagger returns global file system with swagger files.
// It includes the file ./swagger.yml and the folter ./ui.
func Swagger() embed.FS {
	return embededFiles
}
