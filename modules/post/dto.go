package post

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors.
var (
	ErrPostNotFound = errors.New("post not found")
	ErrSlugExists   = errors.New("post slug already exists")
	ErrInvalidPost  = errors.New("invalid post data")
)

// --- Requests ---

// CreateRequest is the input for creating a new post.
type CreateRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Slug    string   `json:"slug"`
	Status  string   `json:"status,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// UpdateRequest is the input for updating a post.
type UpdateRequest struct {
	Title   string   `json:"title,omitempty"`
	Content string   `json:"content,omitempty"`
	Slug    string   `json:"slug,omitempty"`
	Status  string   `json:"status,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// ListFilters holds query parameters for listing posts.
type ListFilters struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	Status   string `json:"status,omitempty"`
	Query    string `json:"query,omitempty"`
}

// --- Responses ---

// PostDTO is the post representation returned by the API.
type PostDTO struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Slug        string     `json:"slug"`
	Status      string     `json:"status"`
	Tags        []string   `json:"tags"`
	ViewCount   int        `json:"viewCount"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
}

// ListResult is the paginated post list response.
type ListResult struct {
	Posts      []*PostDTO `json:"posts"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	TotalPages int        `json:"totalPages"`
}
