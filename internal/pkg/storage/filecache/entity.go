package filecache

import (
	"bytes"
	"io"
	"sync/atomic"
	"time"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage"
)

type cacheEntity struct {
	ContentType string

	data      []byte
	expiresAt time.Time

	uses int64
}

func (ce *cacheEntity) File() storage.File {
	if ce == nil {
		return storage.File{}
	}

	atomic.AddInt64(&ce.uses, 1)

	return storage.File{
		ContentType: ce.ContentType,
		ReadCloser: readerCloser{
			Reader: bytes.NewReader(ce.data),
			Closer: closerFunc(func() error {
				atomic.AddInt64(&ce.uses, -1)

				return nil
			}),
		},
	}
}

func (ce *cacheEntity) CanRemove() bool {
	if ce == nil {
		return true
	}

	return atomic.LoadInt64(&ce.uses) == 0 &&
		time.Since(ce.expiresAt) >= 0
}

type closerFunc func() error

func (cf closerFunc) Close() error {
	return cf()
}

type readerCloser struct {
	io.Reader
	io.Closer
}
