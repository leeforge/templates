# Leeforge Example Template (Gin)

A template project demonstrating how to build a service host using [Leeforge Core](https://github.com/leeforge/core) with the **Gin** HTTP framework.

> Looking for Echo? Switch to the [`echo`](https://github.com/leeforge/templates/tree/echo) branch.

## Architecture

```
cmd/server/main.go          # Entry point (Gin server)
bootstrap/
  app.go                    # Gin adapter wrapping core runtime
  plugin_registrar.go       # Plugin registration (Tenant, OU)
config/config.go            # Configuration structures
configs/*.yaml              # YAML configuration files
modules/post/               # Example business module (vertical slice)
  module.go                 # Module interface implementation
  handler.go                # HTTP handlers (chi-based)
  service.go                # Business logic
  dto.go                    # Data transfer objects
ent/                        # Ent ORM generated code
docs/                       # Swagger documentation
docker/                     # Docker & Compose files
```

## How It Works

1. `bootstrap.NewApp()` creates a `core.Runtime` with config, plugins, and modules
2. Core runtime bootstraps all modules/plugins using an internal Chi router
3. `buildGinEngine()` wraps the Chi router handler via `gin.WrapH()`
4. Gin engine starts and serves all requests through the core routing layer

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 18
- Redis

### Setup Infrastructure

```bash
make setup-db      # Start PostgreSQL (port 15436)
make setup-redis   # Start Redis (port 16379)
```

### Run Development Server

```bash
make dev           # Requires 'air' for hot-reload
# or
go run ./cmd/server
```

### Access

- API: `http://localhost:8080/api/v1/health`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Swagger JSON: `http://localhost:8080/swagger/doc.json`

## Code Generation

```bash
make generate      # Ent ORM code generation
make swagger       # Swagger documentation generation
```

## Testing

```bash
go test ./...
```

## Deployment

```bash
make build         # Build amd64 Docker image
make remote-deploy # Full remote deployment
```

See `Makefile` for all available targets.

## Module Pattern

Each business module follows the vertical slice pattern:

```go
// modules/post/module.go
func NewPostModule(deps corecore.Dependencies) corecore.Module {
    svc := NewPostService(deps.Client)
    return &PostModule{handler: NewPostHandler(svc)}
}

func (m *PostModule) Setup(r chi.Router) {
    r.Route("/posts", func(r chi.Router) {
        r.Get("/", m.handler.List)
        r.Post("/", m.handler.Create)
        // ...
    })
}
```

## License

MIT
