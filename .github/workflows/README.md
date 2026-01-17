# GitHub Actions 工作流文档

本文档详细说明了项目中所有 GitHub Actions 工作流的功能、触发条件和使用方法。

## 📋 目录

- [核心 Agent 工作流](#核心-agent-工作流)
- [任务复杂度路由](#任务复杂度路由)
- [自动化工作流](#自动化工作流)
- [监控工作流](#监控工作流)
- [其他工作流](#其他工作流)
- [使用指南](#使用指南)

---

## 核心 Agent 工作流

### 1. Vibe Agent (`vibe-agent.yml`) ⭐ 主入口

**功能**: 统一的 Agent 入口，处理 UI 设计、后端代码、前端代码生成。

**命令格式**:

```bash
/agent ui              # 生成 UI 设计规格
/agent be              # 生成后端代码
/agent fe              # 生成前端代码
/agent be --spec #123  # 指定 UI Spec 来源
/agent fe --spec #123  # 指定 UI Spec 来源
```

**兼容旧命令**: `/agent-ui`, `/agent-be`, `/agent-fe`

**输出策略**:

| 类型     | 输出位置                               | 说明         |
| -------- | -------------------------------------- | ------------ |
| UI Spec  | `docs/specs/issue-{number}-ui.md` + PR | 避免评论折叠 |
| 后端代码 | PR                                     | 直接生成代码 |
| 前端代码 | PR                                     | 直接生成代码 |

**工作流程**:

```
1. Issue 描述需求
        ↓
2. /agent ui → 生成 UI Spec → PR
        ↓
3. Review & Merge PR
        ↓
4. /agent be --spec #123 → 生成后端代码 → PR
        ↓
5. /agent fe --spec #123 → 生成前端代码 → PR
```

**特点**:

- UI Spec 输出到文件，不再在评论中放长内容
- 支持 `--spec` 参数指定 UI Spec 来源
- Issue 评论只放简短状态，详细内容在 PR 中

---

### 2. Simple Task Agent (`agent-simple.yml`)

**功能**: 处理简单任务，直接实现代码，无需复杂分析。

**触发方式**:

- Issue 评论中包含 `/agent-simple`
- 被 `vibe-router.yml` 自动触发（复杂度为 S）

**特点**:

- 直接开始编码，不进行需求分析
- 适合单文件修改、bug 修复、样式调整
- 最大 30 轮对话
- 若 PR 的 CI 失败，会自动触发修复流程并保持 Issue 打开

**使用场景**:

- 修复简单的 bug
- 调整 UI 样式
- 修改文案
- 添加单个 UI 元素

---

### 3. Medium Task Agent (`agent-medium.yml`)

**功能**: 处理中等复杂度任务，先分析再实现。

**触发方式**:

- Issue 评论中包含 `/agent-medium`
- 被 `vibe-router.yml` 自动触发（复杂度为 M）

**特点**:

- 两阶段处理：先分析需求，再开发实现
- 适合涉及 2-5 个文件的新功能
- 最大 50 轮对话
- 若 PR 的 CI 失败，会自动触发修复流程并保持 Issue 打开

**使用场景**:

- 新增独立功能模块
- 需要前后端都改但逻辑简单
- 新增 API endpoint + 简单 UI

---

### 4. Complex Task Agent (`agent-complex.yml`)

**功能**: 处理复杂任务，自动拆分为多个子 Issue。

**触发方式**:

- Issue 评论中包含 `/agent-complex`
- 被 `vibe-router.yml` 自动触发（复杂度为 L）

**特点**:

- 使用 AI 分析需求并拆分子任务
- 自动创建子 Issue 并设置依赖关系
- 自动触发第一个无依赖的子任务

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

## 任务复杂度路由

### 5. Vibe Router (`vibe-router.yml`)

**功能**: 分析 Issue 复杂度并路由到对应的 Agent。

**触发方式**: 手动添加 `needs-route` 标签时触发

**工作流程**:
```
1. 创建 Issue（不会自动触发）
2. 整理需求内容
3. 添加 needs-route 标签
4. Router 自动分析复杂度并路由
```

**复杂度判断标准**:

| 等级     | 说明                 | 路由目标      |
| -------- | -------------------- | ------------- |
| S (简单) | 单文件修改，< 2 小时 | agent-simple  |
| M (中等) | 2-5 个文件，2-8 小时 | agent-medium  |
| L (复杂) | > 5 个文件，> 1 天   | agent-complex |
| skip     | 非开发任务           | 不处理        |

**输出**:

- 添加复杂度标签：`complexity:simple` / `complexity:medium` / `complexity:complex`
- 添加影响范围标签：`frontend` / `backend` / `database`
- 自动触发对应的 Agent

---

## 自动化工作流

### 6. Auto Trigger Frontend (`auto-trigger-frontend.yml`)

**功能**: 后端 PR 合并后自动触发前端开发。

**触发方式**: PR 合并时自动触发（仅限后端 Agent 创建的 PR）

**使用场景**:

- 前后端分离开发
- 后端完成后自动开始前端开发

---

### 7. Feature Branch Manager (`feature-branch-manager.yml`)

**功能**: 管理功能分支，支持自动创建、同步和合并。

**命令**:

| 命令               | 说明                        |
| ------------------ | --------------------------- |
| `feature:xxx` 标签 | 自动创建 `feature/xxx` 分支 |
| `/sync`            | 同步 main 到功能分支        |
| `/merge-to-main`   | 创建合并到 main 的 PR       |

---

## 监控工作流

### 8. Vibe Continuous (`vibe-continuous.yml`) ⭐ 统一监控

**功能**: 任务监控与自动迭代引擎（合并了原 vibe-monitor 功能）

**触发方式**:

- 每 24 小时自动运行（24 小时自动迭代）
- 手动触发

**运行模式**:

| 模式           | 说明                                                |
| -------------- | --------------------------------------------------- |
| `auto`         | 自动模式：scan + clean-stale + retry-failed（默认） |
| `scan`         | 扫描所有进行中的 issue                              |
| `check`        | 仅检测状态，不做任何操作                            |
| `continue`     | 继续处理未完成任务                                  |
| `verify`       | 验收模式：检测完成度，通过则关闭 issue              |
| `clean-stale`  | 清理超时任务（标记为 stale）                        |
| `retry-failed` | 重试失败的任务                                      |

---

### 9. Fix PR Build Errors (`fix-pr.yml`)

**功能**: 修复 PR 中的构建错误。

**命令**: 在 PR 评论中使用 `/fix`

---

### 10. Vercel Status Monitor (`vercel-status-monitor.yml`)

**功能**: 监控 Vercel 部署状态并更新 Issue/PR。

---

## 其他工作流

### 11. Issue Manager (`issue-manager.yml`)

**功能**: Issue 欢迎消息与自动标签。

**触发方式**: Issue 创建时自动触发

**自动标签**: 根据 Issue 内容自动添加 `frontend`、`backend`、`bug`、`enhancement` 标签

---

### 12. Parent-Child Issue Guard (`parent-child-issue-guard.yml`)

**功能**: 管理父子 Issue 关系，防止父 Issue 在子 Issue 未完成时被关闭。

---

### 13. Daily Maintenance (`daily-maintenance.yml`)

**功能**: 每日仓库维护，检查依赖、安全漏洞等。

**触发方式**: 每天北京时间凌晨 3:00 自动运行

---

### 14. Check API Error Handling (`check-api-error-handling.yml`) ⭐ 新增

**功能**: 自动检查后端 API 错误处理是否符合规范，在 PR 合并前进行验证。

**触发条件**:

- **仅当 PR 包含 `backend/**/\*.go` 文件变更时\*\*自动触发
- 手动触发（workflow_dispatch）

**注意**: 此工作流只检查后端代码，前端和其他代码变更不会触发此检查。

**检查项**:

1. ✅ 是否使用了标准化的 `models.ErrorResponse` 格式（而不是 `gin.H`）
2. ✅ 是否正确处理了 `gorm.ErrRecordNotFound` 错误
3. ✅ 错误日志是否包含了必要的字段（`error_code`, `request_id`）
4. ✅ 404 错误是否返回了正确的错误码（如 `ANALYSIS_NOT_FOUND`, `INSIGHT_NOT_FOUND`）

**输出**:

- 在 PR 中自动评论检查结果
- 如果有错误，PR 检查会失败，阻止合并

**修复指南**:
当检查失败时，PR 评论中会包含详细的修复指南和代码示例。

---

## 使用指南

### 快速开始

1. **创建 Issue 描述需求**
   - Vibe Router 会自动分析复杂度并触发对应的 Agent

2. **手动触发 Agent**（可选）

   ```bash
   # 推荐：统一命令格式
   /agent ui              # 生成 UI 设计规格
   /agent be              # 生成后端代码
   /agent fe              # 生成前端代码
   /agent be --spec #123  # 指定 UI Spec 来源

   # 任务复杂度命令
   /agent-simple          # 简单任务
   /agent-medium          # 中等任务
   /agent-complex         # 复杂任务
   ```

3. **查看进度**
   - 在 PR 中查看生成的代码和 UI Spec
   - 在 Actions 标签页查看 workflow 执行日志

### 常用命令速查

| 命令                    | 说明                                  |
| ----------------------- | ------------------------------------- |
| `/agent ui`             | 生成 UI 设计规格 → `docs/specs/` + PR |
| `/agent be`             | 生成后端代码 → PR                     |
| `/agent fe`             | 生成前端代码 → PR                     |
| `/agent be --spec #123` | 基于指定 Issue 的 UI Spec 生成后端    |
| `/agent-simple`         | 简单任务 Agent                        |
| `/agent-medium`         | 中等任务 Agent                        |
| `/agent-complex`        | 复杂任务拆分                          |
| `/fix`                  | 修复 PR 构建错误                      |
| `/sync`                 | 同步 main 到功能分支                  |
| `/merge-to-main`        | 创建合并 PR                           |
| `/clean-stale`          | 清理超时任务                          |

### 标签说明

**复杂度标签**:

- `complexity:simple` - 简单任务
- `complexity:medium` - 中等任务
- `complexity:complex` - 复杂任务

**状态标签**:

- `ai:processing` - AI 处理中
- `ai:completed` - AI 已完成
- `ai:failed` - AI 处理失败
- `ui-spec-ready` - UI Spec 已生成
- `needs-review` - 需要人工审查（包含以前的 no-pr, ci-failed, ci-pending 等情况）

**类型标签**:

- `frontend` - 涉及前端
- `backend` - 涉及后端
- `feature:xxx` - 功能分支

### 最佳实践

1. **使用统一的 /agent 命令**
   - 推荐使用 `/agent ui|be|fe` 格式
   - 旧命令仍然兼容

2. **UI Spec 输出到文件**
   - UI Spec 保存在 `docs/specs/` 目录
   - 通过 PR 进行 Review
   - 避免 Issue 评论折叠问题

3. **使用 --spec 参数**
   - 生成代码时指定 UI Spec 来源
   - 例如: `/agent be --spec #123`

4. **功能分支开发**
   - 大型功能使用 `feature:xxx` 标签
   - 子任务 PR 自动合并到功能分支

---

## 配置说明

### 必需的 Secrets

- `OPENROUTER_API_KEY`: OpenRouter API Key

### 可复用 Actions

项目提供可复用的 Composite Actions，用于减少工作流代码重复：

#### 1. Load Prompt (`/.github/actions/load-prompt/action.yml`)

从模板文件加载并渲染 Prompt：

```yaml
- uses: ./.github/actions/load-prompt
  with:
    template: agents/vibe/fe-codegen.md
    variables: '{"requirement": "...", "project_context": "..."}'
```

#### 2. Context Discovery (`/.github/actions/context-discovery/action.yml`)

自动发现项目上下文（技术栈、目录结构等）：

```yaml
- uses: ./.github/actions/context-discovery
  with:
    requirement: "需求描述"
    target: frontend # 或 backend
```

### 配置文件

项目使用中央配置文件管理工作流配置：

**`.github/config/workflow-config.json`**

```json
{
  "version": "1.1.0",
  "prd": {
    "issue_number": 176,
    "sub_issues": [...]
  },
  "router": {
    "complexity_thresholds": {
      "simple_max_chars": 500,
      "medium_max_chars": 2000
    }
  },
  "monitor": {
    "stale_threshold_hours": 4,
    "retry_limit": 3
  },
  "agents": {
    "default_model": "anthropic/claude-sonnet-4",
    "ui_model": "google/gemini-2.0-flash-001",
    "router_model": "google/gemini-2.0-flash-001",
    "max_turns": { "simple": 30, "medium": 50, "complex": 60 }
  },
  "paths": {
    "spec_dir": "docs/specs",
    "prompts_dir": ".github/prompts"
  },
  "labels": {
    "status": {...},
    "complexity": {...},
    "scope": {...},
    "ui_spec": "ui-spec-ready"
  },
  "skip_patterns": {...},
  "api": {
    "openrouter_base_url": "https://openrouter.ai/api/v1"
  },
  "git": {
    "bot_name": "vibe-agent[bot]",
    "bot_email": "vibe-agent@github-actions.bot"
  }
}
```

优点：

- 集中管理配置，避免硬编码
- 支持配置 Schema 验证
- 方便修改阈值和标签名

### 文件结构

```
.github/
├── actions/                      # 可复用 Composite Actions (2 个)
│   ├── load-prompt/              # Prompt 模板加载器
│   └── context-discovery/        # 项目上下文发现
├── config/
│   └── workflow-config.json      # 中央配置文件
├── prompts/                      # AI Agent Prompt 模板 (9 个)
│   ├── router/
│   │   └── complexity-analyzer.md
│   └── agents/
│       ├── simple.md             # 简单任务 Agent
│       ├── medium.md             # 中等任务 Agent
│       └── vibe/                 # Vibe Agent 专用
│           ├── pm-compiler.md    # PM 需求编译
│           ├── ui-spec.md        # UI 规格生成
│           ├── be-contract.md    # 后端契约定义
│           ├── be-codegen.md     # 后端代码生成
│           └── fe-codegen.md     # 前端代码生成 (含 Base.org 设计系统)
├── scripts/                      # 独立脚本文件 (3 个)
│   ├── vibe-continuous.js        # 自动迭代引擎脚本
│   ├── vibe-router.js            # 复杂度路由脚本
│   └── agent-utils.js            # Agent 共享工具函数
├── workflows/                    # GitHub Actions 工作流 (16 个)
│   ├── vibe-agent.yml            # 主 Agent 入口
│   ├── vibe-router.yml           # 复杂度路由（使用 vibe-router.js）
│   ├── vibe-continuous.yml       # 任务监控与自动迭代（使用 vibe-continuous.js）
│   ├── agent-simple.yml          # 简单任务处理
│   ├── agent-medium.yml          # 中等任务处理
│   ├── agent-complex.yml         # 复杂任务拆分
│   ├── auto-trigger-frontend.yml # 后端完成后触发前端
│   ├── feature-branch-manager.yml # 功能分支管理
│   ├── dependency-chain-trigger.yml # 任务依赖链触发
│   ├── fix-pr.yml                # PR 构建错误修复
│   ├── check-api-error-handling.yml # API 错误处理检查
│   ├── issue-manager.yml         # Issue 欢迎消息与自动标签
│   ├── parent-child-issue-guard.yml # 父子 Issue 关系守护
│   ├── update-prd-status.yml     # PRD 状态更新
│   ├── vercel-status-monitor.yml # Vercel 部署监控
│   ├── daily-maintenance.yml     # 每日维护任务
│   └── README.md                 # 本文档
└── AGENT_GUIDE.md                # Agent 使用指南

docs/
└── specs/
    └── issue-{number}-ui.md      # 自动生成的 UI Spec
```

---

## 更新日志

- **2026-01-16** (结构优化 - 第四阶段):
  - ✅ **提取共享工具函数**：
    - 新增 `scripts/agent-utils.js`：Agent 共享工具函数
    - 包含：进度更新、标签管理、PR 查找、重复检查、错误分析等
  - 当前保留 **16 个 workflow**、**2 个 actions**、**9 个 prompts**、**3 个脚本**

- **2026-01-16** (结构优化 - 第三阶段):
  - ✅ **清理未使用的资源**：
    - 删除 `prompts/agents/complex.md`（未被 agent-complex.yml 使用）
    - 删除 `actions/openrouter-api/`（未被任何 workflow 使用）
    - 删除 `actions/update-issue-status/`（未被任何 workflow 使用）

- **2026-01-16** (结构优化 - 第二阶段):
  - ✅ **提取内联脚本到独立文件**：
    - 新增 `scripts/vibe-router.js`：复杂度路由逻辑
    - `vibe-router.yml` 从 270+ 行简化到 114 行
  - ✅ **清理冗余功能**：
    - 删除 `issue-manager.yml` 中的 `/clean-stale` 命令（已被 `vibe-continuous` 的 `clean-stale` 模式替代）
    - 简化 `issue-manager.yml` 触发条件（仅在 Issue 创建时触发）

- **2026-01-16** (结构优化 - 第一阶段):
  - ✅ **删除冗余 workflow**：
    - 删除 `auto-trigger-agent.yml`（功能与 `vibe-router.yml` 重复）
    - 删除 `vibe-monitor.yml`（功能合并到 `vibe-continuous.yml`）
  - ✅ **合并 vibe-monitor 到 vibe-continuous**：
    - 新增 `clean-stale` 模式：清理超时任务
    - 新增 `retry-failed` 模式：重试失败任务
    - 改为每小时自动运行
  - ✅ **重命名 workflow**：
    - `weekly-maintenance.yml` → `daily-maintenance.yml`（名称与实际频率一致）
  - 当前保留 **16 个有效 workflow**、**2 个独立脚本**

- **2026-01-17**:
  - ✅ **Prompt 模板化完成**：所有 workflow 中的 prompt 已提取为独立模板文件
    - `vibe-agent.yml` 使用 `load-prompt` Action 加载模板
    - 支持变量替换，便于维护和版本控制
    - 模板位置：`.github/prompts/agents/vibe/`
  - ✅ **统一 Agent 入口优化**：
    - `vibe-agent.yml` 重构，支持 `/agent ui|be|fe` 统一命令格式
    - 兼容旧命令：`/agent-ui`, `/agent-be`, `/agent-fe`
    - 支持 `--spec #123` 参数指定 UI Spec 来源
    - UI Spec 输出到 `docs/specs/issue-{number}-ui.md`，避免评论折叠
  - ✅ **工作流和前端路由系统重构**：
    - 优化任务路由逻辑
    - 改进前端代码生成流程
  - ✅ **每日维护工作流调整**：
    - `weekly-maintenance.yml` 改为每天凌晨 3:00（北京时间）执行
    - 支持手动触发（workflow_dispatch）
    - 更新文档说明查看方法

- **2026-01-16** (目录结构优化):
  - 新增 `update-issue-status` Action：统一 Issue 状态标签管理
  - 整合前端 prompt：合并 `fe/system-prompt.md` 到 `fe-codegen.md`
    - 包含完整的 Base.org 设计系统规范
    - 颜色系统、圆角系统、无阴影/无边框设计原则
  - 清理冗余文件：
    - 删除 `prompts/backend-agent-prompt.md`（已被 be-contract.md 和 be-codegen.md 替代）
    - 删除 `prompts/zhangxiaohao-prompt.md`（未使用的个人 prompt）
    - 删除 `prompts/fe/` 目录（内容已整合）
    - 删除 `scripts/` 目录（7 个脚本均未被工作流使用）
    - 删除 `actions/github-utils/`（功能已整合到 update-issue-status）
  - 更新 `workflow-config.json`：
    - 新增 router.complexity_thresholds 配置
    - 新增 paths 配置（spec_dir, prompts_dir）
    - 新增 git 配置（bot_name, bot_email）
  - 当前保留 **14 个有效 workflow**、**4 个 Actions**、**9 个 Prompt 模板**

- **2026-01-16** (工作流优化):
  - 新增可复用 Composite Actions：
    - `openrouter-api`: 带重试机制的 OpenRouter API 客户端
    - `load-prompt`: Prompt 模板加载器
    - `context-discovery`: 项目上下文发现
  - 新增中央配置文件 `.github/config/workflow-config.json`
  - 重构 `vibe-router.yml`：
    - 升级模型到 `google/gemini-2.0-flash-001`
    - 添加 API 调用重试机制（指数退避）
    - 从配置文件读取跳过规则和标签
  - 重构 `update-prd-status.yml`：从配置文件读取 PRD 配置
  - 重构 `vibe-monitor.yml`：从配置文件读取阈值配置
  - 清理无效 workflow 文件：
    - 删除 `vibe-smoke-test.yml`（依赖不存在的脚本）
    - 删除 `vibe-auto-vision.yml`（YAML 语法错误）
    - 删除 `auto-fix-CI-failures.yml`（监听不存在的 CI workflow）
    - 删除 `sync-issue-status.yml`（硬编码 issue 号，功能过时）
    - 删除 `error-handler.yml`（监听不存在的 workflows）

- **2026-01** (统一 Agent 入口):
  - 统一 Agent 入口 (`vibe-agent.yml`)
  - 合并 issue-router/agent-ui/backend-agent/frontend-agent
  - 新增 `/agent ui|be|fe` 命令格式
  - UI Spec 输出到文件，避免评论折叠
  - 支持 `--spec` 参数指定 UI Spec 来源

- **2025-12** (新增功能):
  - 新增 `parent-child-issue-guard.yml`：管理父子 Issue 关系
  - 新增 `update-prd-status.yml`：自动更新 PRD Issue 状态
  - 新增 `check-api-error-handling.yml`：自动检查 API 错误处理规范

- **2024-2025**:
  - 初始版本，包含所有核心工作流
  - 支持 OpenRouter 集成
  - 支持功能分支管理

---

## 相关文档

- [Backend 开发规范](../../backend/CLAUDE.md)
- [Frontend 开发规范](../../frontend/STYLE_GUIDE.md)
- [UI Specs 目录](../../docs/specs/)
