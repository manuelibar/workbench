package storage

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/storage/upload-url", h.uploadURL)
	mux.HandleFunc("GET /api/storage/resources", h.list)
	mux.HandleFunc("GET /api/storage/{id}/stats", h.stats)
	mux.HandleFunc("GET /api/storage/{id}/download-url", h.downloadURL)
	mux.HandleFunc("POST /api/storage/{id}/update-url", h.updateURL)
	mux.HandleFunc("POST /api/storage/multipart/start", h.multipartStart)
	mux.HandleFunc("POST /internal/webhook/s3-event", h.s3Event)
	return mux
}

func (h *Handler) uploadURL(w http.ResponseWriter, r *http.Request) {
	var req UploadURLRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.UploadURL(r.Context(), req)
	respond(w, result, err)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	result, err := h.service.List(r.Context(), query.Get("org_id"), query.Get("project_id"), query.Get("resource_type"))
	respond(w, map[string]any{"resources": result}, err)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	ref, ok := h.refFromQuery(w, r)
	if !ok {
		return
	}
	result, err := h.service.Stats(r.Context(), ref)
	respond(w, result, err)
}

func (h *Handler) downloadURL(w http.ResponseWriter, r *http.Request) {
	ref, ok := h.refFromQuery(w, r)
	if !ok {
		return
	}
	result, err := h.service.DownloadURL(r.Context(), ref)
	respond(w, result, err)
}

func (h *Handler) updateURL(w http.ResponseWriter, r *http.Request) {
	var req UpdateURLRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.ResourceID = firstNonEmpty(strings.TrimSpace(req.ResourceID), r.PathValue("id"))
	result, err := h.service.UpdateURL(r.Context(), req)
	respond(w, result, err)
}

func (h *Handler) multipartStart(w http.ResponseWriter, r *http.Request) {
	var req MultipartStartRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.StartMultipart(r.Context(), req)
	respond(w, result, err)
}

func (h *Handler) s3Event(w http.ResponseWriter, r *http.Request) {
	err := h.service.HandleS3Event(r.Context(), r.Body)
	respond(w, map[string]string{"status": "ok"}, err)
}

func (h *Handler) refFromQuery(w http.ResponseWriter, r *http.Request) (ResourceRef, bool) {
	query := r.URL.Query()
	ref, err := NewResourceRef(query.Get("org_id"), query.Get("project_id"), query.Get("resource_type"), r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return ResourceRef{}, false
	}
	return ref, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		respond(w, nil, invalid("storage.http.json_invalid", "request JSON is invalid"))
		return false
	}
	return true
}

func respond(w http.ResponseWriter, value any, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrInvalid):
			status = http.StatusBadRequest
		case errors.Is(err, ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, ErrDependencyFailed):
			status = http.StatusBadGateway
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"title": err.Error(),
				"code":  errorCode(err),
			},
		})
		return
	}
	if value == nil {
		value = map[string]string{"status": "ok"}
	}
	_ = json.NewEncoder(w).Encode(value)
}

func errorCode(err error) string {
	var storageErr *Error
	if errors.As(err, &storageErr) && storageErr.Code != "" {
		return storageErr.Code
	}
	switch {
	case errors.Is(err, ErrInvalid):
		return "storage.invalid"
	case errors.Is(err, ErrNotFound):
		return "storage.not_found"
	case errors.Is(err, ErrDependencyFailed):
		return "storage.dependency_failed"
	default:
		return "storage.internal"
	}
}
