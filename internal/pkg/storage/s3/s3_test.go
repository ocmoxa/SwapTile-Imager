package s3_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage/s3"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"

	"github.com/google/uuid"
)

func TestS3Storage(t *testing.T) {
	storage, err := s3.NewS3Storage(test.LoadConfig(t).S3)
	test.AssertErrNil(t, err)

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
	if !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}

	if !bytes.Equal(data, gotData) {
		t.Fatal("exp", data, "got", gotData)
	}

	err = storage.Delete(ctx, im.ID)
	test.AssertErrNil(t, err)

	_, err = storage.Get(ctx, im.ID)
	if !errors.As(err, &imerrors.NotFoundError{}) {
		t.Fatal(err)
	}
}
