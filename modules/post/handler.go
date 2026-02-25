package post

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/leeforge/core/server/httplog"
	"github.com/leeforge/framework/http/responder"
	"github.com/leeforge/framework/logging"
)

// Handler handles post HTTP requests.
type Handler struct {
	service *Service
	logger  logging.Logger
}

// NewHandler creates a new post handler.
func NewHandler(service *Service, logger logging.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// CreatePost handles POST /posts
//
// @Summary Create post
// @Tags Posts
// @Accept json
// @Produce json
// @Param body body CreateRequest true "Post payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/posts [post]
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responder.BindError(w, r, nil)
		return
	}

	result, err := h.service.CreatePost(r.Context(), &req)
	if err != nil {
		h.mapError(w, r, "Failed to create post", err)
		return
	}
	responder.OK(w, r, result)
}

// ListPosts handles GET /posts
//
// @Summary List posts
// @Tags Posts
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param status query string false "Status filter (draft/published/archived)"
// @Param query query string false "Search query"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/posts [get]
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	result, err := h.service.ListPosts(r.Context(), ListFilters{
		Page:     page,
		PageSize: pageSize,
		Status:   r.URL.Query().Get("status"),
		Query:    r.URL.Query().Get("query"),
	})
	if err != nil {
		httplog.Error(h.logger, r, "Failed to list posts", err)
		responder.DatabaseError(w, r, "Failed to list posts")
		return
	}
	responder.OK(w, r, result)
}

// GetPost handles GET /posts/{id}
//
// @Summary Get post by ID
// @Tags Posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/posts/{id} [get]
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		responder.BadRequest(w, r, "Invalid post ID")
		return
	}

	result, err := h.service.GetPost(r.Context(), id)
	if err != nil {
		h.mapError(w, r, "Failed to get post", err)
		return
	}
	responder.OK(w, r, result)
}

// GetPostBySlug handles GET /posts/slug/{slug}
//
// @Summary Get post by slug
// @Tags Posts
// @Produce json
// @Param slug path string true "Post slug"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/posts/slug/{slug} [get]
func (h *Handler) GetPostBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	result, err := h.service.GetPostBySlug(r.Context(), slug)
	if err != nil {
		h.mapError(w, r, "Failed to get post", err)
		return
	}
	responder.OK(w, r, result)
}

// UpdatePost handles PUT /posts/{id}
//
// @Summary Update post
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param body body UpdateRequest true "Post update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/posts/{id} [put]
func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		responder.BadRequest(w, r, "Invalid post ID")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responder.BindError(w, r, nil)
		return
	}

	result, err := h.service.UpdatePost(r.Context(), id, &req)
	if err != nil {
		h.mapError(w, r, "Failed to update post", err)
		return
	}
	responder.OK(w, r, result)
}

// DeletePost handles DELETE /posts/{id}
//
// @Summary Delete post
// @Tags Posts
// @Param id path string true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/posts/{id} [delete]
func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		responder.BadRequest(w, r, "Invalid post ID")
		return
	}

	if err := h.service.DeletePost(r.Context(), id); err != nil {
		h.mapError(w, r, "Failed to delete post", err)
		return
	}
	responder.OK(w, r, map[string]string{"message": "Post deleted successfully"})
}

func (h *Handler) mapError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	switch {
	case errors.Is(err, ErrPostNotFound):
		responder.NotFound(w, r, "Post not found")
	case errors.Is(err, ErrSlugExists):
		responder.Conflict(w, r, "Post slug already exists")
	case errors.Is(err, ErrInvalidPost):
		responder.BadRequest(w, r, "Invalid post data")
	default:
		httplog.Error(h.logger, r, msg, err)
		responder.DatabaseError(w, r, msg)
	}
}
