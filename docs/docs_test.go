package docs_test

import (
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/docs"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"
)

func TestSwagger(t *testing.T) {
	fsys := docs.Swagger()

	entities, err := fsys.ReadDir(".")
	test.AssertErrNil(t, err)

	if len(entities) == 0 {
		t.Fatal(entities)
	}
}
