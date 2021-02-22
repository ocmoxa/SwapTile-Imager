package validate_test

import (
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/validate"

	"github.com/google/uuid"
)

func TestNew(t *testing.T) {
	v := validate.New()

	err := v.Var(uuid.NewString(), "image_id")
	test.AssertErrNil(t, err)

	err = v.Var("a", "category")
	test.AssertErrNil(t, err)
}

func TestValidateImageSize(t *testing.T) {
	supported := []imager.ImageSize{"1024x768", "640x360"}

	err := validate.ValidateImageSize(imager.ImageSize("10x10"), supported)
	if err == nil {
		t.Fatal(err)
	}

	err = validate.ValidateImageSize(imager.ImageSize("1024x768"), supported)
	test.AssertErrNil(t, err)
}

func TestValidateContentType(t *testing.T) {
	supported := []string{"image/jpeg", "image/png"}

	err := validate.ValidateContentType("image/gif", supported)
	if err == nil {
		t.Fatal(err)
	}

	err = validate.ValidateContentType("image/jpeg", supported)
	test.AssertErrNil(t, err)
}
