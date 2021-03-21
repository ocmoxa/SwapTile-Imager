package imhttp

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
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

func middlewareMetrics(registerer prometheus.Registerer) func(next http.Handler) http.Handler {
	histogramVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "swaptile",
		Subsystem: "imager",
		Help:      "HTTP API",
		Name:      "http_duration",
	}, []string{"route", "code"})
	registerer.MustRegister(histogramVec)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := "unknown"
			if cr := mux.CurrentRoute(r); cr != nil {
				var err error
				route, err = cr.GetPathTemplate()
				if err != nil {
					zerolog.Ctx(r.Context()).Warn().Msg("getting template")
				}
			} else if url := r.URL; url != nil {
				route = url.String()
			}

			srw := &statusResponseWriter{ResponseWriter: w}
			t := time.Now()
			next.ServeHTTP(srw, r)
			d := time.Since(t)

			histogramVec.
				WithLabelValues(route, strconv.Itoa(srw.Code)).
				Observe(d.Seconds())
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
