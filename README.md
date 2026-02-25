# Leeforge Example Template (Echo)

A template project demonstrating how to build a service host using [Leeforge Core](https://github.com/leeforge/core) with the **Echo** HTTP framework.

> Looking for Gin? Switch to the [`main`](https://github.com/leeforge/templates/tree/main) branch.

## Architecture

```
cmd/server/main.go          # Entry point (Echo server)
bootstrap/
  app.go                    # Echo adapter wrapping core runtime
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
3. `buildEcho()` wraps the Chi router handler via `echo.WrapHandler()`
4. Echo instance starts and serves all requests through the core routing layer

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

## Configuration

Before running, replace all `<placeholder>` values in the configuration files:

| Placeholder | File(s) | Description |
|---|---|---|
| `<db-username>` | `configs/database.yaml`, `docker/env/db.env` | PostgreSQL username |
| `<db-password>` | `configs/database.yaml`, `docker/env/db.env` | PostgreSQL password |
| `<redis-password>` | `configs/cache.yaml`, `docker/env/redis.env` | Redis password |
| `<jwt-secret>` | `configs/security.yaml` | JWT signing secret |
| `<init-secret-key>` | `configs/init.yaml` | Application init secret key |
| `<registry-host>` | `docker/docker-compose.test.yaml`, `.deploy.env.common`, `.deploy.env.prod` | Docker registry host |
| `<remote-user>` | `.deploy.env.common`, `.deploy.env.prod` | SSH remote deploy user |
| `<remote-host>` | `.deploy.env.common`, `.deploy.env.prod` | SSH remote deploy host |

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
