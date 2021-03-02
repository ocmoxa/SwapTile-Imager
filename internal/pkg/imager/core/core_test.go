// +build integration

package core_test

import (
	"bytes"
	"context"
	"errors"
	"image/color"
	"math"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/core"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository/imredis"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage/s3"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/validate"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

const contentType = "image/jpeg"
const imageSize imager.ImageSize = "128x128"

func newTestCore(tb testing.TB) (c *core.Core, close func(tb testing.TB)) {
	tb.Helper()

	cfg := test.LoadConfig(tb)
	cfg.ImageContentTypes = append(cfg.ImageContentTypes, contentType)
	cfg.SupportedImageSizes = append(cfg.SupportedImageSizes, imageSize)

	kvp := test.InitKVP(tb)

	s3, err := s3.NewS3Storage(cfg.S3)
	test.AssertErrNil(tb, err)

	c = core.NewCore(core.Essentials{
		ImageMetaRepository: imredis.NewImageMetaRepository(kvp),
		ImageIDRepository:   imredis.NewImageIDRepository(kvp),
		FileStorage:         s3,
		Validate:            validate.New(),
	}, cfg.Core)

	return c, func(tb testing.TB) {
		test.DisposeKVP(tb, kvp)
	}
}

func getTestImageBytes(t *testing.T) []byte {
	width, height := imageSize.Size()

	var imageData bytes.Buffer
	image := imaging.New(width*2, height*2, color.Black)
	err := imaging.Encode(&imageData, image, imaging.JPEG)
	test.AssertErrNil(t, err)

	return imageData.Bytes()
}

func TestUploadImage(t *testing.T) {
	testCases := []struct {
		Name      string
		Meta      func(*imager.ImageMeta)
		ErrTarget interface{}
	}{{
		Name:      "ok",
		Meta:      func(im *imager.ImageMeta) {},
		ErrTarget: nil,
	}, {
		Name:      "size_overflow",
		Meta:      func(im *imager.ImageMeta) { im.Size = math.MaxInt64 },
		ErrTarget: &imerrors.OversizeError{},
	}, {
		Name:      "size_zero",
		Meta:      func(im *imager.ImageMeta) { im.Size = 0 },
		ErrTarget: &imerrors.UserError{},
	}, {
		Name:      "size_negative",
		Meta:      func(im *imager.ImageMeta) { im.Size = -1 },
		ErrTarget: &imerrors.UserError{},
	}, {
		Name:      "no_author",
		Meta:      func(im *imager.ImageMeta) { im.Author = "" },
		ErrTarget: &imerrors.UserError{},
	}, {
		Name:      "no_websource",
		Meta:      func(im *imager.ImageMeta) { im.WEBSource = "" },
		ErrTarget: &imerrors.UserError{},
	}, {
		Name:      "no_mimetype",
		Meta:      func(im *imager.ImageMeta) { im.MIMEType = "" },
		ErrTarget: &imerrors.UserError{},
	}, {
		Name:      "invalid_mimetype",
		Meta:      func(im *imager.ImageMeta) { im.MIMEType = "invalid/mime" },
		ErrTarget: &imerrors.MediaTypeError{},
	}, {
		Name:      "no_category",
		Meta:      func(im *imager.ImageMeta) { im.Category = "" },
		ErrTarget: &imerrors.UserError{},
	}}

	c, close := newTestCore(t)
	defer close(t)

	imageBytes := getTestImageBytes(t)

	ctx := context.Background()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			im := imager.ImageMeta{
				ID:        "",
				Author:    "author",
				WEBSource: "localhost",
				MIMEType:  contentType,
				Size:      int64(len(imageBytes)),
				Category:  "test",
			}
			tc.Meta(&im)

			_, err := c.UploadImage(ctx, im, bytes.NewReader(imageBytes))
			switch {
			case tc.ErrTarget == nil:
				test.AssertErrNil(t, err)
			case !errors.As(err, tc.ErrTarget):
				t.Fatal(err)
			}
		})
	}
}

