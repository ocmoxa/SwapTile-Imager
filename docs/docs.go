package docs

import "embed"

//go:embed swagger.yml ui
var embededFiles embed.FS

func Swagger() embed.FS {
	return embededFiles
}
