package imager_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"
)

func TestImageSize(t *testing.T) {
	testCases := []struct {
		ImageSize imager.ImageSize
		ExpWidth  int
		ExpHeight int
	}{{
		ImageSize: "10x15",
		ExpWidth:  10,
		ExpHeight: 15,
	}, {
		ImageSize: "10x-1",
		ExpWidth:  0,
		ExpHeight: 0,
	}, {
		ImageSize: "10x0",
		ExpWidth:  0,
		ExpHeight: 0,
	}, {
		ImageSize: "10X10",
		ExpWidth:  0,
		ExpHeight: 0,
	}, {
		ImageSize: "10xb",
		ExpWidth:  0,
		ExpHeight: 0,
	}, {
		ImageSize: "4294967295x4294967295",
		ExpWidth:  4294967295,
		ExpHeight: 4294967295,
	}}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.ImageSize), func(t *testing.T) {
			width, height := tc.ImageSize.Size()

			switch {
			case height != tc.ExpHeight:
				t.Fatal("exp", tc.ExpHeight, "got", height)
			case width != tc.ExpWidth:
				t.Fatal("exp", tc.ExpWidth, "got", width)
			}
		})
	}
}

func TestRawImageMetaJSON(t *testing.T) {
	im := imager.ImageMeta{
		ID:        "test",
		Author:    "test_author",
		WEBSource: "test_websource",
		MIMEType:  "image/jpeg",
		Category:  "test_category",
		Size:      1,
	}

	rawIM, err := im.RawJSON()
	test.AssertErrNil(t, err)

	gotIm, err := rawIM.ImageMeta()
	test.AssertErrNil(t, err)
	if gotIm.ID != im.ID {
		t.Fatal("exp", im.ID, "got", gotIm.ID)
	}

	if !strings.Contains(rawIM.String(), im.ID) {
		t.Fatal(rawIM.String())
	}

	gotRawIM, err := json.Marshal(&rawIM)
	test.AssertErrNil(t, err)
	if !bytes.Contains(gotRawIM, []byte(im.ID)) {
		t.Fatal(string(gotRawIM))
	}

	err = json.Unmarshal(gotRawIM, &rawIM)
	test.AssertErrNil(t, err)
}
