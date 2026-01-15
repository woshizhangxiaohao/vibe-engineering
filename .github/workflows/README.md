# GitHub Actions 工作流文档

本文档详细说明了项目中所有 GitHub Actions 工作流的功能、触发条件和使用方法。

## 📋 目录

- [AI Agent 工作流](#ai-agent-工作流)
- [路由和管理工作流](#路由和管理工作流)
- [自动化工作流](#自动化工作流)
- [监控和错误处理](#监控和错误处理)
- [其他工作流](#其他工作流)
- [使用指南](#使用指南)

---

## AI Agent 工作流

### 1. Simple Task Agent (`agent-simple.yml`)

**功能**: 处理简单任务，直接实现代码，无需复杂分析。

**触发方式**:

- Issue 评论中包含 `/agent-simple`
- 被 `vibe-router.yml` 自动触发（复杂度为 S）

**特点**:

- 直接开始编码，不进行需求分析
- 适合单文件修改、bug 修复、样式调整
- 最大 30 轮对话
- 使用 `claude-sonnet-4` 模型

**使用场景**:

- 修复简单的 bug
- 调整 UI 样式
- 修改文案
- 添加单个 UI 元素

---

### 2. Medium Task Agent (`agent-medium.yml`)

**功能**: 处理中等复杂度任务，先分析再实现。

**触发方式**:

- Issue 评论中包含 `/agent-medium`
- 被 `vibe-router.yml` 自动触发（复杂度为 M）

**特点**:

- 两阶段处理：先分析需求，再开发实现
- 适合涉及 2-5 个文件的新功能
- 最大 50 轮对话
- 使用 `claude-sonnet-4` 模型

**使用场景**:

- 新增独立功能模块
- 需要前后端都改但逻辑简单
- 新增 API endpoint + 简单 UI

---

### 3. Complex Task Agent (`agent-complex.yml`)

**功能**: 处理复杂任务，自动拆分为多个子 Issue。

**触发方式**:

- Issue 评论中包含 `/agent-complex`
- 被 `vibe-router.yml` 自动触发（复杂度为 L）

**特点**:

- 使用 AI 分析需求并拆分子任务
- 自动创建子 Issue 并设置依赖关系
- 自动触发第一个无依赖的子任务
- 使用 `claude-sonnet-4` 模型

**使用场景**:

- 涉及多个模块的大型功能
- 需要数据库 schema 变更
- 需要架构设计或重构
- 涉及第三方服务集成

**输出**:

- 创建 3-8 个子 Issue
- 每个子 Issue 包含任务描述、验收标准、预估工时
- 自动设置依赖关系和优先级

---

### 4. UI Agent (`agent-ui.yml`)

**功能**: 生成 UI 设计规格文档。

**触发方式**:

- Issue 评论中包含 `/agent-ui` 或 `/agent ui`

**特点**:

- 两步处理：
  1. **PM Compiler**: 将 PRD 转换为清晰的产品需求规格
  2. **UI Spec Generator**: 生成 UI 设计规格
- 严格遵循 Base.org 设计系统规范
- 优先使用 shadcn/ui 组件

**输出格式**:

- 功能概述
- 页面布局
- 组件设计（使用哪些 shadcn/ui 组件）
- 交互流程
- 响应式设计
- 样式规范

**使用场景**:

- 需要先设计 UI 再开发
- 产品需求不够明确，需要转换为技术规格

---

### 5. Backend Agent (`backend-agent.yml`)

**功能**: 基于 UI Spec 生成后端 API 代码。

**触发方式**:

- Issue 评论中包含 `/agent-be [UI Spec URL]`

**特点**:

- 自动推导 API Contract（使用 AI）
- 读取项目现有代码结构
- 生成完整的后端实现（handler/service/repository）
- 支持功能分支（`feature:xxx` 标签）
- 最大 60 轮对话
- 使用 `claude-sonnet-4.5` 模型

**使用场景**:

- 已有 UI Spec，需要实现后端 API
- 前后端分离开发

**参数**:

- 可选：UI Spec URL（Issue 或 Comment URL）
- 可选：用户额外指令

**示例**:

```
/agent-be https://github.com/owner/repo/issues/123
/agent-be https://github.com/owner/repo/issues/123#issuecomment-456 使用 PostgreSQL
```

---

### 6. Frontend Agent (`frontend-agent.yml`)

**功能**: 基于 UI Spec 生成前端代码。

**触发方式**:

- Issue 评论中包含 `/agent-fe [UI Spec URL]`
- 被 `auto-trigger-frontend.yml` 自动触发

**特点**:

- 读取项目前端代码结构
- 严格遵循 STYLE_GUIDE.md 和 Base.org 设计系统
- 使用 shadcn/ui 组件
- 支持功能分支（`feature:xxx` 标签）
- 最大 60 轮对话
- 使用 `claude-sonnet-4.5` 模型

**使用场景**:

- 已有 UI Spec，需要实现前端页面
- 后端 API 已完成，需要实现前端调用

**参数**:

- 可选：UI Spec URL（Issue 或 Comment URL）
- 可选：用户额外指令

---

## 路由和管理工作流

### 7. Vibe Router (`vibe-router.yml`)

**功能**: 自动分析 Issue 复杂度并路由到对应的 Agent。

**触发方式**:

- Issue 创建时自动触发

**特点**:

- 使用 AI 分析需求复杂度（S/M/L/skip）
- 自动添加复杂度标签和影响范围标签
- 自动触发对应的 Agent（Simple/Medium/Complex）
- 跳过特殊类型的 Issue（PRD、Auto、sub-issue 等）

**复杂度判断标准**:

- **S (简单)**: 单文件修改，< 2 小时工作量
- **M (中等)**: 2-5 个文件，2-8 小时工作量
- **L (复杂)**: > 5 个文件，> 1 天工作量
- **skip**: 非开发任务（讨论、文档等）

**输出**:

- 添加复杂度标签：`complexity:simple` / `complexity:medium` / `complexity:complex`
- 添加影响范围标签：`frontend` / `backend` / `database`
- 发布分析结果评论
- 自动触发对应的 Agent

---

### 8. Issue Router (`issue-router.yml`)

**功能**: 处理评论过多的 Issue，自动创建子 Issue 避免折叠。

**触发方式**:

- Issue 评论中包含 Agent 命令（`/agent-ui`, `/agent-be`, `/agent-fe`）
- 支持 `-new-` 前缀强制创建新 Issue

**特点**:

- 当评论数超过阈值（默认 8 条）时自动创建子 Issue
- 支持 `-new-` 命令强制创建新 Issue
- 自动查找根父 Issue（避免多层嵌套）
- 在新 Issue 中自动触发对应的 Agent

**命令**:

- `/agent-ui` / `/agent-be` / `/agent-fe` - 在当前 Issue 处理
- `/agent-new-ui` / `/agent-new-be` / `/agent-new-fe` - 强制创建新 Issue

**配置**:

- `COMMENT_THRESHOLD`: 评论数量阈值（默认 8）

---

### 9. Issue Manager (`issue-manager.yml`)

**功能**: 自动管理 Issue，包括标签和欢迎消息。

**触发方式**:

- Issue 创建时自动触发
- Issue 评论中包含 `/clean-stale` 命令

**功能**:

1. **自动标签**:
   - 根据标题和内容自动识别类型（frontend/backend/bug/enhancement）
2. **欢迎消息**:
   - 显示可用的 Agent 命令
   - 说明 Vibe Router 会自动处理
3. **清理超时任务**:
   - `/clean-stale` 命令清理超过 24 小时未更新的 processing 标签

---

## 自动化工作流

### 10. Auto Trigger Frontend (`auto-trigger-frontend.yml`)

**功能**: 后端 PR 合并后自动触发前端开发。

**触发方式**:

- PR 合并时自动触发（仅限后端 Agent 创建的 PR）

**特点**:

- 检测 PR 标题是否包含 "Backend:"
- 从 PR 标题提取关联的 Issue 编号
- 检查 Issue 是否包含前端任务
- 如果包含，自动在 Issue 中评论 `/agent-fe` 触发前端 Agent

**使用场景**:

- 前后端分离开发
- 后端完成后自动开始前端开发

---

### 11. Auto Fix CI Failures (`auto-fix-CI-failures.yml`)

**功能**: 自动修复 CI 构建失败。

**触发方式**:

- CI workflow 失败时自动触发

**特点**:

- 获取失败的 CI 日志
- 使用 Claude Code Action 自动修复
- 创建修复分支（`claude-auto-fix-ci-*`）
- 避免循环触发（不处理自己创建的分支）

**限制**:

- 仅处理有关联 PR 的 CI 失败
- 不处理修复分支的 CI 失败

---

### 12. Feature Branch Manager (`feature-branch-manager.yml`)

**功能**: 管理功能分支，支持自动创建、同步和合并。

**触发方式**:

- Issue 被打上 `feature:xxx` 标签时自动创建分支
- Issue 评论中包含 `/sync` 命令
- Issue 评论中包含 `/merge-to-main` 命令

**功能**:

1. **自动创建功能分支**:
   - 标签格式：`feature:insightflow` → 分支：`feature/insightflow`
   - 基于 main 分支创建
2. **同步 main 到功能分支**:
   - `/sync` 命令将 main 最新代码合并到功能分支
   - 如果存在冲突，会提示手动解决
3. **创建合并 PR**:
   - `/merge-to-main` 命令创建合并到 main 的 PR
   - 检查是否已存在 PR，避免重复创建

**使用场景**:

- 大型功能需要独立分支开发
- 多个子 Issue 的 PR 合并到功能分支
- 功能完成后合并回 main

---

## 监控和错误处理

### 13. Error Handler (`error-handler.yml`)

**功能**: 自动分析 workflow 失败原因并提供修复建议。

**触发方式**:

- Backend Agent、Frontend Agent、UI Agent 失败时自动触发
- 手动触发（提供 run_id）

**特点**:

- 收集失败的 workflow 日志
- 使用 AI 分析错误原因
- 在关联的 Issue 中发布分析报告
- 如果可自动修复，触发 `fix-pr.yml`

**输出**:

- 错误分析报告（根本原因、错误类型、修复方案）
- 更新 Issue 标签（`❌ ai-failed`, `🔍 analyzed`）

---

### 14. Fix PR Build Errors (`fix-pr.yml`)

**功能**: 修复 PR 中的构建错误。

**触发方式**:

- PR 评论中包含 `/fix` 命令
- 被 `error-handler.yml` 自动触发

**特点**:

- 运行前端构建并捕获错误
- 解析错误信息（Next.js/Turbopack 格式）
- 使用 Claude Code Action 直接修复代码
- 直接推送到当前 PR 分支（`direct_push: true`）

**支持的错误类型**:

- Module not found 错误
- Type error
- 其他构建错误

**使用场景**:

- PR 构建失败需要修复
- Vercel Preview 部署失败

---

### 15. Vercel Status Monitor (`vercel-status-monitor.yml`)

**功能**: 监控 Vercel 部署状态并更新 Issue/PR。

**触发方式**:

- Vercel 部署状态变化（`deployment_status` 事件）
- main 分支 push（测试模式）
- 手动触发

**功能**:

1. **Preview 部署失败**:
   - 在 PR 中评论，提示使用 `/fix` 命令修复
2. **生产环境部署**:
   - **Pending**: 在 PR 中通知部署开始
   - **Success**: 在 PR 中庆祝，关闭关联的 Issue
   - **Failure**: 在 PR 和 Issue 中报错，提示修复

**使用场景**:

- 跟踪 Vercel 部署状态
- 部署成功后自动关闭 Issue

---

### 16. Vibe Monitor (`vibe-monitor.yml`)

**功能**: 监控任务状态，自动检测超时和失败任务。

**触发方式**:

- 每小时自动运行（cron: `0 * * * *`）
- 手动触发（支持不同操作）

**功能**:

1. **检查状态** (`check`):
   - 统计处理中、超时、失败、待处理的任务
2. **清理超时任务** (`clean-stale`):
   - 移除超过 4 小时未更新的 `🤖 ai-processing` 标签
   - 添加 `⚠️ stale` 标签
3. **重试失败任务** (`retry-failed`):
   - 自动重试失败的任务（最多 3 次）
   - 根据任务复杂度选择对应的 Agent

**配置**:

- `STALE_HOURS`: 超时阈值（默认 4 小时）
- `RETRY_LIMIT`: 最大重试次数（默认 3 次）

---

## 其他工作流

### 17. Vibe Auto Vision (`vibe-auto-vision.yml`)

**功能**: AI 产品经理分析，对需求进行产品化拆解。

**触发方式**:

- Issue 被打上 `💡 insight` 标签时自动触发

**特点**:

- 分析核心痛点、用户故事、MVP 功能范围、潜在风险点
- 使用 `claude-sonnet-4.5` 模型

**使用场景**:

- 需要产品视角分析需求
- 评估需求的可行性和优先级

---

### 18. Vibe Smoke Test (`vibe-smoke-test.yml`)

**功能**: 功能验证测试，确保部署后功能正常。

**触发方式**:

- Issue 评论中包含 `/deploy` 命令
- PR 被打上 `vibe-deploy` 标签

**特点**:

- 等待云端构建完成（60 秒）
- 运行 smoke test 脚本验证功能

**使用场景**:

- 部署后验证功能是否正常
- 自动化测试流程

---

### 19. Weekly Maintenance (`weekly-maintenance.yml`)

**功能**: 每周仓库维护，检查依赖、安全漏洞等。

**触发方式**:

- 每周一早上 8 点（UTC）自动运行
- 手动触发

**功能**:

- 检查依赖是否过时
- 扫描安全漏洞
- 查看 90 天以上未处理的 Issue
- 检查代码中 TODO 注释
- 创建总结 Issue

---

### 20. Parent-Child Issue Guard (`parent-child-issue-guard.yml`)

**功能**: 管理父子 Issue 关系，防止父 Issue 在子 Issue 未完成时被关闭。

**触发方式**:

- Issue 关闭时自动触发
- Issue 评论中包含 `/force-close` 命令

**功能**:

1. **检查子 Issue 状态**:
   - 当父 Issue 被关闭时，检查所有子 Issue 是否已完成
   - 如果有未完成的子 Issue，自动重新打开父 Issue
   - 显示子 Issue 状态表格和进度

2. **自动关闭父 Issue**:
   - 当最后一个子 Issue 关闭时，自动关闭父 Issue
   - 更新进度评论

3. **强制关闭**:
   - `/force-close <原因>` 命令可以强制关闭父 Issue
   - 会添加 `force-closed` 标签

**使用场景**:

- 复杂任务拆分后的进度管理
- 确保所有子任务完成后再关闭主任务

---

### 21. Sync Issue Status (`sync-issue-status.yml`)

**功能**: 同步 Issue 实现状态，检测代码实现情况。

**触发方式**:

- 每周一早上 9 点（UTC）自动运行
- 手动触发
- main 分支 push（仅 backend/frontend 目录）

**功能**:

- 运行状态检测脚本
- 在指定 Issue 中发布状态报告
- 生成实现状态摘要

**使用场景**:

- 定期检查功能实现进度
- 验证 Issue 是否已通过代码实现

---

## 使用指南

### 快速开始

1. **创建 Issue 描述需求**
   - Vibe Router 会自动分析复杂度并触发对应的 Agent

2. **手动触发 Agent**（可选）

   ```
   /agent-simple    # 简单任务
   /agent-medium    # 中等任务
   /agent-complex   # 复杂任务
   /agent-ui        # 生成 UI 设计规格
   /agent-be <url>  # 后端开发
   /agent-fe <url>  # 前端开发
   ```

3. **查看进度**
   - 在 Issue 评论中查看 AI 的进度追踪
   - 在 Actions 标签页查看 workflow 执行日志

### 常用命令

| 命令              | 说明                 | 适用场景                |
| ----------------- | -------------------- | ----------------------- |
| `/agent-simple`   | 简单任务 Agent       | Bug 修复、样式调整      |
| `/agent-medium`   | 中等任务 Agent       | 新功能、多文件修改      |
| `/agent-complex`  | 复杂任务 Agent       | 大型功能、需要拆分      |
| `/agent-ui`       | UI 设计规格生成      | 需要先设计 UI           |
| `/agent-be <url>` | 后端开发             | 已有 UI Spec            |
| `/agent-fe <url>` | 前端开发             | 已有 UI Spec 或后端 API |
| `/fix`            | 修复构建错误         | PR 构建失败             |
| `/sync`           | 同步 main 到功能分支 | 功能分支需要更新        |
| `/merge-to-main`  | 创建合并 PR          | 功能完成后合并          |
| `/clean-stale`    | 清理超时任务         | 任务卡住时              |
| `/deploy`         | 触发部署验证         | 部署后测试              |

### 标签说明

**复杂度标签**:

- `complexity:simple` - 简单任务
- `complexity:medium` - 中等任务
- `complexity:complex` - 复杂任务

**状态标签**:

- `🤖 ai-processing` - AI 处理中
- `✅ ai-completed` - AI 已完成
- `❌ ai-failed` - AI 处理失败
- `⚠️ stale` - 任务超时
- `🔍 analyzed` - 已分析错误

**类型标签**:

- `frontend` - 涉及前端
- `backend` - 涉及后端
- `database` - 涉及数据库
- `sub-issue` - 子 Issue
- `epic` - 大型功能

**功能分支标签**:

- `feature:xxx` - 功能分支（自动创建 `feature/xxx` 分支）

### 最佳实践

1. **需求描述要清晰**
   - 提供具体的功能描述
   - 包含验收标准
   - 说明技术约束

2. **合理选择 Agent**
   - 简单任务用 Simple Agent
   - 需要分析用 Medium Agent
   - 大型功能用 Complex Agent 先拆分

3. **使用功能分支**
   - 大型功能使用 `feature:xxx` 标签
   - 子 Issue 的 PR 会自动合并到功能分支
   - 功能完成后使用 `/merge-to-main` 合并

4. **监控任务状态**
   - 定期查看 Vibe Monitor 的报告
   - 及时处理超时和失败的任务

5. **利用自动化**
   - 后端完成后会自动触发前端开发
   - CI 失败会自动尝试修复
   - 部署状态会自动更新 Issue

### 故障排查

**Agent 失败**:

1. 查看 Actions 日志了解详细错误
2. Error Handler 会自动分析并提供修复建议
3. 根据建议修复后重试

**构建错误**:

1. 在 PR 中评论 `/fix` 命令
2. AI 会自动分析并修复
3. 如果自动修复失败，查看错误日志手动修复

**任务超时**:

1. 使用 `/clean-stale` 清理超时任务
2. 查看 Actions 日志确认状态
3. 使用对应的 Agent 命令重试

**部署失败**:

1. 查看 Vercel 日志
2. 在 PR 中评论 `/fix` 修复构建错误
3. 修复后 Vercel 会自动重新部署

---

## 配置说明

### 必需的 Secrets

- `OPENROUTER_API_KEY`: OpenRouter API Key（用于 AI 调用）
- `RAILWAY_BACKEND_URL`: Railway 后端 URL（用于 Smoke Test）

### 环境变量配置

**Issue Router**:

- `COMMENT_THRESHOLD`: 评论数量阈值（默认 8）

**Vibe Router**:

- `SIMPLE_MAX_CHARS`: 简单需求最大字符数（默认 500）
- `MEDIUM_MAX_CHARS`: 中等需求最大字符数（默认 2000）

**Vibe Monitor**:

- `STALE_HOURS`: 超时阈值（默认 4 小时）
- `RETRY_LIMIT`: 最大重试次数（默认 3 次）

---

## 更新日志

- **2026**: 
  - 重写完整的工作流文档
  - 新增所有工作流的详细说明
  - 完善配置说明和故障排查指南
  - 更新所有日期信息为 2026 年
- **2024-2025**: 
  - 初始版本，包含所有核心工作流
  - 支持 OpenRouter 集成
  - 支持功能分支管理
  - 支持自动错误分析和修复

---

## 贡献

如需添加新的工作流或改进现有工作流，请：

1. 创建新的 workflow 文件
2. 更新本文档
3. 添加必要的测试
4. 提交 PR

---

## 相关文档

- [Backend 开发规范](../backend/CLAUDE.md)
- [Frontend 开发规范](../frontend/STYLE_GUIDE.md)
- [项目设计文档](../docs/development/project-design.md)
- [Agent 协议文档](../docs/workflow/agent-protocol.md)
