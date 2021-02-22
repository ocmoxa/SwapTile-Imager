package filecache

import (
	"io/ioutil"
	"sync"
	"time"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage"
)

const errNotFound = imerrors.Error("not found in cache")

// FileCache implements storage.FileCache.
type FileCache struct {
	maxAge time.Duration

	mu       sync.Mutex
	entities map[string]*cacheEntity
}

// NewFileCache creates new in-memory cache for files.
func NewFileCache(maxAge time.Duration) *FileCache {
	return &FileCache{
		maxAge: maxAge,

		mu:       sync.Mutex{},
		entities: make(map[string]*cacheEntity),
	}
}

func (fc *FileCache) Put(id string, f storage.File) (err error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if ce, ok := fc.entities[id]; ok {
		ce.expiresAt = time.Now().Add(fc.maxAge)

		return nil
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	fc.entities[id] = &cacheEntity{
		ContentType: f.ContentType,

		data:      data,
		expiresAt: time.Now().Add(fc.maxAge),

		uses: 0,
	}

	return nil
}

func (fc *FileCache) Get(id string) (f storage.File, err error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if ce, ok := fc.entities[id]; ok {
		return ce.File(), nil
	}

	return storage.File{}, imerrors.NewNotFoundError(errNotFound)
}

func (fc *FileCache) ClearExpired() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	for e, ce := range fc.entities {
		if ce.CanRemove() {
			delete(fc.entities, e)
		}
	}
}

func (fc *FileCache) Count() int {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	return len(fc.entities)
}
