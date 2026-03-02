# Leeforge Example Template (Gin)

> 基于 [Leeforge Core](https://github.com/leeforge/core) 构建服务的示例模板，使用 **Gin** 作为 HTTP 框架。
> 文档基于 `main` 分支，如有出入以代码为准。
> 寻找 Echo 版本？切换到 [`echo`](https://github.com/leeforge/templates/tree/echo) 分支。

---

## 目录

- [项目概览](#项目概览)
- [快速上手](#快速上手)
- [配置参考](#配置参考)
- [核心概念](#核心概念)
- [开发新业务模块](#开发新业务模块)
- [测试](#测试)
- [部署](#部署)
- [Makefile 命令速查](#makefile-命令速查)

---

## 项目概览

### 这是什么

`leeforge/templates` 是一个**可直接 fork 使用的服务脚手架**，展示了如何用 Leeforge Core 框架搭建一个具备以下能力的后端服务：

- 多租户 / OU 权限体系（开箱即用）
- 基于 Ent ORM 的数据库访问
- Swagger API 文档自动生成
- 垂直切片业务模块模式
- 生产就绪的 Docker 部署流程

### 技术栈

| 层次 | 技术 |
|------|------|
| HTTP 框架 | [Gin](https://github.com/gin-gonic/gin) |
| 路由层 | [Chi](https://github.com/go-chi/chi)（由 Core 管理） |
| ORM | [Ent](https://entgo.io/) |
| 数据库 | PostgreSQL 18 |
| 缓存 | Redis |
| 日志 | [Zap](https://github.com/uber-go/zap) |
| API 文档 | Swagger（swag） |

### 目录结构

```
cmd/server/
  main.go                   # 入口：启动 Gin 服务器

bootstrap/
  app.go                    # 创建 Core Runtime，包装成 Gin Engine
  app_test.go               # 集成测试
  plugin_registrar.go       # 注册业务插件（Tenant、OU）
  plugin_registrar_test.go  # 插件注册测试

config/
  config.go                 # 配置结构体定义（含 Default() 函数）

configs/
  server.yaml               # 服务器与 CORS 配置
  database.yaml             # 数据库连接配置
  cache.yaml                # Redis 配置
  security.yaml             # JWT 与限流配置
  access_control.yaml       # 多租户/OU 开关
  log.yaml                  # 日志格式与存储配置
  tracing.yaml              # 链路追踪（OpenTelemetry）
  metrics.yaml              # Prometheus 指标端口
  init.yaml                 # 初始化密钥
  frontend.yaml             # 前端 URL（CORS 白名单等）
  captcha.yaml              # 验证码配置

modules/
  post/                     # 示例业务模块（垂直切片）
    module.go               # Module 接口实现，路由注册
    handler.go              # HTTP 处理层
    service.go              # 业务逻辑层
    dto.go                  # 请求/响应数据结构 + Sentinel 错误

ent/
  schema/
    post.go                 # Ent Schema 定义
  *.go                      # 自动生成的 ORM 代码（勿手动修改）

docs/                       # Swagger 生成文档（勿手动修改）
docker/                     # Dockerfile 与 Compose 文件
```

---

## 快速上手

### 前置依赖

| 工具 | 版本 | 安装 |
|------|------|------|
| Go | 1.25+ | https://go.dev/dl |
| air | 最新 | `go install github.com/air-verse/air@latest` |
| swag | 最新 | `go install github.com/swaggo/swag/cmd/swag@latest` |
| Docker | 任意 | https://docs.docker.com/get-docker |

### 1. 启动基础设施

```bash
make setup-db     # 启动 PostgreSQL（端口 15436）
make setup-redis  # 启动 Redis（端口 16379）
```

### 2. 配置文件准备

替换 `configs/` 目录下各 yaml 文件中的 `<placeholder>` 占位符：

| 占位符 | 所在文件 | 说明 |
|--------|----------|------|
| `<db-username>` | `configs/database.yaml` | PostgreSQL 用户名 |
| `<db-password>` | `configs/database.yaml` | PostgreSQL 密码 |
| `<redis-password>` | `configs/cache.yaml` | Redis 密码（无密码可留空） |
| `<jwt-secret>` | `configs/security.yaml` | JWT 签名密钥（建议 32 位以上随机字符串） |
| `<init-secret-key>` | `configs/init.yaml` | 应用初始化密钥 |

> 可通过环境变量 `CONFIG_PATH` 指向自定义配置目录，默认使用项目根目录下的 `configs/`。

### 3. 启动开发服务器

```bash
make dev          # 使用 air 热重载（推荐开发时使用）
# 或
go run ./cmd/server
```

### 4. 验证服务

| 地址 | 说明 |
|------|------|
| `http://localhost:8080/api/v1/health` | 健康检查 |
| `http://localhost:8080/swagger/index.html` | Swagger UI |
| `http://localhost:8080/swagger/doc.json` | Swagger JSON |

---

## 配置参考

### `configs/server.yaml` — 服务器与 CORS

```yaml
server:
  port: "8080"           # 监听端口
  mode: release          # gin 模式：debug / release / test
  cors:
    enabled: true
    allowed_origins:     # 允许的前端来源
      - "http://localhost:3000"
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["Accept", "Authorization", "Content-Type"]
    allow_credentials: true
    max_age: 300          # 预检请求缓存时间（秒）
```

### `configs/database.yaml` — 数据库

```yaml
database:
  host: "localhost"
  port: "15436"
  username: "<db-username>"
  password: "<db-password>"
  name: "postgres"       # 数据库名
  sslmode: "disable"
  auto_migrate: true     # 启动时自动执行数据库迁移
  max_open_conns: 10
  max_idle_conns: 5
```

> 也支持直接通过 `url` 字段提供完整 DSN：
> ```yaml
> database:
>   url: "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
> ```

### `configs/cache.yaml` — Redis

```yaml
cache:
  host: "127.0.0.1"
  port: "16379"
  password: ""
  db: 0
```

### `configs/security.yaml` — JWT 与限流

```yaml
security:
  jwt_secret: "<jwt-secret>"
  token_expiry: 24        # Access Token 有效期（小时）
  refresh_expiry: 72      # Refresh Token 有效期（小时）
  password_cost: 12       # bcrypt 加密强度
  enable_rate_limit: true
  rate_limit: 60          # 每分钟最大请求数
  cookie:
    secure: false         # 生产环境应设为 true（需要 HTTPS）
    same_site: "lax"
    path: "/api/v1/auth"
```

### `configs/access_control.yaml` — 多租户与访问控制

```yaml
access_control:
  multi_tenancy:
    enabled: true
    default_tenant_id: ""   # 单租户模式时设置默认租户 ID
  project:
    enabled: true
  domain:
    mode: "domain"          # 域模式
  abac:
    enabled: false          # 基于属性的访问控制（高级功能）
```

### 环境变量覆盖

任意配置项均可通过环境变量覆盖，格式为 `LEEFORGE_<SECTION>_<KEY>`（全大写，以 `_` 分隔）：

```bash
LEEFORGE_DATABASE_HOST=db.example.com
LEEFORGE_DATABASE_PORT=5432
LEEFORGE_SECURITY_JWT_SECRET=mysecret
LEEFORGE_CACHE_HOST=redis.example.com
```

---

## 核心概念

### 三层模型

```
┌─────────────────────────────────────────┐
│              Gin Engine                 │  ← 对外暴露 HTTP 服务
│   engine.Any("/*any", gin.WrapH(...))  │
└──────────────────┬──────────────────────┘
                   │ 所有请求透传
┌──────────────────▼──────────────────────┐
│            Core Runtime                 │  ← 管理认证、插件、路由
│   rt.Handler() → Chi Router            │
└────────┬─────────────────┬─────────────┘
         │                 │
┌────────▼──────┐  ┌───────▼──────────────┐
│   Plugins     │  │      Modules          │
│  (Tenant/OU)  │  │  (Post / 你的业务)    │
└───────────────┘  └──────────────────────┘
```

- **Gin Engine**：仅作为 HTTP 服务器，通过 `gin.WrapH` 将所有请求转发给 Core Runtime。
- **Core Runtime**：负责中间件链（认证、日志、限流）、插件生命周期、Chi 路由挂载。
- **Module**：业务逻辑单元，通过实现 `core.Module` 接口注册到 Runtime。
- **Plugin**：系统级能力扩展（如多租户、OU），在 Runtime 启动阶段初始化。

### 请求生命周期

```
HTTP Request
  → Gin middleware（recover、logger）
  → gin.WrapH → Chi Router
  → Core middleware（JWT 验证、租户解析、限流）
  → Module.RegisterPrivateRoutes / RegisterPublicRoutes
  → Handler → Service → Ent ORM → PostgreSQL
```

### 路由注册

每个 Module 有两类路由注册方法：

```go
// 公开路由：无需认证（如注册、登录、公开内容）
func (m *PostModule) RegisterPublicRoutes(r chi.Router) {}

// 私有路由：经过 JWT 认证和权限校验
func (m *PostModule) RegisterPrivateRoutes(r chi.Router) {
    r.Route("/posts", func(r chi.Router) {
        r.Get("/", m.handler.ListPosts)
        r.Post("/", m.handler.CreatePost)
    })
}
```

最终挂载路径为 `/api/v1/<module-routes>`。

### 获取认证用户

在 Handler 或 Service 中，通过 Context 获取当前登录用户 ID：

```go
import "github.com/leeforge/core"

userID, ok := core.GetUserID(ctx)
if !ok {
    return nil, fmt.Errorf("missing user context")
}
```

---

## 开发新业务模块

以下以新增 `comment`（评论）模块为例，演示完整的垂直切片开发流程。

### 第一步：定义 Ent Schema

在 `ent/schema/` 下新建 `comment.go`：

```go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "github.com/google/uuid"
    frameEntities "github.com/leeforge/framework/entities"
)

type Comment struct{ ent.Schema }

func (Comment) Mixin() []ent.Mixin {
    return []ent.Mixin{
        frameEntities.BaseEntitySchema{}, // 自动注入 id、created_at、updated_at、deleted_at 等公共字段
    }
}

func (Comment) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("post_id", uuid.UUID{}),
        field.UUID("author_id", uuid.UUID{}),
        field.Text("content").NotEmpty(),
    }
}
```

生成 ORM 代码：

```bash
make generate
```

### 第二步：定义 DTO 与错误

新建 `modules/comment/dto.go`：

```go
package comment

import (
    "errors"
    "github.com/google/uuid"
    "time"
)

var (
    ErrCommentNotFound = errors.New("comment not found")
    ErrInvalidComment  = errors.New("invalid comment data")
)

type CreateRequest struct {
    PostID  uuid.UUID `json:"post_id"`
    Content string    `json:"content"`
}

type CommentDTO struct {
    ID        uuid.UUID `json:"id"`
    PostID    uuid.UUID `json:"post_id"`
    AuthorID  uuid.UUID `json:"author_id"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 第三步：实现 Service

新建 `modules/comment/service.go`：

```go
package comment

import (
    "context"
    "fmt"
    "github.com/leeforge/core"
    examplesent "leeforge-example-service/ent"
)

type Service struct{ client *examplesent.Client }

func NewService(client *examplesent.Client) *Service {
    return &Service{client: client}
}

func (s *Service) CreateComment(ctx context.Context, req *CreateRequest) (*CommentDTO, error) {
    authorID, ok := core.GetUserID(ctx)
    if !ok {
        return nil, fmt.Errorf("missing user context")
    }
    c, err := s.client.Comment.Create().
        SetPostID(req.PostID).
        SetAuthorID(authorID).
        SetContent(req.Content).
        Save(ctx)
    if err != nil {
        return nil, fmt.Errorf("create comment: %w", err)
    }
    return &CommentDTO{
        ID: c.ID, PostID: c.PostID,
        AuthorID: c.AuthorID, Content: c.Content, CreatedAt: c.CreatedAt,
    }, nil
}
```

### 第四步：实现 Handler

新建 `modules/comment/handler.go`：

```go
package comment

import (
    "encoding/json"
    "net/http"
    "github.com/leeforge/framework/http/responder"
    "github.com/leeforge/framework/logging"
)

type Handler struct {
    service *Service
    logger  logging.Logger
}

func NewHandler(svc *Service, logger logging.Logger) *Handler {
    return &Handler{service: svc, logger: logger}
}

// CreateComment handles POST /comments
//
// @Summary Create comment
// @Tags Comments
// @Accept json
// @Produce json
// @Param body body CreateRequest true "Comment payload"
// @Success 200 {object} CommentDTO
// @Router /api/v1/comments [post]
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
    var req CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        responder.BindError(w, r, nil)
        return
    }
    result, err := h.service.CreateComment(r.Context(), &req)
    if err != nil {
        responder.DatabaseError(w, r, "Failed to create comment")
        return
    }
    responder.OK(w, r, result)
}
```

### 第五步：实现 Module 并注册路由

新建 `modules/comment/module.go`：

```go
package comment

import (
    "strings"
    "entgo.io/ent/dialect"
    "github.com/go-chi/chi/v5"
    "github.com/leeforge/core/core"
    frameLogging "github.com/leeforge/framework/logging"
    "go.uber.org/zap"
    examplesent "leeforge-example-service/ent"
)

type CommentModule struct{ handler *Handler }

func (m *CommentModule) Name() string { return "comment" }

func (m *CommentModule) RegisterPublicRoutes(_ chi.Router) {}

func (m *CommentModule) RegisterPrivateRoutes(r chi.Router) {
    r.Route("/comments", func(r chi.Router) {
        r.Post("/", m.handler.CreateComment)
    })
}

func NewCommentModule(logger frameLogging.Logger, deps *core.Dependencies) core.Module {
    dsn := deps.Config.Database.DSN()
    driver := resolveDriver(dsn)
    client, err := examplesent.Open(driver, dsn)
    if err != nil {
        logger.Error("failed to open ent client", zap.Error(err))
        return &CommentModule{}
    }
    return &CommentModule{handler: NewHandler(NewService(client), logger)}
}

func resolveDriver(dsn string) string {
    lower := strings.ToLower(strings.TrimSpace(dsn))
    switch {
    case strings.HasPrefix(lower, "postgres://"), strings.HasPrefix(lower, "postgresql://"):
        return dialect.Postgres
    default:
        return dialect.Postgres
    }
}

var _ core.Module = (*CommentModule)(nil)
```

### 第六步：在 bootstrap 中注册

编辑 `bootstrap/app.go`，将新 Module 加入 `ModuleFactories`：

```go
import (
    commentmodule "leeforge-example-service/modules/comment"
    postmodule    "leeforge-example-service/modules/post"
)

func NewApp() (*App, error) {
    return newApp(zap.NewNop(), core.RuntimeOptions{
        ConfigPath:      resolveConfigPath(),
        PluginRegistrar: registerExamplePlugins,
        ModuleFactories: []corecore.ModuleFactory{
            postmodule.NewPostModule,
            commentmodule.NewCommentModule, // ← 新增
        },
    })
}
```

### 第七步：生成 Swagger 文档

```bash
make swagger
```

重启服务后，访问 `http://localhost:8080/swagger/index.html` 可以看到新的 Comments API。

---

## 测试

### 单元测试

```bash
go test ./...
```

### 集成测试（使用 `NewAppForTest`）

`bootstrap.NewAppForTest()` 会跳过外部插件注册和数据库迁移，适合在 CI 或无数据库环境下进行路由和中间件层面的集成测试。

```go
func TestMyHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    app, err := bootstrap.NewAppForTest()
    require.NoError(t, err)

    req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
    w := httptest.NewRecorder()
    app.Engine().ServeHTTP(w, req)
    require.Equal(t, http.StatusOK, w.Code)
}
```

> `NewAppForTest` 使用 `noopResourceProvider`，不会连接真实数据库或 Redis，适合快速冒烟测试。

---

## 部署

### 本地 Docker 联调

```bash
make test
```

启动 `docker/docker-compose.local.yaml` 中定义的完整本地服务栈（含数据库和 Redis）。

### 部署配置文件

部署前需在项目根目录创建以下两个文件：

**`.deploy.env.common`**（所有环境共享）

```env
REGISTRY_HOST=your-registry.example.com
REMOTE_USER=deploy
REMOTE_HOST=your-server.example.com
REMOTE_PORT=22
REMOTE_COMPOSE_PATH=/opt/leeforge
```

**`.deploy.env.<ENV_MODE>`**（按环境区分，如 `.deploy.env.prod`）

```env
# 生产环境特定配置（如不同的端口、路径等）
```

验证配置是否完整：

```bash
make check-config ENV_MODE=prod
```

### 构建镜像

> 注意：Dockerfile 采用多阶段构建，需要从 **Monorepo 根目录** 作为 Docker 构建上下文，以便复制 `core`、`framework`、`plugins` 等本地依赖。

```bash
# 构建 amd64 镜像（推荐，用于部署到 Linux 服务器）
make build VERSION=1.0.0 MONOREPO_ROOT=../

# 构建本机架构镜像（Apple Silicon 开发调试用）
make build-arm VERSION=1.0.0 MONOREPO_ROOT=../
```

### 完整远程部署流程

```bash
make remote-deploy VERSION=1.0.0 ENV_MODE=prod
```

该命令依次执行：

| 步骤 | 操作 |
|------|------|
| 1 | `check-config` — 校验部署配置文件完整性 |
| 2 | `build` — buildx 构建 amd64 镜像 |
| 3 | `tag` — 为镜像打 Registry 标签 |
| 4 | `push` — 推送镜像到私有 Registry |
| 5 | `local-clean` — 清理本地构建镜像 |
| 6 | `push-compose-file` — 上传 Compose 文件到远端服务器 |
| 7 | SSH: `docker compose down` — 停止旧容器 |
| 8 | SSH: `docker compose pull` — 拉取新镜像 |
| 9 | SSH: `docker compose up -d` — 启动新容器 |

### 部署后运维

```bash
# 查看远端容器状态
make remote-status ENV_MODE=prod

# 查看最近 200 行日志
make remote-logs ENV_MODE=prod
```

---

## Makefile 命令速查

### 开发类

| 命令 | 说明 |
|------|------|
| `make dev` | 启动开发服务器（air 热重载） |
| `make generate` | 运行 Ent 代码生成 |
| `make swagger` | 生成 Swagger 文档 |
| `make setup-db` | 启动 PostgreSQL 容器（端口 15436） |
| `make setup-redis` | 启动 Redis 容器（端口 16379） |
| `make clean` | 清理构建产物和 Go 编译缓存 |

### 测试类

| 命令 | 说明 |
|------|------|
| `go test ./...` | 运行所有单元和集成测试 |
| `make test` | 启动本地 Docker Compose 联调环境 |

### 部署类

| 命令 | 说明 |
|------|------|
| `make check-config` | 校验部署配置文件 |
| `make build` | 构建 amd64 Docker 镜像 |
| `make build-arm` | 构建本机架构 Docker 镜像 |
| `make tag` | 为镜像打 Registry 标签 |
| `make push` | 推送镜像到 Registry |
| `make remote-deploy` | 完整远程部署（构建→推送→重启） |
| `make remote-status` | 查看远端容器运行状态 |
| `make remote-logs` | 查看远端容器日志（最近 200 行） |
| `make remote-clean` | 清理远端悬空镜像 |
| `make local-clean` | 清理本地构建镜像 |

> `ENV_MODE` 默认为 `test`，`VERSION` 默认为 `latest`。
> 示例：`make remote-deploy ENV_MODE=prod VERSION=1.2.0`

---

## License

MIT
