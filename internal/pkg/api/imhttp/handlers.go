package imhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/core"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

const (
	headerContentType  = "Content-Type"
	headerServer       = "Server"
	headerCacheControl = "Cache-Control"
)

const (
	contentTypeJSON = "application/json"
)

type handlers struct {
	exposeErrors bool

	core *core.Core
}

func (h *handlers) ListImages(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx := r.Context()
	query := r.URL.Query()

	category := query.Get("category")

	var limit int
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			err = fmt.Errorf("limit: %w", err)
			h.respondErr(ctx, w, imerrors.NewUserError(err))

			return
		}
	}

	var offset int

	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			err = fmt.Errorf("offset: %w", err)
			h.respondErr(ctx, w, imerrors.NewUserError(err))

			return
		}
	}

	images, err := h.core.ListImages(ctx, category, repository.Pagination{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		h.respondErr(ctx, w, err)

		return
	}

	h.repondJSON(ctx, w, images)
}

func (h *handlers) ListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	categories, err := h.core.ListCategories(ctx)
	if err != nil {
		h.respondErr(ctx, w, err)

		return
	}

	h.repondJSON(ctx, w, categories)
}

func (h *handlers) GetImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := zerolog.Ctx(ctx)

	id := mux.Vars(r)["id"]
	size := imager.ImageSize(mux.Vars(r)["size"])

	f, err := h.core.GetImage(ctx, id, size)
	if err != nil {
		h.respondErr(ctx, w, err)

		return
	}

	w.Header().Set(headerContentType, f.ContentType)

	defer func() {
		cerr := f.Close()
		if cerr != nil {
			l.Warn().Err(cerr).Msg("closing file")
		}
	}()

	_, err = io.Copy(w, f)
	if err != nil {
		l.Warn().Err(err).Msg("copying data to response")

		return
	}
}

func (h *handlers) PutImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	err := r.ParseForm()
	if err != nil {
		err = fmt.Errorf("form: %w", err)
		h.respondErr(ctx, w, imerrors.NewUserError(err))

		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		err = fmt.Errorf("image: %w", err)
		h.respondErr(ctx, w, imerrors.NewUserError(err))

		return
	}

	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("closing form file")
		}
	}()

	im := imager.ImageMeta{
		Author:    r.FormValue("author"),
		WEBSource: r.FormValue("websource"),
		Category:  r.FormValue("category"),
		ID:        r.FormValue("id"),

		MIMEType: fileHeader.Header.Get(headerContentType),
		Size:     fileHeader.Size,
	}

	im, err = h.core.UploadImage(ctx, im, file)
	if err != nil {
		h.respondErr(ctx, w, err)

		return
	}

	h.repondJSON(ctx, w, im)
}

func (h *handlers) repondJSON(ctx context.Context, w http.ResponseWriter, data interface{}) {
	w.Header().Set(headerContentType, contentTypeJSON)

	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msgf("encoding json data")

		return
	}
}

func (h *handlers) respondErr(ctx context.Context, w http.ResponseWriter, err error) {
	l := zerolog.Ctx(ctx)

	var status int
	logEvt := l.Warn()

	switch {
	case err == nil:
		w.WriteHeader(status)

		return
	case errors.As(err, &imerrors.NotFoundError{}):
		status = http.StatusNotFound
	case errors.As(err, &imerrors.ConflictError{}):
		status = http.StatusConflict
	case errors.As(err, &imerrors.UserError{}):
		status = http.StatusBadRequest
	case errors.As(err, &imerrors.MediaTypeError{}):
		status = http.StatusUnsupportedMediaType
	case errors.As(err, &imerrors.OversizeError{}):
		status = http.StatusRequestEntityTooLarge
	case imerrors.IsTemporaryError(err):
		status = http.StatusServiceUnavailable
	default:
		status = http.StatusInternalServerError
		logEvt = zerolog.Ctx(ctx).Error()
	}

	logEvt.Int("status", status).Err(err).Msg("http server error")

	w.WriteHeader(status)

	if h.exposeErrors && err != nil {
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			l.Warn().Err(err).Msg("writing response")
		}
	}
}

func (h *handlers) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	const errNotFound imerrors.Error = "not found"
	h.respondErr(r.Context(), w, imerrors.NewNotFoundError(errNotFound))
}
