# 配置参考

所有配置文件位于 `configs/` 目录，格式为 YAML。支持通过环境变量覆盖：`LEEFORGE_<SECTION>_<KEY>`（全大写，`.` 替换为 `_`）。

也可通过环境变量 `CONFIG_PATH` 指定自定义配置目录，默认为项目根目录的 `configs/`。

---

## server.yaml — 服务器与 CORS

```yaml
server:
  port: "8080"       # 监听端口
  mode: release      # gin 模式：debug / release / test
  cors:
    enabled: true
    allowed_origins:
      - "http://localhost:3000"   # 允许的前端来源（生产需改为实际域名）
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["Accept", "Authorization", "Content-Type"]
    allow_credentials: true
    max_age: 300     # 预检请求缓存时间（秒）
```

---

## database.yaml — 数据库

```yaml
database:
  # 方式一：分字段配置
  host: "localhost"
  port: "15436"
  username: "<db-username>"
  password: "<db-password>"
  name: "postgres"       # 数据库名
  sslmode: "disable"     # 生产环境建议 require
  params: ""             # 额外 DSN 参数，如 connect_timeout=10
  auto_migrate: true     # 启动时自动执行 Ent 迁移
  max_open_conns: 10
  max_idle_conns: 5

  # 方式二：直接指定 DSN（优先级更高）
  # url: "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
```

**环境变量示例：**
```bash
LEEFORGE_DATABASE_HOST=db.prod.example.com
LEEFORGE_DATABASE_PORT=5432
LEEFORGE_DATABASE_NAME=myapp
```

---

## cache.yaml — Redis

```yaml
cache:
  host: "127.0.0.1"
  port: "16379"
  password: ""   # 无密码可留空
  db: 0
```

---

## security.yaml — JWT 与限流

```yaml
security:
  jwt_secret: "<jwt-secret>"    # 建议 32 位以上随机字符串，生产必填
  token_expiry: 24              # Access Token 有效期（小时）
  refresh_expiry: 72            # Refresh Token 有效期（小时）
  password_cost: 12             # bcrypt 加密强度（10-14，越高越安全越慢）
  enable_rate_limit: true
  rate_limit: 60                # 每分钟最大请求数（per IP）
  cookie:
    secure: false               # 生产环境必须设为 true（需要 HTTPS）
    same_site: "lax"            # strict / lax / none
    domain: ""                  # 留空则使用请求域名
    path: "/api/v1/auth"
```

---

## access_control.yaml — 多租户与访问控制

```yaml
access_control:
  multi_tenancy:
    enabled: true
    default_tenant_id: ""     # 单租户模式：设置固定的租户 UUID
  project:
    enabled: true
    default_project_id: ""
    require_project: false    # true 则所有请求必须携带项目 ID
  domain:
    mode: "domain"            # 域隔离模式
    default_domain_id: ""
  abac:
    enabled: false            # 基于属性的访问控制（扩展功能）
  share:
    enabled: false
  quota:
    enabled: false
  audit:
    enabled: false
```

**常见场景：**
- **SaaS 多租户**：`multi_tenancy.enabled: true`，每个租户数据自动隔离
- **单租户/内部系统**：`multi_tenancy.enabled: true` + `default_tenant_id: <固定UUID>`

---

## log.yaml — 日志

```yaml
log:
  director: "logs"            # 日志文件目录
  level: "info"               # debug / info / warn / error
  format: "console"           # console（人类可读）/ json（结构化，适合生产）
  log-in-terminal: true       # 同时输出到终端
  show-line-number: true
  time-format: "2006/01/02 - 15:04:05"
  max-age: 7                  # 日志文件保留天数
  max-size: 100               # 单文件最大大小（MB）
  max-backups: 10             # 保留的备份文件数量
  compress: true              # 压缩旧日志文件
```

---

## tracing.yaml — 链路追踪

```yaml
tracing:
  enabled: false
  endpoint: "localhost:4317"  # OpenTelemetry Collector gRPC 地址
```

---

## metrics.yaml — Prometheus 指标

```yaml
metrics:
  port: "9090"   # Prometheus 指标暴露端口
```

---

## init.yaml — 应用初始化

```yaml
init:
  secret_key: "<init-secret-key>"   # 初始化接口的鉴权密钥，首次部署后建议修改
```

---

## captcha.yaml — 验证码

```yaml
captcha:
  enabled: false
  ttl: "5m"
  generate_limit: 10
  generate_window: "1m"
  max_attempts: 5
  attempt_window: "5m"
  math:
    width: 240
    height: 80
    noise_count: 5
```

---

## 占位符速查

| 占位符 | 所在文件 | 说明 |
|--------|----------|------|
| `<db-username>` | `database.yaml` | PostgreSQL 用户名 |
| `<db-password>` | `database.yaml` | PostgreSQL 密码 |
| `<redis-password>` | `cache.yaml` | Redis 密码 |
| `<jwt-secret>` | `security.yaml` | JWT 签名密钥 |
| `<init-secret-key>` | `init.yaml` | 初始化密钥 |
