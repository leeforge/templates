package post

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/leeforge/core"
	examplesent "leeforge-example-service/ent"
	entpost "leeforge-example-service/ent/post"
)

// Service handles post CRUD operations.
type Service struct {
	client *examplesent.Client
}

// NewService creates a new post service.
func NewService(client *examplesent.Client) *Service {
	return &Service{client: client}
}

// CreatePost creates a new post owned by the authenticated user.
func (s *Service) CreatePost(ctx context.Context, req *CreateRequest) (*PostDTO, error) {
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Slug) == "" || strings.TrimSpace(req.Content) == "" {
		return nil, ErrInvalidPost
	}

	authorID, ok := core.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("missing user context")
	}

	builder := s.client.Post.Create().
		SetTitle(strings.TrimSpace(req.Title)).
		SetContent(req.Content).
		SetSlug(strings.TrimSpace(req.Slug)).
		SetAuthorID(authorID)

	if req.Status != "" {
		builder.SetStatus(entpost.Status(req.Status))
	}
	if len(req.Tags) > 0 {
		builder.SetTags(req.Tags)
	}

	p, err := builder.Save(ctx)
	if err != nil {
		if examplesent.IsConstraintError(err) {
			return nil, ErrSlugExists
		}
		return nil, fmt.Errorf("create post: %w", err)
	}

	return toDTO(p), nil
}

// ListPosts returns a paginated list of posts.
func (s *Service) ListPosts(ctx context.Context, filters ListFilters) (*ListResult, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 20
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}

	query := s.client.Post.Query().Where(entpost.DeletedAtIsNil())
	if filters.Status != "" {
		query = query.Where(entpost.StatusEQ(entpost.Status(filters.Status)))
	}
	if filters.Query != "" {
		q := strings.TrimSpace(filters.Query)
		query = query.Where(entpost.Or(
			entpost.TitleContainsFold(q),
			entpost.SlugContainsFold(q),
		))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count posts: %w", err)
	}

	offset := (filters.Page - 1) * filters.PageSize
	items, err := query.
		Offset(offset).
		Limit(filters.PageSize).
		Order(examplesent.Desc(entpost.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}

	dtos := make([]*PostDTO, len(items))
	for i, p := range items {
		dtos[i] = toDTO(p)
	}

	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	return &ListResult{
		Posts:      dtos,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetPost returns a single post by ID.
func (s *Service) GetPost(ctx context.Context, id uuid.UUID) (*PostDTO, error) {
	p, err := s.client.Post.Get(ctx, id)
	if err != nil {
		if examplesent.IsNotFound(err) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("get post: %w", err)
	}
	if !p.DeletedAt.IsZero() {
		return nil, ErrPostNotFound
	}
	return toDTO(p), nil
}

// GetPostBySlug returns a single post by slug.
func (s *Service) GetPostBySlug(ctx context.Context, slug string) (*PostDTO, error) {
	p, err := s.client.Post.Query().
		Where(entpost.SlugEQ(slug), entpost.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if examplesent.IsNotFound(err) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("get post by slug: %w", err)
	}
	return toDTO(p), nil
}

// UpdatePost updates the fields of an existing post.
func (s *Service) UpdatePost(ctx context.Context, id uuid.UUID, req *UpdateRequest) (*PostDTO, error) {
	p, err := s.client.Post.Get(ctx, id)
	if err != nil {
		if examplesent.IsNotFound(err) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("get post: %w", err)
	}
	if !p.DeletedAt.IsZero() {
		return nil, ErrPostNotFound
	}

	updater := s.client.Post.UpdateOne(p)
	if req.Title != "" {
		updater.SetTitle(strings.TrimSpace(req.Title))
	}
	if req.Content != "" {
		updater.SetContent(req.Content)
	}
	if req.Slug != "" {
		updater.SetSlug(strings.TrimSpace(req.Slug))
	}
	if req.Status != "" {
		newStatus := entpost.Status(req.Status)
		updater.SetStatus(newStatus)
		if newStatus == entpost.StatusPublished && p.PublishedAt.IsZero() {
			updater.SetPublishedAt(time.Now())
		}
		if newStatus == entpost.StatusArchived && p.ArchivedAt.IsZero() {
			updater.SetArchivedAt(time.Now())
		}
	}
	if req.Tags != nil {
		updater.SetTags(req.Tags)
	}

	updated, err := updater.Save(ctx)
	if err != nil {
		if examplesent.IsConstraintError(err) {
			return nil, ErrSlugExists
		}
		return nil, fmt.Errorf("update post: %w", err)
	}
	return toDTO(updated), nil
}

// DeletePost soft-deletes a post.
func (s *Service) DeletePost(ctx context.Context, id uuid.UUID) error {
	p, err := s.client.Post.Get(ctx, id)
	if err != nil {
		if examplesent.IsNotFound(err) {
			return ErrPostNotFound
		}
		return fmt.Errorf("get post: %w", err)
	}
	if !p.DeletedAt.IsZero() {
		return ErrPostNotFound
	}

	if _, err := s.client.Post.UpdateOneID(id).SetDeletedAt(time.Now()).Save(ctx); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}

func toDTO(p *examplesent.Post) *PostDTO {
	dto := &PostDTO{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		Slug:      p.Slug,
		Status:    string(p.Status),
		Tags:      p.Tags,
		ViewCount: p.ViewCount,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if !p.PublishedAt.IsZero() {
		t := p.PublishedAt
		dto.PublishedAt = &t
	}
	return dto
}
