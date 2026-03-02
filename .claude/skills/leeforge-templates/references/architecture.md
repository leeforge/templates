# 架构原理

## 目录
- [三层模型](#三层模型)
- [请求生命周期](#请求生命周期)
- [路由注册](#路由注册)
- [认证上下文](#认证上下文)
- [Plugin vs Module](#plugin-vs-module)

---

## 三层模型

```
┌─────────────────────────────────────────┐
│              Gin Engine                 │  ← 对外暴露 HTTP 端口
│   engine.Any("/*any", gin.WrapH(...))  │    所有请求透传给 Core
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│            Core Runtime                 │  ← 管理认证/插件/路由
│   rt.Handler() → Chi Router            │    中间件链在此执行
└────────┬─────────────────┬─────────────┘
         │                 │
┌────────▼──────┐  ┌───────▼──────────────┐
│   Plugins     │  │      Modules          │
│  Tenant / OU  │  │  Post / Order / ...   │
│（系统级能力）  │  │  （业务逻辑单元）      │
└───────────────┘  └──────────────────────┘
```

- **Gin Engine**：仅作 HTTP 服务器，通过 `gin.WrapH` 将所有请求交给 Core Runtime 处理，自身不做路由。
- **Core Runtime**：执行中间件链（JWT 校验、租户解析、限流、日志），管理 Plugin 生命周期，维护 Chi Router。
- **Plugin**：系统级能力（多租户、OU），在 Runtime 启动时初始化一次，提供跨模块共享服务。
- **Module**：业务逻辑单元，通过 `ModuleFactory` 注册，启动时由 Core 调用 `RegisterPrivateRoutes` / `RegisterPublicRoutes`。

---

## 请求生命周期

```
HTTP Request
  │
  ▼ Gin middleware（recover、logger）
  │
  ▼ gin.WrapH → Core Runtime
  │
  ▼ Core middleware chain
  │   ├── 请求日志
  │   ├── JWT 校验（私有路由）
  │   ├── 租户/域解析（写入 ctx）
  │   └── 限流
  │
  ▼ Chi Router 匹配路由
  │
  ▼ Module Handler
  │   ├── 解析请求体 / 路径参数
  │   └── 调用 Service
  │
  ▼ Service
  │   ├── 业务校验
  │   ├── core.GetUserID(ctx) 获取用户
  │   └── Ent ORM 操作数据库
  │
  ▼ HTTP Response（responder 工具函数）
```

---

## 路由注册

每个 Module 实现两个路由注册方法，最终挂载路径为 `/api/v1/<route>`：

```go
// 公开路由：无 JWT 校验（适合健康检查、公开内容）
func (m *OrderModule) RegisterPublicRoutes(r chi.Router) {
    r.Get("/orders/public", m.handler.ListPublicOrders)
}

// 私有路由：经过完整中间件链（JWT + 租户）
func (m *OrderModule) RegisterPrivateRoutes(r chi.Router) {
    r.Route("/orders", func(r chi.Router) {
        r.Get("/", m.handler.ListOrders)
        r.Post("/", m.handler.CreateOrder)
        r.Get("/{id}", m.handler.GetOrder)
        r.Put("/{id}", m.handler.UpdateOrder)
        r.Delete("/{id}", m.handler.DeleteOrder)
    })
}
```

Chi 路由支持中间件嵌套、子路由、参数捕获（`{id}`、`{slug}`）等标准特性。

---

## 认证上下文

Core Runtime 在 JWT 校验通过后，将用户信息写入 `context.Context`，Service 层通过以下方式读取：

```go
import "github.com/leeforge/core"

// 获取当前登录用户 ID（私有路由保证一定存在）
userID, ok := core.GetUserID(ctx)
if !ok {
    return nil, fmt.Errorf("missing user context")
}

// 租户 ID 由 Core 自动注入，Ent Query 会自动过滤 owner_domain_id
// 通常无需手动读取，除非有跨租户逻辑
```

**多租户隔离原理：**
Ent Schema 混入 `BaseEntitySchema` 后，每条记录都有 `owner_domain_id`。Core Runtime 在租户解析中间件中将当前租户 ID 写入 ctx，Ent Client 自动在查询中附加 `WHERE owner_domain_id = ?`。

---

## Plugin vs Module

| | Plugin | Module |
|--|--------|--------|
| **注册方式** | `PluginRegistrar` 函数 | `ModuleFactories` 切片 |
| **生命周期** | Runtime 启动时初始化 | 每次请求路由匹配时调用 Handler |
| **用途** | 系统级跨模块服务（认证、多租户、OU） | 业务逻辑（CRUD、领域操作） |
| **示例** | `TenantPlugin`、`OUPlugin` | `PostModule`、`OrderModule` |
| **开发者需要** | 极少自定义 Plugin | 频繁开发新 Module |

新业务需求几乎都通过新增 Module 实现，Plugin 仅在需要扩展系统级能力时才自定义。
