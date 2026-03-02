# Templates 开发文档规划 — Brainstorming Doc

**日期**：2026-03-02
**角色**：General Strategist
**状态**：已批准，待实施

---

## 问题陈述与目标

`leeforge/templates` 是一个基于 Leeforge Core 框架的 Gin HTTP 服务模板项目。
当前 README 仅提供了简单的架构示意和快速启动命令，缺乏足够的指引让开发者顺利完成：

- 完整的本地环境搭建
- 添加第一个业务模块
- 理解 Runtime / Module / Plugin 三层模型
- 生产环境部署

**目标**：将 `README.md` 升级为一份完整的开发者指南，受众是**框架使用者**（基于模板搭建自己业务服务的开发者）。

---

## 约束与假设

- 文档写在 `README.md`（单文件），不拆分多文件
- 使用中文编写（与项目团队语言一致）
- 技术栈：Go 1.25+ / Gin / Chi / Ent ORM / PostgreSQL 18 / Redis
- 现有代码结构不变，文档仅描述现状

---

## 候选方案与权衡

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| A（选定）| 单文件 README，全量覆盖 | 易搜索、fork 即用、无跳转 | 文件较长 |
| B | 多文件分章节 | 适合扩展成文档站 | 维护成本高 |
| C | README 入口 + 专项文档 | README 简洁 | 与现有 README 有重叠 |

**选定方案 A**。

---

## 推荐设计：README.md 章节规划

### 1. 项目概览
- 定位说明（这是什么、适合谁用）
- 技术栈一览表
- 带注释的完整目录结构

### 2. 快速上手
- 前置依赖：Go 1.25+、air、swag、Docker
- 启动基础设施：`make setup-db` / `make setup-redis`
- 配置文件准备：占位符替换说明（表格形式）
- 运行开发服务器：`make dev` 或 `go run ./cmd/server`
- 验证：health check URL、Swagger UI 地址

### 3. 配置参考
- 各 yaml 文件对应的配置结构说明
- 关键字段逐一注释（database / security / access_control / plugins）
- `CONFIG_PATH` 环境变量覆盖方式
- 最小可运行配置示例

### 4. 核心概念
- Runtime / Module / Plugin 三层模型示意图（ASCII）
- 请求生命周期：`Gin.Any → Chi Router → Module.Handler`
- 认证上下文：`core.GetUserID(ctx)` 用法说明
- 公开路由 vs 私有路由（`RegisterPublicRoutes` / `RegisterPrivateRoutes`）

### 5. 开发新业务模块
- 垂直切片模式说明
- Step-by-step 教程：
  1. 创建 `modules/<name>/` 目录
  2. 实现 `module.go`（Module 接口）
  3. 编写 `handler.go`（HTTP 处理层）
  4. 编写 `service.go`（业务逻辑层）
  5. 编写 `dto.go`（数据传输对象）
  6. 在 `bootstrap/app.go` 中注册 ModuleFactory
- Ent Schema 定义与代码生成：`make generate`
- Swagger 注解与生成：`make swagger`

### 6. 测试
- 单元测试：`go test ./...`
- 集成测试：`bootstrap.NewAppForTest()` 用法（跳过插件和迁移）
- 测试示例代码说明

### 7. 部署
- 本地 Docker 联调：`make test`
- 镜像构建：`make build`（amd64 buildx）/ `make build-arm`（本机架构）
- 部署配置文件：`.deploy.env.common` + `.deploy.env.<ENV_MODE>` 字段说明
- 完整远程部署流程：`make remote-deploy` 内部步骤拆解
- 运维命令：`make remote-status` / `make remote-logs`

### 8. Makefile 命令速查表
- 全命令列表，含说明和典型用法

---

## 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| README 过长，阅读体验差 | 使用 `<details>` 折叠次要内容（如完整配置字段表） |
| 文档与代码不同步 | 在 README 顶部标注"基于 `main` 分支，如有差异以代码为准" |
| 占位符配置项未来变化 | 配置表格从 `config/config.go` 的 `Default()` 函数推导，变化时同步更新 |

---

## 验证策略

- 找一位未接触过此项目的开发者，仅凭 README 完成本地运行
- 验证所有 `make` 命令与文档描述一致
- 确认 Swagger UI 能正常访问

---

## 开放问题

- Echo 分支是否需要同步一份文档？（暂不覆盖，仅针对 Gin 主分支）
- 是否需要中英双语？（当前计划仅中文）