func TestGetImage(t *testing.T) {
	imageID := uuid.NewString()

	testCases := []struct {
		Name      string
		ID        string
		Size      imager.ImageSize
		ErrTarget interface{}
	}{{
		Name:      "ok",
		ID:        imageID,
		Size:      imageSize,
		ErrTarget: nil,
	}, {
		Name:      "not_found",
		ID:        uuid.NewString(),
		Size:      imageSize,
		ErrTarget: &imerrors.NotFoundError{},
	}, {
		Name:      "invalid_image_size",
		ID:        imageID,
		Size:      imager.ImageSize("1x0"),
		ErrTarget: &imerrors.UserError{},
	}, {
		Name:      "invalid_image_id",
		ID:        "invalid",
		Size:      imageSize,
		ErrTarget: &imerrors.UserError{},
	}}

	c, close := newTestCore(t)
	defer close(t)

	ctx := context.Background()

	imageBytes := getTestImageBytes(t)
	_, err := c.UploadImage(ctx, imager.ImageMeta{
		ID:        imageID,
		Author:    "author",
		WEBSource: "websource",
		MIMEType:  contentType,
		Size:      int64(len(imageBytes)),
		Category:  "test",
	}, bytes.NewReader(imageBytes))
	test.AssertErrNil(t, err)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			f, err := c.GetImage(ctx, tc.ID, tc.Size)

			if tc.ErrTarget != nil {
				if !errors.As(err, tc.ErrTarget) {
					t.Fatal(err)
				}

				return
			}

			test.AssertErrNil(t, err)

			image, err := imaging.Decode(f)
			test.AssertErrNil(t, err)

			gotSize := image.Bounds().Size()
			expWidth, expHeight := tc.Size.Size()

			switch {
			case f.ContentType != contentType:
				t.Fatal("exp", contentType, "got", f.ContentType)
			case gotSize.X != expWidth:
				t.Fatal("exp", expWidth, "got", gotSize.X)
			case gotSize.Y != expHeight:
				t.Fatal("exp", expHeight, "got", gotSize.Y)
			}

			err = f.Close()
			test.AssertErrNil(t, err)
		})
	}
}

func TestListCategories(t *testing.T) {
	const category = "test"

	c, close := newTestCore(t)
	defer close(t)

	ctx := context.Background()

	imageBytes := getTestImageBytes(t)
	_, err := c.UploadImage(ctx, imager.ImageMeta{
		ID:        "",
		Author:    "author",
		WEBSource: "websource",
		MIMEType:  contentType,
		Size:      int64(len(imageBytes)),
		Category:  category,
	}, bytes.NewReader(imageBytes))
	test.AssertErrNil(t, err)

	categories, err := c.ListCategories(ctx)
	test.AssertErrNil(t, err)

	var found bool
	for _, c := range categories {
		if c == category {
			found = true
		}
	}

	if !found {
		t.Fatalf("%s not found in %v", category, categories)
	}
}

func TestListImages(t *testing.T) {
	const category = "test"

	testCases := []struct {
		Name       string
		Category   string
		Pagination repository.Pagination
		ExpCount   int
		ErrTarget  interface{}
	}{{
		Name:       "ok",
		Category:   category,
		Pagination: repository.Pagination{Limit: 1, Offset: 0},
		ExpCount:   1,
		ErrTarget:  nil,
	}, {
		Name:       "invalid_category",
		Category:   "",
		Pagination: repository.Pagination{Limit: 1, Offset: 0},
		ExpCount:   0,
		ErrTarget:  &imerrors.UserError{},
	}, {
		Name:       "invalid_limit",
		Category:   category,
		Pagination: repository.Pagination{Limit: -1, Offset: 0},
		ExpCount:   0,
		ErrTarget:  &imerrors.UserError{},
	}, {
		Name:       "invalid_offset",
		Category:   category,
		Pagination: repository.Pagination{Limit: 1, Offset: -1},
		ExpCount:   0,
		ErrTarget:  &imerrors.UserError{},
	}}

	c, close := newTestCore(t)
	defer close(t)

	ctx := context.Background()

	imageBytes := getTestImageBytes(t)
	_, err := c.UploadImage(ctx, imager.ImageMeta{
		ID:        "",
		Author:    "author",
		WEBSource: "websource",
		MIMEType:  contentType,
		Size:      int64(len(imageBytes)),
		Category:  category,
	}, bytes.NewReader(imageBytes))
	test.AssertErrNil(t, err)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			images, err := c.ListImages(ctx, tc.Category, tc.Pagination)

			if tc.ErrTarget != nil {
				if !errors.As(err, tc.ErrTarget) {
					t.Fatal(err)
				}

				return
			}

			test.AssertErrNil(t, err)

			if len(images) != tc.ExpCount {
				t.Fatal("exp", tc.ExpCount, "got", len(images))
			}
		})
	}
}

func BenchmarkImage(b *testing.B) {
	// cpu: Intel(R) Core(TM) i5-7300HQ CPU @ 2.50GHz
	// BenchmarkImage
	// BenchmarkImage-4             288           4092604 ns/op           28785 B/op        386 allocs/op

	c, close := newTestCore(b)
	defer close(b)

	var imageData bytes.Buffer
	image := imaging.New(1, 1, color.Black)
	err := imaging.Encode(&imageData, image, imaging.JPEG)
	test.AssertErrNil(b, err)

	ctx := context.Background()
	im, err := c.UploadImage(ctx, imager.ImageMeta{
		ID:        "",
		Author:    "author",
		WEBSource: "websource",
		MIMEType:  contentType,
		Size:      int64(imageData.Len()),
		Category:  "test",
	}, &imageData)
	test.AssertErrNil(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		f, err := c.GetImage(ctx, im.ID, imageSize)
		test.AssertErrNil(b, err)

		err = f.Close()
		test.AssertErrNil(b, err)
	}
}
