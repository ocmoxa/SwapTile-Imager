package improto_test

import (
	"reflect"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/improto"
)

func TestToImageMeta(t *testing.T) {
	exp := imager.ImageMeta{
		ID:        "test_id",
		Author:    "test_author",
		WEBSource: "test_websource",
		MIMEType:  "test_mimetype",
	}
	got := improto.FromImageMeta(&improto.ImageMeta{
		Id:        "test_id",
		Author:    "test_author",
		WebSource: "test_websource",
		MimeType:  "test_mimetype",
	})
	if !reflect.DeepEqual(exp, got) {
		t.Fatal("exp", exp, "got", got)
	}
}

func TestFromImageMeta(t *testing.T) {
	exp := &improto.ImageMeta{
		Id:        "test_id",
		Author:    "test_author",
		WebSource: "test_websource",
		MimeType:  "test_mimetype",
	}
	got := improto.ToImageMetaProto(imager.ImageMeta{
		ID:        "test_id",
		Author:    "test_author",
		WEBSource: "test_websource",
		MIMEType:  "test_mimetype",
	})
	if !reflect.DeepEqual(exp, got) {
		t.Fatal("exp", exp, "got", got)
	}
}
