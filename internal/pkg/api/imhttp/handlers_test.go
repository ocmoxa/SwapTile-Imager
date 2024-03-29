// +build integration

package imhttp_test

import (
	"bytes"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/api/imhttp"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/core"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository/imredis"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage/s3"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/validate"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
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
			image := image.NewNRGBA(image.Rect(0, 0, 10, 10))
			err := jpeg.Encode(&imageData, image, &jpeg.Options{Quality: 10})
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
				"/api/v1/images/"+imageID+"/"+string(size),
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
	}, {
		Request: func() *http.Request {
			const data = `{"category":"test","depth":1}`
			return httptest.NewRequest(
				http.MethodPost,
				"/internal/api/v1/images/shuffle",
				strings.NewReader(data),
			)
		},
		ExpStatus: http.StatusOK,
	}, {
		Request: func() *http.Request {
			return httptest.NewRequest(
				http.MethodDelete,
				"/internal/api/v1/images/"+imageID,
				nil,
			)
		},
		ExpStatus: http.StatusOK,
	}, {
		Request: func() *http.Request {
			return httptest.NewRequest(http.MethodGet, "/health", nil)
		},
		ExpStatus: http.StatusOK,
	}, {
		Request: func() *http.Request {
			return httptest.NewRequest(http.MethodGet, "/not-found", nil)
		},
		ExpStatus: http.StatusNotFound,
	}}

	cfg := test.LoadConfig(t)
	cfg.Core.ImageContentTypes = append(cfg.Core.ImageContentTypes, contentType)
	cfg.Core.SupportedImageSizes = append(cfg.Core.SupportedImageSizes, size)
	cfg.Server.ExposeErrors = true

	kvp := test.InitKVP(t)
	defer test.DisposeKVP(t, kvp)

	s3, err := s3.NewStorage(cfg.S3)
	test.AssertErrNil(t, err)

	c := core.NewCore(core.Essentials{
		KVP:                 kvp,
		ImageMetaRepository: imredis.NewImageMetaRepository(kvp),
		FileStorage:         s3,
		Validate:            validate.New(),
	}, cfg.Core)

	s, err := imhttp.NewServer(
		imhttp.Essentials{
			Logger:       zerolog.New(os.Stdout),
			Core:         c,
			PromRegistry: prometheus.NewRegistry(),
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
