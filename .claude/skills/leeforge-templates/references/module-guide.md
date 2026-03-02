# 业务模块开发指南

## 目录
- [垂直切片模式](#垂直切片模式)
- [第一步：Ent Schema](#第一步ent-schema)
- [第二步：代码生成](#第二步代码生成)
- [第三步：DTO 与错误定义](#第三步dto-与错误定义)
- [第四步：Service](#第四步service)
- [第五步：Handler](#第五步handler)
- [第六步：Module](#第六步module)
- [第七步：注册到 Bootstrap](#第七步注册到-bootstrap)

## 垂直切片模式

每个业务模块按功能垂直切分，自包含所有层次（Schema → Service → Handler），放在 `modules/<name>/` 目录下。模块通过实现 `core.Module` 接口接入框架，不与其他模块产生直接依赖。

```
modules/<name>/
  module.go    # Module 接口实现，路由注册
  handler.go   # HTTP 处理层（解析请求、返回响应）
  service.go   # 业务逻辑层（操作数据库）
  dto.go       # 请求/响应结构体 + Sentinel 错误
```

---

## 第一步：Ent Schema

路径：`ent/schema/<name>.go`

```go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
    frameEntities "github.com/leeforge/framework/entities"
)

type Order struct{ ent.Schema }

// Mixin 注入公共字段：id、created_at、updated_at、deleted_at、
// published_at、archived_at、owner_domain_id
func (Order) Mixin() []ent.Mixin {
    return []ent.Mixin{
        frameEntities.BaseEntitySchema{},
    }
}

func (Order) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("user_id", uuid.UUID{}).Comment("下单用户ID"),
        field.String("no").NotEmpty().Comment("订单编号"),
        field.Enum("status").
            Values("pending", "paid", "cancelled").
            Default("pending"),
        field.Int("amount").Default(0).Comment("金额（分）"),
    }
}

func (Order) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("owner_domain_id", "no").Unique(),
        index.Fields("user_id"),
    }
}
```

**关键点：**
- 始终混入 `frameEntities.BaseEntitySchema{}`，自动获得 `id`（UUID）、时间戳、软删除字段
- `owner_domain_id` 由 Mixin 提供，用于多租户隔离
- 枚举状态推荐用 `field.Enum`，便于 Ent 类型校验

---

## 第二步：代码生成

```bash
make generate
# 等价于：cd ent && go generate ./...
```

生成后 `ent/` 目录下会出现 `order.go`、`order_create.go` 等文件，**勿手动修改**。

---

## 第三步：DTO 与错误定义

路径：`modules/<name>/dto.go`

```go
package order

import (
    "errors"
    "github.com/google/uuid"
    "time"
)

// Sentinel 错误，在 Handler 的 mapError 中映射为 HTTP 状态码
var (
    ErrOrderNotFound = errors.New("order not found")
    ErrInvalidOrder  = errors.New("invalid order data")
    ErrOrderNoExists = errors.New("order no already exists")
)

// CreateRequest 创建订单的请求体
type CreateRequest struct {
    No     string `json:"no"`
    Amount int    `json:"amount"`
}

// UpdateRequest 更新订单的请求体（字段全部可选）
type UpdateRequest struct {
    Status string `json:"status,omitempty"`
    Amount int    `json:"amount,omitempty"`
}

// OrderDTO 返回给客户端的订单数据
type OrderDTO struct {
    ID        uuid.UUID `json:"id"`
    No        string    `json:"no"`
    UserID    uuid.UUID `json:"userId"`
    Status    string    `json:"status"`
    Amount    int       `json:"amount"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// ListFilters 列表查询参数
type ListFilters struct {
    Page     int    `json:"page,omitempty"`
    PageSize int    `json:"pageSize,omitempty"`
    Status   string `json:"status,omitempty"`
}

// ListResult 分页列表响应
type ListResult struct {
    Orders     []*OrderDTO `json:"orders"`
    Total      int         `json:"total"`
    Page       int         `json:"page"`
    PageSize   int         `json:"pageSize"`
    TotalPages int         `json:"totalPages"`
}
```

---

## 第四步：Service

路径：`modules/<name>/service.go`

```go
package order

import (
    "context"
    "fmt"
    "strings"

    "github.com/google/uuid"
    "github.com/leeforge/core"
    examplesent "leeforge-example-service/ent"
    entorder "leeforge-example-service/ent/order"
)

type Service struct{ client *examplesent.Client }

func NewService(client *examplesent.Client) *Service {
    return &Service{client: client}
}

func (s *Service) CreateOrder(ctx context.Context, req *CreateRequest) (*OrderDTO, error) {
    if strings.TrimSpace(req.No) == "" {
        return nil, ErrInvalidOrder
    }
    // 从认证上下文中获取当前用户 ID
    userID, ok := core.GetUserID(ctx)
    if !ok {
        return nil, fmt.Errorf("missing user context")
    }

    o, err := s.client.Order.Create().
        SetNo(req.No).
        SetUserID(userID).
        SetAmount(req.Amount).
        Save(ctx)
    if err != nil {
        if examplesent.IsConstraintError(err) {
            return nil, ErrOrderNoExists
        }
        return nil, fmt.Errorf("create order: %w", err)
    }
    return toDTO(o), nil
}

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (*OrderDTO, error) {
    o, err := s.client.Order.Get(ctx, id)
    if err != nil {
        if examplesent.IsNotFound(err) {
            return nil, ErrOrderNotFound
        }
        return nil, fmt.Errorf("get order: %w", err)
    }
    if !o.DeletedAt.IsZero() {
        return nil, ErrOrderNotFound
    }
    return toDTO(o), nil
}

// 软删除：设置 deleted_at 而非物理删除
func (s *Service) DeleteOrder(ctx context.Context, id uuid.UUID) error {
    _, err := s.GetOrder(ctx, id)
    if err != nil {
        return err
    }
    if _, err := s.client.Order.UpdateOneID(id).
        SetDeletedAt(timeNow()).Save(ctx); err != nil {
        return fmt.Errorf("delete order: %w", err)
    }
    return nil
}

func toDTO(o *examplesent.Order) *OrderDTO {
    return &OrderDTO{
        ID:        o.ID,
        No:        o.No,
        UserID:    o.UserID,
        Status:    string(o.Status),
        Amount:    o.Amount,
        CreatedAt: o.CreatedAt,
        UpdatedAt: o.UpdatedAt,
    }
}
```

**关键点：**
- 用 `core.GetUserID(ctx)` 获取认证用户，而非从请求体接收
- 用 `examplesent.IsConstraintError(err)` 捕获唯一索引冲突
- 用 `examplesent.IsNotFound(err)` 判断记录不存在
- 软删除通过设置 `DeletedAt` 时间戳实现

---

## 第五步：Handler

路径：`modules/<name>/handler.go`

```go
package order

import (
    "encoding/json"
    "errors"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "github.com/leeforge/core/server/httplog"
    "github.com/leeforge/framework/http/responder"
    "github.com/leeforge/framework/logging"
)

type Handler struct {
    service *Service
    logger  logging.Logger
}

func NewHandler(service *Service, logger logging.Logger) *Handler {
    return &Handler{service: service, logger: logger}
}

// CreateOrder handles POST /orders
//
// @Summary Create order
// @Tags Orders
// @Accept json
// @Produce json
// @Param body body CreateRequest true "Order payload"
// @Success 200 {object} OrderDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders [post]
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        responder.BindError(w, r, nil)
        return
    }
    result, err := h.service.CreateOrder(r.Context(), &req)
    if err != nil {
        h.mapError(w, r, "Failed to create order", err)
        return
    }
    responder.OK(w, r, result)
}

// GetOrder handles GET /orders/{id}
//
// @Summary Get order by ID
// @Tags Orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} OrderDTO
// @Router /api/v1/orders/{id} [get]
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        responder.BadRequest(w, r, "Invalid order ID")
        return
    }
    result, err := h.service.GetOrder(r.Context(), id)
    if err != nil {
        h.mapError(w, r, "Failed to get order", err)
        return
    }
    responder.OK(w, r, result)
}

// mapError 将 Sentinel 错误映射为 HTTP 响应
func (h *Handler) mapError(w http.ResponseWriter, r *http.Request, msg string, err error) {
    switch {
    case errors.Is(err, ErrOrderNotFound):
        responder.NotFound(w, r, "Order not found")
    case errors.Is(err, ErrOrderNoExists):
        responder.Conflict(w, r, "Order no already exists")
    case errors.Is(err, ErrInvalidOrder):
        responder.BadRequest(w, r, "Invalid order data")
    default:
        httplog.Error(h.logger, r, msg, err)
        responder.DatabaseError(w, r, msg)
    }
}
```

**Swagger 注解要点：**
- `@Tags` 与模块名一致，便于 Swagger UI 分组
- `@Router` 路径前缀固定为 `/api/v1/`
- 错误响应统一用 `map[string]interface{}`

**responder 工具函数速查：**

| 函数 | HTTP 状态码 | 用途 |
|------|------------|------|
| `responder.OK(w, r, data)` | 200 | 正常返回 |
| `responder.BadRequest(w, r, msg)` | 400 | 请求参数错误 |
| `responder.NotFound(w, r, msg)` | 404 | 资源不存在 |
| `responder.Conflict(w, r, msg)` | 409 | 数据冲突 |
| `responder.BindError(w, r, nil)` | 400 | JSON 解析失败 |
| `responder.DatabaseError(w, r, msg)` | 500 | 数据库操作失败 |

---

## 第六步：Module

路径：`modules/<name>/module.go`

```go
package order

import (
    "strings"

    "entgo.io/ent/dialect"
    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"

    "github.com/leeforge/core/core"
    frameLogging "github.com/leeforge/framework/logging"
    examplesent "leeforge-example-service/ent"
)

type OrderModule struct{ handler *Handler }

func (m *OrderModule) Name() string { return "order" }

// RegisterPublicRoutes 注册无需认证的路由（可为空）
func (m *OrderModule) RegisterPublicRoutes(_ chi.Router) {}

// RegisterPrivateRoutes 注册需要 JWT 认证的路由
func (m *OrderModule) RegisterPrivateRoutes(r chi.Router) {
    r.Route("/orders", func(r chi.Router) {
        r.Get("/", m.handler.ListOrders)
        r.Post("/", m.handler.CreateOrder)
        r.Get("/{id}", m.handler.GetOrder)
        r.Put("/{id}", m.handler.UpdateOrder)
        r.Delete("/{id}", m.handler.DeleteOrder)
    })
}

// NewOrderModule 是 core.ModuleFactory，由 bootstrap 调用
func NewOrderModule(logger frameLogging.Logger, deps *core.Dependencies) core.Module {
    dsn := deps.Config.Database.DSN()
    client, err := examplesent.Open(resolveDriver(dsn), dsn)
    if err != nil {
        logger.Error("failed to open ent client for order module", zap.Error(err))
        return &OrderModule{}
    }
    return &OrderModule{handler: NewHandler(NewService(client), logger)}
}

func resolveDriver(dsn string) string {
    lower := strings.ToLower(strings.TrimSpace(dsn))
    switch {
    case strings.HasPrefix(lower, "postgres://"), strings.HasPrefix(lower, "postgresql://"):
        return dialect.Postgres
    case strings.HasPrefix(lower, "sqlite://"), strings.HasPrefix(lower, "file:"):
        return dialect.SQLite
    default:
        return dialect.Postgres
    }
}

var _ core.Module = (*OrderModule)(nil)
```

**关键点：**
- `NewOrderModule` 签名必须是 `func(frameLogging.Logger, *core.Dependencies) core.Module`
- 每个模块自己创建独立的 Ent Client，数据库连接从 `deps.Config.Database.DSN()` 获取
- `var _ core.Module = (*OrderModule)(nil)` 编译期接口校验

---

## 第七步：注册到 Bootstrap

编辑 `bootstrap/app.go`：

```go
import (
    ordermodule   "leeforge-example-service/modules/order"
    postmodule    "leeforge-example-service/modules/post"
)

func NewApp() (*App, error) {
    return newApp(zap.NewNop(), core.RuntimeOptions{
        ConfigPath:      resolveConfigPath(),
        PluginRegistrar: registerExamplePlugins,
        ModuleFactories: []corecore.ModuleFactory{
            postmodule.NewPostModule,
            ordermodule.NewOrderModule,  // ← 新增
        },
    })
}
```

完成后：

```bash
make swagger   # 更新 Swagger 文档
make dev       # 启动服务，访问 http://localhost:8080/swagger/index.html 验证
```
