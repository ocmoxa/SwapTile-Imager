package imhttp

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

func middlewareServerHeader(server string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(headerServer, server)

			next.ServeHTTP(w, r)
		})
	}
}

func middlewareLogger(log zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = log.WithContext(ctx)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func middlewareDump(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		log.Debug().
			Interface("header", r.Header).
			Str("method", r.Method).
			Stringer("url", r.URL).
			Msg("requesting")

		srw := &statusResponseWriter{ResponseWriter: w}
		next.ServeHTTP(srw, r)

		var logEvt *zerolog.Event
		switch {
		case srw.Code < http.StatusBadRequest:
			logEvt = log.Debug()
		case srw.Code < http.StatusInternalServerError:
			logEvt = log.Warn()
		default:
			logEvt = log.Error()
		}

		logEvt.
			Interface("header", r.Header).
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", srw.Code).
			Msg("respond")
	})
}

func middlewareRecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		defer func() {
			if rerr := recover(); rerr != nil {
				w.WriteHeader(http.StatusInternalServerError)

				zerolog.Ctx(ctx).Error().
					Str("stack", string(debug.Stack())).
					Msgf("panic: %v", rerr)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func middlewareCacheControl(maxAge time.Duration) func(next http.Handler) http.Handler {
	seconds := int(maxAge.Seconds())

	maxAgeHeaderValue := "no-cache"
	if seconds > 0 {
		maxAgeHeaderValue = "max-age=" + strconv.Itoa(seconds)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(headerCacheControl, maxAgeHeaderValue)

			next.ServeHTTP(w, r)
		})
	}
}

type statusResponseWriter struct {
	http.ResponseWriter
	Code int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.Code = code
	w.ResponseWriter.WriteHeader(code)
}
