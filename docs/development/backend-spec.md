# Backend 技术规范

## 技术栈

| 类别 | 技术选型 | 版本 |
|------|---------|------|
| 语言 | Go | 1.24+ |
| Web 框架 | Gin | v1.10+ |
| ORM | GORM | v1.25+ |
| 数据库 | PostgreSQL | 16+ |
| 缓存 | Redis | 7+ |
| 日志 | Zap | v1.27+ |
| 配置 | env (caarlos0) | v11+ |
| 迁移 | golang-migrate | v4+ |

## 项目结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
├── internal/
│   ├── config/
│   │   └── config.go            # 环境变量配置
│   ├── database/
│   │   └── postgres.go          # PostgreSQL 连接
│   ├── cache/
│   │   └── redis.go             # Redis 连接
│   ├── middleware/
│   │   ├── cors.go              # CORS 中间件
│   │   ├── logger.go            # 请求日志
│   │   ├── recovery.go          # Panic 恢复
│   │   └── requestid.go         # 请求 ID 追踪
│   ├── handlers/
│   │   ├── health.go            # 健康检查
│   │   └── pomodoro.go          # 业务处理器
│   ├── models/
│   │   └── pomodoro.go          # 数据模型
│   ├── repository/
│   │   └── pomodoro.go          # 数据访问层
│   └── router/
│       └── router.go            # 路由配置
├── migrations/
│   ├── 000001_init.up.sql       # 初始迁移
│   └── 000001_init.down.sql     # 回滚迁移
├── Dockerfile
├── go.mod
└── go.sum
```

## 环境变量

| 变量名 | 说明 | 必填 | 默认值 |
|--------|------|------|--------|
| `PORT` | 服务端口 | 否 | 8080 |
| `ENV` | 运行环境 (development/production) | 否 | development |
| `DATABASE_URL` | PostgreSQL 连接串 | **是** | - |
| `REDIS_URL` | Redis 连接串 | **是** | - |
| `LOG_LEVEL` | 日志级别 (debug/info/warn/error) | 否 | info |
| `ALLOWED_ORIGINS` | CORS 允许的源（逗号分隔） | 否 | * |

### 连接串格式

```bash
# PostgreSQL
DATABASE_URL=postgres://user:password@host:5432/database?sslmode=disable

# Redis
REDIS_URL=redis://[:password@]host:6379[/db]
```

## API 规范

### 健康检查

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 基础健康检查（无依赖检查） |
| `/ready` | GET | 就绪检查（检查 DB 和 Redis） |

### Pomodoro API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/pomodoros` | GET | 获取 Pomodoro 列表 |
| `/api/pomodoros` | POST | 创建 Pomodoro |
| `/api/pomodoros/:id` | GET | 获取单个 Pomodoro |
| `/api/pomodoros/:id` | PATCH | 更新 Pomodoro |
| `/api/pomodoros/:id` | DELETE | 删除 Pomodoro |
| `/api/pomodoros/:id/complete` | POST | 标记完成 |

### 统一响应格式

**成功响应**：
```json
{
  "data": { ... },
  "limit": 20,
  "offset": 0
}
```

**错误响应**：
```json
{
  "error": "错误信息",
  "request_id": "uuid"
}
```

## 中间件

### 执行顺序

1. **RequestID** - 生成/提取请求 ID
2. **Logger** - 记录请求日志
3. **Recovery** - Panic 恢复
4. **CORS** - 跨域处理

### 请求追踪

- 所有请求都会生成 `X-Request-ID` header
- 可以在请求中传入 `X-Request-ID`，系统会复用
- 错误响应中包含 `request_id` 字段

## 数据库迁移

使用 `golang-migrate` 管理数据库迁移。

```bash
# 运行迁移
migrate -path ./migrations -database "$DATABASE_URL" up

# 回滚迁移
migrate -path ./migrations -database "$DATABASE_URL" down 1

# 创建新迁移
migrate create -ext sql -dir ./migrations -seq <name>
```

## 本地开发

### 启动依赖服务

```bash
docker-compose up -d db redis
```

### 运行后端

```bash
cd backend
export DATABASE_URL="postgres://vibe:vibe_secret@localhost:5432/vibe_db?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
go run ./cmd/server/main.go
```

### 或使用 Docker Compose 一键启动

```bash
docker-compose up --build
```

## 部署

### Railway 部署

1. 在 Railway 创建项目
2. 连接 GitHub 仓库
3. Railway 会自动检测 `railway.toml` 配置
4. 配置环境变量：
   - Railway 会自动注入 PostgreSQL 的 `DATABASE_URL`
   - Railway 会自动注入 Redis 的 `REDIS_URL`
5. 每次 push 到 main 分支，Railway 自动部署

### 环境变量配置（Railway）

在 Railway Dashboard 中设置：

```
ENV=production
LOG_LEVEL=info
ALLOWED_ORIGINS=https://your-frontend-domain.com
```

### 健康检查

Railway 会自动使用 `/health` 端点进行健康检查。

## 最佳实践

### 代码规范

- 遵循 Go 标准项目布局
- 使用 `internal/` 目录隔离私有包
- Handler 只做请求解析和响应，业务逻辑放在 Service 层（如需扩展）
- Repository 层封装所有数据库操作

### 错误处理

- 使用统一的错误响应格式
- 不暴露内部错误详情到生产环境
- 使用 Zap 记录详细错误日志

### 日志规范

- 生产环境使用 JSON 格式
- 开发环境使用彩色控制台输出
- 所有请求都记录：方法、路径、状态码、耗时、请求 ID

### 安全

- 使用环境变量管理敏感配置
- 启用 CORS 白名单
- 所有 API 返回请求 ID 便于追踪
