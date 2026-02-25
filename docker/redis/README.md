# Redis 部署说明

本目录提供一个可通过环境变量配置密码的 Redis 镜像构建方式。

## 环境变量
- `REDIS_PASSWORD`：设置 Redis `requirepass`。不设则无密码。
- `REDIS_MASTER_PASSWORD`：用于主从复制时的 `masterauth`，可与 `REDIS_PASSWORD` 不同；不设则不写入配置。

## 构建镜像
```bash
docker build -t gva-redis:latest deploy/redis
```

## 运行示例
```bash
# 启动并指定客户端访问密码
docker run -d --name gva-redis \
  -p 6379:6379 \
  -e REDIS_PASSWORD=yourStrongPass \
  gva-redis:latest

# 如需主从认证（仅主从模式需要）
docker run -d --name gva-redis \
  -p 6379:6379 \
  -e REDIS_PASSWORD=yourStrongPass \
  -e REDIS_MASTER_PASSWORD=upstreamPass \
  gva-redis:latest
```

## 工作方式
- 镜像内的 `redis-entrypoint.sh` 会用模板 `redis.conf.template` 生成最终配置文件：
  - `__REDIS_REQUIREPASS__` 替换为 `requirepass <REDIS_PASSWORD>`（未设置则为空）。
  - `__REDIS_MASTERAUTH__` 替换为 `masterauth <REDIS_MASTER_PASSWORD>`（未设置则为空）。
- 最后以生成的配置启动 `redis-server`。

## 注意事项
- 如密码包含 `\`、`|`、`&` 等字符已做转义，仍建议避免使用换行等特殊字符。
- 未设置 `REDIS_PASSWORD` 时实例无密码，请确保网络安全或开启防护。***
