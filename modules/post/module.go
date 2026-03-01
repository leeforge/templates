package post

import (
	"strings"

	"entgo.io/ent/dialect"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/leeforge/core/core"
	frameLogging "github.com/leeforge/framework/logging"

	examplesent "leeforge-example-service/ent"
)

// PostModule implements corecore.Module for post management.
type PostModule struct {
	handler *Handler
}

func (m *PostModule) Name() string { return "post" }

func (m *PostModule) RegisterPublicRoutes(_ chi.Router) {}

func (m *PostModule) RegisterPrivateRoutes(router chi.Router) {
	router.Route("/posts", func(r chi.Router) {
		r.Get("/", m.handler.ListPosts)
		r.Post("/", m.handler.CreatePost)
		r.Get("/slug/{slug}", m.handler.GetPostBySlug)
		r.Get("/{id}", m.handler.GetPost)
		r.Put("/{id}", m.handler.UpdatePost)
		r.Delete("/{id}", m.handler.DeletePost)
	})
}

// NewPostModule is a core.ModuleFactory that creates a PostModule.
// It creates its own ent client from the database config for examples-owned schemas.
func NewPostModule(logger frameLogging.Logger, deps *core.Dependencies) core.Module {
	dsn := deps.Config.Database.DSN()
	driver := resolveDriver(dsn)

	client, err := examplesent.Open(driver, dsn)
	if err != nil {
		logger.Error("failed to create examples ent client", zap.Error(err))
		return &PostModule{}
	}

	svc := NewService(client)
	return &PostModule{handler: NewHandler(svc, logger)}
}

func resolveDriver(dsn string) string {
	lower := strings.ToLower(strings.TrimSpace(dsn))
	switch {
	case strings.HasPrefix(lower, "postgres://"), strings.HasPrefix(lower, "postgresql://"):
		return dialect.Postgres
	case strings.HasPrefix(lower, "sqlite://"), strings.HasPrefix(lower, "file:"):
		return dialect.SQLite
	case strings.HasPrefix(lower, "mysql://"), strings.HasPrefix(lower, "mariadb://"):
		return dialect.MySQL
	default:
		return dialect.Postgres
	}
}

var _ core.Module = (*PostModule)(nil)
