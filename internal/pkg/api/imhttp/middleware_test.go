package imhttp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestMiddlewareServerHeader(t *testing.T) {
	const server = "test-server"

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {})

	h = middlewareServerHeader(server)(h).ServeHTTP
	h(w, r)

	if w.Header().Get(headerServer) != server {
		t.Fatal("exp", server, "got", w.Header().Get(headerServer))
	}
}

func TestMiddlewareLogger(t *testing.T) {
	l := zerolog.New(ioutil.Discard).Level(zerolog.DebugLevel)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	var called bool
	h := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		called = true

		level := zerolog.Ctx(r.Context()).GetLevel()
		if level != l.GetLevel() {
			t.Error("exp", level, "got", l.GetLevel())
		}
	})

	h = middlewareLogger(l)(h).ServeHTTP
	h(w, r)

	if !called {
		t.Fatal()
	}
}

func TestMiddlewareDump(t *testing.T) {
	var logBuf bytes.Buffer
	l := zerolog.New(&logBuf).Level(zerolog.DebugLevel)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(l.WithContext(r.Context()))
	w := httptest.NewRecorder()

	h := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {})

	h = middlewareDump(h).ServeHTTP
	h(w, r)

	if logBuf.Len() == 0 {
		t.Fatal(logBuf.String())
	}
}

func TestMiddlewareRecoverPanic(t *testing.T) {
	const expCode = http.StatusInternalServerError

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		panic("test")
	})

	h = middlewareRecoverPanic(h).ServeHTTP
	h(w, r)

	if w.Code != expCode {
		t.Fatal("exp", expCode, "got", w.Code)
	}
}

func TestMiddlewareCacheControl_MaxAge(t *testing.T) {
	const maxAge = 5 * time.Second
	const expCacheControl = "max-age=5"

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {})

	h = middlewareCacheControl(maxAge)(h).ServeHTTP
	h(w, r)

	if w.Header().Get(headerCacheControl) != expCacheControl {
		t.Fatal("exp", expCacheControl, "got", w.Header().Get(headerCacheControl))
	}
}

func TestMiddlewareCacheControl_NoCache(t *testing.T) {
	const maxAge = 0
	const expCacheControl = "no-cache"

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {})

	h = middlewareCacheControl(maxAge)(h).ServeHTTP
	h(w, r)

	if w.Header().Get(headerCacheControl) != expCacheControl {
		t.Fatal("exp", expCacheControl, "got", w.Header().Get(headerCacheControl))
	}
}
