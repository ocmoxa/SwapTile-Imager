package validate_test

import (
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/validate"
)

func TestNew(t *testing.T) {
	v := validate.New()

	err := v.Var("0m-cPyH_WDU", "image_id")
	test.AssertErrNil(t, err)

	err = v.Var("a", "category")
	test.AssertErrNil(t, err)
}

func TestValidateImageSize(t *testing.T) {
	supported := []imager.ImageSize{"1024x768", "640x360"}

	err := validate.ImageSize(imager.ImageSize("10x10"), supported)
	if err == nil {
		t.Fatal(err)
	}

	err = validate.ImageSize(imager.ImageSize("1024x768"), supported)
	test.AssertErrNil(t, err)
}

func TestValidateContentType(t *testing.T) {
	supported := []string{"image/jpeg", "image/png"}

	err := validate.ContentType("image/gif", supported)
	if err == nil {
		t.Fatal(err)
	}

	err = validate.ContentType("image/jpeg", supported)
	test.AssertErrNil(t, err)
}
