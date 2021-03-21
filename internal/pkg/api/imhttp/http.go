package imhttp

import (
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/ocmoxa/SwapTile-Imager/docs"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/core"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// Server HTTP.
type Server struct {
	http.Server
}

// Essentials of the server.
type Essentials struct {
	zerolog.Logger
	*core.Core
	PromRegistry *prometheus.Registry
}

// NewServer creates new http server.
func NewServer(es Essentials, cfg config.Server) (*Server, error) {
	h := handlers{
		core:         es.Core,
		exposeErrors: cfg.ExposeErrors,
	}

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(h.notFoundHandler)

	mountDebug(r, &h, es.PromRegistry)
	mountSwagger(r)

	r.Use(
		mux.CORSMethodMiddleware(r),
		middlewareServerHeader(cfg.Name),
		middlewareLogger(es.Logger),
		middlewareDump,
		middlewareMetrics(es.PromRegistry),
		middlewareRecoverPanic,
	)

	mountInternalAPI(r, &h)

	mountPublicAPI(r, &h, cfg)

	return &Server{
		Server: http.Server{
			Addr:    cfg.Address,
			Handler: r,

			ReadTimeout:  time.Duration(cfg.ReadTimeout),
			WriteTimeout: time.Duration(cfg.WriteTimeout),
		},
	}, nil
}

func mountSwagger(r *mux.Router) {
	fsys := docs.Swagger()

	r.Handle("/", http.RedirectHandler("/swagger/ui/", http.StatusPermanentRedirect))

	h := http.FileServer(http.FS(fsys))
	h = http.StripPrefix("/swagger", h)

	r.PathPrefix("/swagger/").Handler(h)
}

func mountDebug(r *mux.Router, h *handlers, promRegistry *prometheus.Registry) {
	r.Path("/health").Methods(http.MethodGet).HandlerFunc(h.getHealth)

	debugAPI := r.PathPrefix("/debug").Subrouter()
	debugAPI.HandleFunc("/pprof/", pprof.Index)
	debugAPI.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	debugAPI.HandleFunc("/pprof/profile", pprof.Profile)
	debugAPI.HandleFunc("/pprof/trace", pprof.Trace)
	debugAPI.HandleFunc("/pprof/symbol", pprof.Symbol)
	debugAPI.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	debugAPI.Handle("/pprof/heap", pprof.Handler("heap"))
	debugAPI.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
	debugAPI.Handle("/pprof/block", pprof.Handler("block"))
	debugAPI.Handle("/pprof/allocs", pprof.Handler("allocs"))

	promHandler := promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{
		Registry: promRegistry,
	})
	r.Path("/metrics").Methods(http.MethodGet).Handler(promHandler)
}

func mountPublicAPI(r *mux.Router, h *handlers, cfg config.Server) {
	apiV1 := r.PathPrefix("/api/v1").Subrouter()

	apiV1.Path("/images").
		Methods(http.MethodGet).
		HandlerFunc(h.ListImages)

	apiV1.Path("/images/{id}/{size}").
		Methods(http.MethodGet).
		HandlerFunc(h.GetImage)

	apiV1.Path("/categories").
		Methods(http.MethodGet).
		HandlerFunc(h.ListCategories)

	apiV1.Use(
		middlewareCacheControl(
			time.Duration(cfg.CacheControlMaxAge),
		),
	)
}

func mountInternalAPI(r *mux.Router, h *handlers) {
	internalAPIV1 := r.PathPrefix("/internal/api/v1").Subrouter()

	internalAPIV1.
		Path("/images").
		Methods(http.MethodPut).
		HandlerFunc(h.PutImage)

	internalAPIV1.
		Path("/images/{image_id}").
		Methods(http.MethodDelete).
		HandlerFunc(h.DeleteImage)

	internalAPIV1.
		Path("/images/shuffle").
		Methods(http.MethodPost).
		HandlerFunc(h.PostShuffle)
}
