package filecache_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage/filecache"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"
)

func TestFileCache(t *testing.T) {
	const cacheID = "test_id"
	const contentType = "plain/text"
	data := []byte("hello world")

	fc := filecache.NewFileCache(0)

	_, err := fc.Get(cacheID)
	switch {
	case err == nil:
		t.Fatal(err)
	case fc.Count() != 0:
		t.Fatal(fc.Count())
	}

	err = fc.Put(cacheID, storage.File{
		ReadCloser:  ioutil.NopCloser(bytes.NewReader(data)),
		ContentType: contentType,
	})
	test.AssertErrNil(t, err)
	if fc.Count() != 1 {
		t.Fatal(fc.Count())
	}

	f, err := fc.Get(cacheID)
	test.AssertErrNil(t, err)

	fc.ClearExpired()
	if fc.Count() != 1 {
		t.Fatal(fc.Count())
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(f)
	test.AssertErrNil(t, err)

	gotData := buf.Bytes()
	if !bytes.Equal(gotData, data) {
		t.Fatal("exp", data, "got", gotData)
	}

	err = f.Close()
	test.AssertErrNil(t, err)

	fc.ClearExpired()
	if fc.Count() != 0 {
		t.Fatal(fc.Count())
	}
}
