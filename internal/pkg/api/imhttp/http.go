package imhttp

import (
	"net/http"
	"time"

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

func NewServer(cfg config.Server, es Essentials) *Server {
	h := handlers{
		core:         es.Core,
		exposeErrors: cfg.ExposeErrors,
	}

	r := mux.NewRouter()

	apiV1 := r.PathPrefix("/api/v1").Subrouter()

	apiV1.Path("/images").
		Methods(http.MethodGet).
		HandlerFunc(h.ListImages)

	apiV1.Path("/images/{id}").
		Methods(http.MethodGet).
		HandlerFunc(h.GetImage)

	apiV1.Path("/categories").
		Methods(http.MethodGet).
		HandlerFunc(h.ListCategories)

	internalAPIV1 := r.PathPrefix("/internal/api/v1").Subrouter()

	internalAPIV1.
		Path("/images").
		Methods(http.MethodPut).
		HandlerFunc(h.PutImage)

	r.Use(
		mux.CORSMethodMiddleware(r),
		middlewareServerHeader(cfg.Name),
		middlewareCacheControl(time.Duration(cfg.CacheControlMaxAge)),
		middlewareLogger(es.Logger),
		middlewareDump,
		middlewareRecoverPanic,
	)

	r.NotFoundHandler = http.HandlerFunc(h.NotFoundHandler)

	return &Server{
		Server: http.Server{
			Addr:    cfg.Address,
			Handler: r,

			ReadTimeout:  time.Duration(cfg.ReadTimeout),
			WriteTimeout: time.Duration(cfg.WriteTimeout),
		},
	}
}
