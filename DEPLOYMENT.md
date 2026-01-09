# 部署指南

本文档提供前后端完整的部署步骤。

## 目录

- [后端部署 (Railway)](#后端部署-railway)
- [前端部署 (Vercel)](#前端部署-vercel)
- [环境变量配置](#环境变量配置)
- [数据库设置](#数据库设置)
- [域名和 CORS 配置](#域名和-cors-配置)

---

## 后端部署 (Railway)

Railway 是一个现代化的云平台，支持 Go 应用、PostgreSQL 和 Redis。

### 步骤 1: 创建 Railway 账号

1. 访问 https://railway.app/
2. 使用 GitHub 账号登录
3. 授权 Railway 访问你的仓库

### 步骤 2: 创建新项目

1. 点击 "New Project"
2. 选择 "Deploy from GitHub repo"
3. 选择 `vibe-engineering-playbook` 仓库
4. Railway 会自动检测 `railway.toml` 配置

### 步骤 3: 添加 PostgreSQL 数据库

1. 在项目中点击 "New"
2. 选择 "Database" → "Add PostgreSQL"
3. Railway 会自动创建数据库并生成连接字符串
4. 数据库会自动链接到你的服务

### 步骤 4: 添加 Redis

1. 在项目中点击 "New"
2. 选择 "Database" → "Add Redis"
3. Railway 会自动创建 Redis 实例
4. Redis 会自动链接到你的服务

### 步骤 5: 配置环境变量

点击后端服务 → "Variables" → "Raw Editor"，添加以下变量：

```env
# Server
PORT=8080
GIN_MODE=release
ENVIRONMENT=production

# Database (Railway 会自动注入 DATABASE_URL)
# DATABASE_URL 已由 PostgreSQL 服务自动提供

# Redis (Railway 会自动注入 REDIS_URL)
# REDIS_URL 已由 Redis 服务自动提供

# YouTube API
YOUTUBE_API_KEY=你的_youtube_api_key

# Google OAuth 2.0
GOOGLE_CLIENT_ID=你的_client_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=你的_client_secret
GOOGLE_REDIRECT_URL=https://your-frontend.vercel.app/auth/google/callback

# OpenRouter API
OPENROUTER_API_KEY=你的_openrouter_api_key

# CORS
ALLOWED_ORIGINS=https://your-frontend.vercel.app,http://localhost:3000
```

**重要说明:**
- `DATABASE_URL` 和 `REDIS_URL` 由 Railway 自动提供，无需手动配置
- 将 `your-frontend.vercel.app` 替换为你的 Vercel 域名

### 步骤 6: 部署

1. Railway 会自动开始构建
2. 构建过程使用 `backend/Dockerfile`
3. 等待部署完成（通常 2-5 分钟）
4. 获取后端 URL（例如：`https://your-app.railway.app`）

### 步骤 7: 验证部署

访问健康检查端点：
```bash
curl https://your-app.railway.app/health
```

应该返回：
```json
{
  "status": "ok",
  "timestamp": "2026-01-09T16:30:00Z"
}
```

### 步骤 8: 配置自定义域名（可选）

1. 在 Railway 项目中点击 "Settings"
2. 找到 "Domains" 部分
3. 点击 "Generate Domain" 或添加自定义域名
4. 更新 Vercel 的 `NEXT_PUBLIC_API_URL` 环境变量

---

## 前端部署 (Vercel)

### 步骤 1: 导入项目

1. 访问 https://vercel.com/
2. 点击 "Add New..." → "Project"
3. 选择 `vibe-engineering-playbook` 仓库
4. 点击 "Import"

### 步骤 2: 配置项目

Vercel 会自动从 `vercel.json` 读取配置：

- **Framework Preset**: Next.js ✅
- **Root Directory**: `frontend` ✅
- **Build Command**: `cd frontend && npm run build` ✅
- **Output Directory**: `frontend/.next` ✅

点击 "Deploy" 继续。

### 步骤 3: 配置环境变量

在 Vercel 项目设置中添加：

```env
NEXT_PUBLIC_API_URL=https://your-app.railway.app
```

将 `your-app.railway.app` 替换为 Railway 提供的后端 URL。

### 步骤 4: 重新部署

1. 添加环境变量后
2. 点击 "Deployments"
3. 找到最新的部署
4. 点击 "..." → "Redeploy"
5. 等待部署完成

### 步骤 5: 获取前端 URL

部署完成后，Vercel 会提供一个 URL：
```
https://vibe-engineering-playbook-xxx.vercel.app
```

### 步骤 6: 配置自定义域名（可选）

1. 在 Vercel 项目中点击 "Settings" → "Domains"
2. 添加你的自定义域名
3. 按照提示配置 DNS 记录

---

## 环境变量配置

### 获取必要的 API Keys

#### 1. YouTube API Key

1. 访问 https://console.cloud.google.com/
2. 创建新项目或选择现有项目
3. 启用 "YouTube Data API v3"
4. 创建凭据 → API 密钥
5. 复制 API 密钥

#### 2. Google OAuth 2.0

1. 在 Google Cloud Console 中
2. 创建凭据 → OAuth 客户端 ID
3. 应用类型：Web 应用
4. 已授权的重定向 URI：
   ```
   https://your-frontend.vercel.app/auth/google/callback
   http://localhost:3000/auth/google/callback
   ```
5. 复制客户端 ID 和客户端密钥

#### 3. OpenRouter API Key

1. 访问 https://openrouter.ai/
2. 注册账号
3. 前往 https://openrouter.ai/keys
4. 创建新的 API 密钥
5. 充值余额（用于 AI 视频分析）

---

## 数据库设置

### 自动迁移

后端服务启动时会自动运行数据库迁移：
- 读取 `backend/migrations/` 目录
- 自动创建所有必要的表
- 无需手动操作

### 手动迁移（如需要）

如果需要手动运行迁移：

```bash
# 连接到 Railway PostgreSQL
psql $DATABASE_URL

# 运行迁移 SQL
\i migrations/001_create_videos_table.sql
```

---

## 域名和 CORS 配置

### 更新 CORS 设置

部署后，需要在 Railway 中更新 `ALLOWED_ORIGINS`：

```env
ALLOWED_ORIGINS=https://your-frontend.vercel.app,https://your-custom-domain.com,http://localhost:3000
```

### 更新 OAuth 回调 URL

在 Google Cloud Console 中添加生产环境的回调 URL：

```
https://your-frontend.vercel.app/auth/google/callback
```

并更新 Railway 的环境变量：

```env
GOOGLE_REDIRECT_URL=https://your-frontend.vercel.app/auth/google/callback
```

---

## 部署检查清单

### 后端 (Railway)

- [ ] 创建 Railway 项目
- [ ] 添加 PostgreSQL 数据库
- [ ] 添加 Redis 缓存
- [ ] 配置所有环境变量
- [ ] 等待部署完成
- [ ] 测试健康检查端点
- [ ] 测试 API 端点
- [ ] 配置自定义域名（可选）

### 前端 (Vercel)

- [ ] 导入 GitHub 仓库
- [ ] 配置 `NEXT_PUBLIC_API_URL`
- [ ] 部署完成
- [ ] 获取 Vercel URL
- [ ] 配置自定义域名（可选）

### API 配置

- [ ] 获取 YouTube API Key
- [ ] 创建 Google OAuth 客户端
- [ ] 获取 OpenRouter API Key
- [ ] 配置 OAuth 回调 URL
- [ ] 更新 CORS 设置

---

## 故障排除

### 后端无法启动

**检查日志:**
```bash
# 在 Railway 项目中查看 Logs
```

**常见问题:**
- 数据库连接失败 → 检查 `DATABASE_URL`
- Redis 连接失败 → 检查 `REDIS_URL`
- 缺少环境变量 → 检查所有必需的变量

### 前端无法连接后端

**检查:**
1. `NEXT_PUBLIC_API_URL` 是否正确
2. Railway 后端是否正常运行
3. CORS 配置是否包含 Vercel 域名
4. 浏览器控制台是否有错误

### OAuth 授权失败

**检查:**
1. Google OAuth 客户端 ID 是否正确
2. 回调 URL 是否在 Google Cloud Console 中配置
3. `GOOGLE_REDIRECT_URL` 是否与前端 URL 匹配

---

## 监控和维护

### Railway 监控

Railway 提供内置监控：
- CPU 使用率
- 内存使用率
- 网络流量
- 部署历史

### Vercel 监控

Vercel 提供分析功能：
- 页面访问量
- 性能指标
- 错误追踪
- Web Vitals

### 日志查看

**Railway:**
```
Project → Service → Logs
```

**Vercel:**
```
Project → Logs (Functions)
```

---

## 成本估算

### Railway (后端)

- **Starter Plan**: $5/月
  - 包含 $5 使用额度
  - PostgreSQL 数据库
  - Redis 缓存
  - 适合小型项目

### Vercel (前端)

- **Hobby Plan**: 免费
  - 无限部署
  - 100 GB 带宽/月
  - 适合个人项目

### API 成本

- **YouTube API**: 免费
  - 每天 10,000 单位配额
  - 通常足够使用

- **OpenRouter**: 按使用付费
  - Gemini Flash: ~$0.075/1M tokens
  - 建议预充值 $10 起

---

## 下一步

部署完成后：

1. ✅ 测试完整的 YouTube 视频查询流程
2. ✅ 测试 OAuth 授权流程
3. ✅ 测试播放列表和字幕功能
4. ✅ 监控 API 配额使用情况
5. ✅ 设置错误告警（可选）

---

## 获取帮助

遇到问题？

- Railway 文档: https://docs.railway.app/
- Vercel 文档: https://vercel.com/docs
- 项目 Issues: https://github.com/lessthanno/vibe-engineering-playbook/issues

---

**部署愉快！** 🚀
