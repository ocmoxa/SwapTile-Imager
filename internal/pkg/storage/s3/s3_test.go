// +build integration

package s3_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage/s3"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"

	"github.com/google/uuid"
)

func TestStorage(t *testing.T) {
	s, err := s3.NewStorage(test.LoadConfig(t).S3)
	test.AssertErrNil(t, err)

	var storage storage.FileStorage = s

	ctx := context.Background()
	data := []byte("hello world")
	im := imager.ImageMeta{
		ID:        uuid.New().String(),
		Author:    "test_author",
		WEBSource: "test_websource",
		MIMEType:  "text/plain",
		Size:      int64(len(data)),
	}

	err = storage.Upload(ctx, im, bytes.NewReader(data))
	test.AssertErrNil(t, err)

	r, err := storage.Get(ctx, im.ID)
	test.AssertErrNil(t, err)

	gotData := make([]byte, len(data))
	_, err = r.Read(gotData)
	switch {
	case !errors.Is(err, io.EOF):
		t.Fatal(err)
	case !bytes.Equal(data, gotData):
		t.Fatal("exp", data, "got", gotData)
	case r.ContentType != im.MIMEType:
		t.Fatal("exp", im.MIMEType, "got", r.ContentType)
	}

	err = storage.Delete(ctx, im.ID)
	test.AssertErrNil(t, err)

	_, err = storage.Get(ctx, im.ID)
	if !errors.As(err, &imerrors.NotFoundError{}) {
		t.Fatal(err)
	}
}
