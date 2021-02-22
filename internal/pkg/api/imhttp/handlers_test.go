// +build integration

package imhttp_test

import (
	"bytes"
	"image/color"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/api/imhttp"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/core"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository/imredis"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage/s3"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/validate"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func TestServer(t *testing.T) {
	imageID := uuid.NewString()
	size := imager.ImageSize("128x128")

	const category = "test"
	// This Content-Type will be set for form file after CreateFormFile.
	const contentType = "application/octet-stream"

	testCases := []struct {
		Request   func() *http.Request
		ExpStatus int
	}{{
		Request: func() *http.Request {
			var imageData bytes.Buffer
			err := imaging.Encode(
				&imageData,
				imaging.New(10, 10, color.Black),
				imaging.JPEG,
			)
			test.AssertErrNil(t, err)

			var b bytes.Buffer
			w := multipart.NewWriter(&b)

			err = w.WriteField("author", "author")
			test.AssertErrNil(t, err)

			err = w.WriteField("websource", "localhost")
			test.AssertErrNil(t, err)

			err = w.WriteField("category", category)
			test.AssertErrNil(t, err)

			err = w.WriteField("id", imageID)
			test.AssertErrNil(t, err)

			fw, err := w.CreateFormFile("image", "file.jpg")
			test.AssertErrNil(t, err)

			_, err = fw.Write(imageData.Bytes())
			test.AssertErrNil(t, err)

			err = w.Close()
			test.AssertErrNil(t, err)

			r := httptest.NewRequest(
				http.MethodPut,
				"/internal/api/v1/images",
				&b,
			)

			r.Header.Set("Content-Type", w.FormDataContentType())
			return r
		},
		ExpStatus: http.StatusOK,
	}, {
		Request: func() *http.Request {
			return httptest.NewRequest(
				http.MethodGet,
				"/api/v1/categories",
				nil,
			)
		},
		ExpStatus: http.StatusOK,
	}, {
		Request: func() *http.Request {
			return httptest.NewRequest(
				http.MethodGet,
				"/api/v1/images?limit=1&offset=0&category="+category,
				nil,
			)
		},
		ExpStatus: http.StatusOK,
	}, {
		Request: func() *http.Request {
			return httptest.NewRequest(
				http.MethodGet,
				"/api/v1/images/"+imageID+"?size="+string(size),
				nil,
			)
		},
		ExpStatus: http.StatusOK,
	}, {
		Request: func() *http.Request {
			return httptest.NewRequest(
				http.MethodGet,
				"/api/v1/images/"+uuid.NewString()+"?size="+string(size),
				nil,
			)
		},
		ExpStatus: http.StatusNotFound,
	}}

	cfg := test.LoadConfig(t)
	cfg.Core.ImageContentTypes = append(cfg.Core.ImageContentTypes, contentType)
	cfg.Core.SupportedImageSizes = append(cfg.Core.SupportedImageSizes, size)
	cfg.Server.ExposeErrors = true

	kvp := test.InitKVP(t)
	defer test.DisposeKVP(t, kvp)

	s3, err := s3.NewS3Storage(cfg.S3)
	test.AssertErrNil(t, err)

	c := core.NewCore(core.Essentials{
		ImageMetaRepository: imredis.NewImageMetaRepository(kvp),
		ImageIDRepository:   imredis.NewImageIDRepository(kvp),
		FileStorage:         s3,
		FileCache:           nil,
		Validate:            validate.New(),
	}, cfg.Core)

	s, err := imhttp.NewServer(
		imhttp.Essentials{
			Logger: zerolog.New(os.Stdout),
			Core:   c,
		},
		cfg.Server,
	)
	test.AssertErrNil(t, err)

	for _, tc := range testCases {
		tc := tc
		r := tc.Request()
		t.Run(r.Method+" "+r.URL.String(), func(t *testing.T) {
			w := httptest.NewRecorder()

			s.Handler.ServeHTTP(w, r)

			t.Log(w.Body.String())

			if w.Code != tc.ExpStatus {
				t.Fatal("exp", tc.ExpStatus, "got", w.Code)
			}
		})
	}
}
