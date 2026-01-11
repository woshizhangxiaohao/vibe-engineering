# 本地开发指南

## 本地前端测试本地 Go 后端

### 方式一：本地运行前端 + 本地运行 Go 后端（推荐）

#### 1. 启动后端服务

**选项 A：使用 Docker Compose（推荐）**
```bash
# 只启动数据库和 Redis，不启动前端
docker-compose -f docker-compose.dev.yml up db redis backend
```

**选项 B：本地直接运行 Go**
```bash
cd backend
go run cmd/server/main.go
```

后端服务将运行在 `http://localhost:8080`

#### 2. 启动前端开发服务器

```bash
cd frontend

# 安装依赖（如果还没有）
npm install

# 启动开发服务器
npm run dev
```

前端将运行在 `http://localhost:3000`，并自动连接到 `http://localhost:8080` 的后端 API。

#### 3. 配置说明

**环境变量配置**

后端需要以下环境变量（通过 `.env` 文件或系统环境变量设置）：

```env
# OpenRouter API Key（必需，用于视频分析功能）
OPENROUTER_API_KEY=your_api_key_here
```

**方式 A：使用 `.env` 文件（推荐）**

在项目根目录创建 `.env` 文件：
```env
OPENROUTER_API_KEY=your_api_key_here
```

Docker Compose 会自动读取 `.env` 文件中的环境变量。

**方式 B：使用系统环境变量**

```bash
export OPENROUTER_API_KEY=your_api_key_here
docker-compose -f docker-compose.dev.yml up db redis backend
```

**前端配置**

前端默认配置（`frontend/lib/api/config.ts`）：
- API 基础 URL: `http://localhost:8080`（默认值）
- 可通过环境变量 `NEXT_PUBLIC_API_URL` 覆盖

如果需要修改 API 地址，创建 `frontend/.env.local`：
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 方式二：使用 Docker Compose（前端和后端都在 Docker 中）

```bash
# 启动所有服务（包括前端）
docker-compose -f docker-compose.dev.yml up
```

前端：`http://localhost:3000`
后端：`http://localhost:8080`

### 方式三：混合模式（后端 Docker + 前端本地）

```bash
# 1. 启动后端相关服务（数据库、Redis、后端）
docker-compose -f docker-compose.dev.yml up db redis backend

# 2. 在另一个终端，本地运行前端
cd frontend
npm run dev
```

## 生产环境

生产环境（`docker-compose.yml`）**不包含前端服务**，只运行后端、数据库和 Redis。

## 常见问题

### 前端无法连接到后端

1. 确认后端服务正在运行：访问 `http://localhost:8080/health`
2. 检查 CORS 配置：后端需要允许 `http://localhost:3000` 的请求
3. 检查环境变量：确认 `NEXT_PUBLIC_API_URL` 设置正确

### 端口冲突

- 后端默认端口：`8080`
- 前端默认端口：`3000`
- 数据库端口：`5432`
- Redis 端口：`6379`

如需修改端口，请更新相应的配置文件。

