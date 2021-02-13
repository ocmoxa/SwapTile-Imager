package imager_test

import (
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
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
