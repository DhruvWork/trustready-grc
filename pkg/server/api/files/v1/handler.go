// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package files_v1

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/brand"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/jsonx"
)

const presignedURLExpiry = 1 * time.Hour

type Handler struct {
	logger  *log.Logger
	fileSvc *filemanager.Service
	probo   *probo.Service
	iamSvc  *iam.Service
	assets  *brand.Assets
	baseURL *baseurl.BaseURL
}

func NewMux(
	logger *log.Logger,
	fileSvc *filemanager.Service,
	proboSvc *probo.Service,
	iamSvc *iam.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
	baseURL *baseurl.BaseURL,
) *chi.Mux {
	h := &Handler{
		logger:  logger,
		fileSvc: fileSvc,
		probo:   proboSvc,
		iamSvc:  iamSvc,
		assets:  brand.NewAssets(),
		baseURL: baseURL,
	}

	r := chi.NewRouter()

	r.Get("/static/{file}", h.handleGetStaticFile)
	r.Get("/public/{fileID}", h.handleGetPublicFile)

	r.Group(func(r chi.Router) {
		r.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
		r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
		r.Use(authn.NewOAuth2AccessTokenMiddleware(iamSvc))
		r.Use(authn.NewIdentityPresenceMiddleware(baseURL))
		r.Get("/{fileID}", h.handleGetFile)
	})

	return r
}

func (h *Handler) handleGetStaticFile(w http.ResponseWriter, r *http.Request) {
	file := chi.URLParam(r, "file")

	if _, statErr := h.assets.Stat(file); statErr != nil {
		jsonx.RenderNotFound(w, fmt.Errorf("file not found"))
		return
	}

	// ServeAssets honors the ETag header we set above for If-None-Match (and
	// If-Range), so it emits 304 Not Modified and handles range requests without
	// any extra conditional logic here.
	h.assets.ServeAssets(w, r, file)
}

func (h *Handler) handleGetPublicFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "fileID")

	fileID, err := gid.ParseGID(fileIDStr)
	if err != nil {
		jsonx.RenderNotFound(w, fmt.Errorf("file not found"))
		return
	}

	err = h.fileSvc.ServePublicFile(r.Context(), w, r, fileID)
	if errors.Is(err, filemanager.ErrPublicFileNotFound) {
		jsonx.RenderNotFound(w, fmt.Errorf("file not found"))
		return
	}

	if err != nil {
		h.logger.ErrorCtx(
			r.Context(),
			"cannot serve public file",
			log.Error(err),
			log.String("file_id", fileIDStr),
		)
		jsonx.RenderInternalServerError(w)
	}
}

func (h *Handler) handleGetFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "fileID")

	fileID, err := gid.ParseGID(fileIDStr)
	if err != nil {
		jsonx.RenderNotFound(w, fmt.Errorf("file not found"))
		return
	}

	ctx := r.Context()
	identity := authn.IdentityFromContext(ctx)
	session := authn.SessionFromContext(ctx)

	params := iam.AuthorizeParams{
		Principal:          identity.ID,
		Resource:           fileID,
		Action:             probo.ActionFileGet,
		ResourceAttributes: make(map[string]string),
	}
	if session != nil {
		params.Session = &session.ID
	}

	scope, err := h.iamSvc.Authorizer.Authorize(ctx, params)
	if err != nil {
		if scopeErr, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](err); ok {
			bearertoken.SetBearerInsufficientScope(w, h.baseURL, scopeErr.Scopes...)
			jsonx.RenderForbidden(w)

			return
		}

		if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
			jsonx.RenderForbidden(w)
			return
		}

		jsonx.RenderNotFound(w, fmt.Errorf("file not found"))

		return
	}

	f, err := h.probo.Files.Get(ctx, scope, fileID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			jsonx.RenderNotFound(w, fmt.Errorf("file not found"))
			return
		}

		h.logger.ErrorCtx(ctx, "cannot get file", log.Error(err), log.String("file_id", fileIDStr))
		jsonx.RenderInternalServerError(w)

		return
	}

	// Stream the file through the app rather than redirecting to a presigned
	// object-storage URL. Self-hosted deployments (e.g. the bundled SeaweedFS
	// gateway) publish an internal-only storage endpoint, so a presigned URL is
	// not reachable from the browser; proxying keeps downloads working and stays
	// same-origin (the session cookie is sent).
	conds := filemanager.FileConditions{
		IfNoneMatch: r.Header.Get("If-None-Match"),
		IfRange:     r.Header.Get("If-Range"),
		Range:       r.Header.Get("Range"),
	}
	if ifModifiedSince := r.Header.Get("If-Modified-Since"); ifModifiedSince != "" {
		if t, parseErr := http.ParseTime(ifModifiedSince); parseErr == nil {
			conds.IfModifiedSince = t
		}
	}

	obj, err := h.fileSvc.OpenFile(ctx, f, conds)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot open file", log.Error(err), log.String("file_id", fileIDStr))
		jsonx.RenderInternalServerError(w)

		return
	}

	defer func() { _ = obj.Body.Close() }()

	w.Header().Set("Accept-Ranges", "bytes")
	if obj.ETag != "" {
		w.Header().Set("ETag", obj.ETag)
	}

	if obj.NotModified {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if obj.RangeNotSatisfiable {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", f.FileSize))
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)

		return
	}

	w.Header().Set("Content-Type", f.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(obj.ContentLength, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", f.FileName))

	if !obj.LastModified.IsZero() {
		w.Header().Set("Last-Modified", obj.LastModified.UTC().Format(http.TimeFormat))
	}

	if obj.PartialContent {
		w.Header().Set("Content-Range", obj.ContentRange)
		w.WriteHeader(http.StatusPartialContent)
	}

	if _, err := io.Copy(w, obj.Body); err != nil {
		h.logger.ErrorCtx(ctx, "cannot stream file", log.Error(err), log.String("file_id", fileIDStr))
		return
	}
}
