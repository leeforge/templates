# 部署指南

## 目录
- [本地 Docker 联调](#本地-docker-联调)
- [部署配置文件](#部署配置文件)
- [镜像构建](#镜像构建)
- [完整远程部署流程](#完整远程部署流程)
- [运维命令](#运维命令)

---

## 本地 Docker 联调

```bash
make test
# 等价于：docker compose -f docker/docker-compose.local.yaml up --build
```

启动完整本地服务栈（含 PostgreSQL 和 Redis），用于功能验证。

---

## 部署配置文件

在项目根目录创建以下两个文件（已加入 `.gitignore`，**勿提交**）：

### `.deploy.env.common`（所有环境共享）

```env
REGISTRY_HOST=your-registry.example.com   # 私有镜像仓库地址
REMOTE_USER=deploy                          # SSH 登录用户名
REMOTE_HOST=your-server.example.com        # 目标服务器 IP 或域名
REMOTE_PORT=22                              # SSH 端口
REMOTE_COMPOSE_PATH=/opt/leeforge          # 服务器上 Compose 文件存放路径
```

### `.deploy.env.<ENV_MODE>`（如 `.deploy.env.prod`）

```env
# 生产环境特定变量（如有差异可在此覆盖 common 中的值）
```

### 验证配置完整性

```bash
make check-config ENV_MODE=prod
```

`check-config` 会校验所有必填变量（`REGISTRY_HOST`、`REMOTE_USER`、`REMOTE_HOST`、`REMOTE_PORT`、`REMOTE_COMPOSE_PATH`）是否存在，以及 `REMOTE_PORT` 是否为有效端口号（1-65535）。

---

## 镜像构建

> **重要**：Dockerfile 采用多阶段构建，需要从 **Monorepo 根目录**作为构建上下文，以便复制 `core`、`framework`、`plugins` 本地依赖。

```bash
# 构建 amd64 镜像（部署到 Linux 服务器，推荐）
make build VERSION=1.0.0 MONOREPO_ROOT=../

# 构建本机架构镜像（Apple Silicon 开发调试）
make build-arm VERSION=1.0.0 MONOREPO_ROOT=../

# 保存镜像为 tar 包（离线传输）
make save VERSION=1.0.0 MONOREPO_ROOT=../
```

**Makefile 变量说明：**

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `VERSION` | `latest` | 镜像标签 |
| `ENV_MODE` | `test` | 环境模式（test / prod） |
| `MONOREPO_ROOT` | `.` | Monorepo 根目录路径 |
| `APP_NAME` | `leeforge-examples` | 镜像名 |

---

## 完整远程部署流程

```bash
make remote-deploy VERSION=1.2.0 ENV_MODE=prod MONOREPO_ROOT=../
```

该命令依次执行以下步骤：

| 步骤 | 命令 | 说明 |
|------|------|------|
| 1 | `check-config` | 校验 `.deploy.env.*` 配置文件完整性 |
| 2 | `build` | buildx 构建 amd64 Docker 镜像 |
| 3 | `tag` | 为镜像打 `<REGISTRY_HOST>/<APP_NAME>:<VERSION>` 标签 |
| 4 | `push` | 推送镜像到私有 Registry |
| 5 | `local-clean` | 删除本地构建镜像，释放磁盘空间 |
| 6 | `push-compose-file` | SSH 上传 `docker/docker-compose.<ENV_MODE>.yaml` 到服务器 |
| 7 | SSH `docker compose down` | 停止并移除旧容器 |
| 8 | SSH `docker compose pull` | 从 Registry 拉取新镜像 |
| 9 | SSH `docker compose up -d` | 后台启动新容器 |

### Docker Compose 环境变量（服务器端）

Compose 文件通过 `env_file` 加载服务器上的环境变量文件：

```yaml
env_file:
  - ./env/.env               # 通用密钥（jwt_secret、db 密码等）
  - ./env/leeforge_examples.env  # 应用特定配置
```

这些文件需要在服务器上手动创建并维护，不通过 Makefile 上传。

### Compose 网络依赖

`docker-compose.test.yaml` 中使用了外部网络：

```yaml
networks:
  redis-network:
    external: true   # 需要提前在服务器上创建：docker network create redis-network
  pgsql_network:
    external: true   # 需要提前在服务器上创建：docker network create pgsql_network
```

首次部署前确保这些外部网络已存在。

---

## 运维命令

```bash
# 查看远端容器运行状态
make remote-status ENV_MODE=prod

# 查看最近 200 行日志
make remote-logs ENV_MODE=prod

# 清理远端悬空镜像（释放磁盘）
make remote-clean ENV_MODE=prod

# 清理本地构建镜像
make local-clean VERSION=1.0.0
```

---

## Makefile 完整命令速查

### 开发

| 命令 | 说明 |
|------|------|
| `make dev` | 启动开发服务器（air 热重载） |
| `make generate` | Ent ORM 代码生成 |
| `make swagger` | 生成 Swagger 文档 |
| `make setup-db` | 启动 PostgreSQL（端口 15436） |
| `make setup-redis` | 启动 Redis（端口 16379） |
| `make clean` | 清理构建产物和 Go 编译缓存 |

### 部署

| 命令 | 说明 |
|------|------|
| `make check-config` | 校验部署配置 |
| `make build` | 构建 amd64 镜像（buildx） |
| `make build-arm` | 构建本机架构镜像 |
| `make tag` | 打 Registry 标签 |
| `make push` | 推送到 Registry |
| `make remote-deploy` | 完整远程部署 |
| `make remote-status` | 查看容器状态 |
| `make remote-logs` | 查看远端日志 |
| `make remote-clean` | 清理远端悬空镜像 |
| `make local-clean` | 清理本地镜像 |
