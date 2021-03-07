package imhttp

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/ocmoxa/SwapTile-Imager/docs"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/core"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type Server struct {
	http.Server
}

type Essentials struct {
	zerolog.Logger
	*core.Core
}

func NewServer(es Essentials, cfg config.Server) (*Server, error) {
	h := handlers{
		core:         es.Core,
		exposeErrors: cfg.ExposeErrors,
	}

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(h.NotFoundHandler)

	mountDebug(r)
	err := mountSwagger(r)
	if err != nil {
		return nil, fmt.Errorf("mounting swagger: %w", err)
	}

	r.Use(
		mux.CORSMethodMiddleware(r),
		middlewareServerHeader(cfg.Name),
		middlewareLogger(es.Logger),
		middlewareDump,
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

func mountSwagger(r *mux.Router) (err error) {
	fsys := docs.Swagger()

	r.Handle("/", http.RedirectHandler("/swagger/ui/", http.StatusPermanentRedirect))

	h := http.FileServer(http.FS(fsys))
	h = http.StripPrefix("/swagger", h)

	r.PathPrefix("/swagger/").Handler(h)

	return nil
}

func mountDebug(r *mux.Router) {
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
		Path("/images/shuffle").
		Methods(http.MethodPost).
		HandlerFunc(h.PostShuffle)
}
