# 🌊 VibeFlow: AI-Native Development Workflow

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Status](https://img.shields.io/badge/status-experimental-orange)
![AI-Powered](https://img.shields.io/badge/AI-OpenRouter-purple)

**VibeFlow** 是一个探索性的 GitHub Action 工作流套件，旨在通过 AI Agent (Claude-3.5-Sonnet) 将 GitHub Issue 直接转化为可运行的代码 PR，实现“需求即代码”的自动化闭环。

## 1. VibeFlow 思维导图 (Conceptual Mind Map)
```mermaid
graph LR
    direction LR

    subgraph S1 [阶段一：需求]
        A[👤 Issue] --> B["🤖 PM Agent<br/>(识别 Label/艾特)"]
    end

    subgraph S2 [阶段二：快速编码]
        B --> C["⚙️ Runner 扫描"]
        C -- "生成目录树 + Config" --> D["🤖 Codegen Agent<br/>(FE/BE)"]
        D --> E[📦 提交 PR]
    end

    subgraph S3 [阶段三：真人复核]
        E --> F["🤖 AI Review<br/>(Guard)"]
        F --> G["👨‍💻 真人工程师<br/>(Reviewer)"]
        G --> H["🚀 Merge"]
    end

    style C fill:#fff1f0,stroke:#ff4d4f,stroke-dasharray: 5 5
```

这个流程图强调了**“人类设定目标，AI 执行路径，人类验收结果”**的循环。AI 不再是一个简单的辅助工具，而是介入了特定环节的“虚拟员工”。

## 2. VibeFlow 技术架构流程图 (Technical Architecture Flowchart)

这张泳道图展示了如何在 GitHub 平台、GitHub Actions 运行环境和 OpenRouter AI API 之间流转的。

```mermaid
graph LR
    %% 方向：从左到右
    direction LR

    %% 定义样式
    classDef human fill:#e1f5fe,stroke:#0288d1,stroke-width:2px;
    classDef github fill:#f3e5f5,stroke:#7b1fa2;
    classDef runner fill:#fff3e0,stroke:#e65100;
    classDef ai fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px;

    %% 阶段一：需求识别与分发
    subgraph S1 ["阶段一：需求识别 (Planning)"]
        H1["👤 创建 Issue<br/>(含标签/艾特)"]:::human --> G1("🐙 GitHub Event"):::github
        G1 --> R1["🤖 PM Agent<br/>(任务拆解)"]:::ai
        R1 -- "识别 FE/BE 标签" --> R2{任务分发器}
    end

    %% 阶段二：快速编码与上下文注入
    subgraph S2 ["阶段二：AI 编码 (Execution)"]
        R2 -- "/codegen" --> R3["⚙️ Runner 静态扫描<br/>(生成目录树+依赖)"]:::runner
        R3 -- "注入上下文" --> R4["🤖 定向 Agent<br/>(FE 或 BE)"]:::ai
        R4 --> R5["📦 自动生成 PR"]:::runner
    end

    %% 阶段三：AI 初审与真人闭环
    subgraph S3 ["阶段三：质量闭环 (Review)"]
        R5 --> R6["🤖 Guard Agent<br/>(AI 报告)"]:::ai
        R6 -- "自动指派" --> H2["👨‍💻 真人专家<br/>(FE/BE Lead)"]:::human
        H2 -- "Final CR" --> H3{决策: Merge?}:::human
        H3 -- "Approve" --> END["🚀 生产发布"]:::github
        H3 -- "Reject" --> H1
    end

    %% 连线美化
    R2 -- "反馈方案" --> G_COM["💬 Issue 评论"]:::github
```

**事件驱动 (Event-Driven)**：整个系统是“休眠”的，只有当 GitHub 上发生特定事件（开 Issue、写评论、提 PR）时才会被唤醒。这非常高效且节省资源。

**上下文增强 (Context RAG)**：注意 R2b 节点。这是我之前建议补全的关键步骤。AI 不是在真空中写代码，Action Runner 必须先读取当前仓库的文件结构和关键配置（如 go.mod, package.json），把这些“上下文”一起喂给 AI，它才能写出正确的、可运行的代码。

## 🚀 核心功能

### 1. 📝 Spec Generation (规划)
当你创建一个 **Issue** 时，VibeFlow 会自动分析需求，生成一份结构化的 **Vibe Relay Card**（技术接力卡）。
- **作用**: 将模糊需求转化为 Context, Backend, Frontend 明确的技术方案。
- **触发**: `New Issue`

### 2. ⚡️ Auto Codegen (编码)
在 Issue 评论区输入 `/codegen` 指令，AI 工程师将接管键盘。
- **流程**: 读取 Issue 上下文 + 项目目录结构 -> 生成代码 -> 自动创建分支 -> 提交 PR。
- **触发**: `Issue Comment: /codegen`

### 3. 🛡️ Night Watch (审查)
当有 **Pull Request** 提交或更新时，AI 会自动进行 Code Review。
- **输出**: Vibe Score (1-10)、关键 Bug 预警、优化建议。
- **触发**: `PR Open / Synchronize`

---

## 📚 文档结构

项目文档已重新组织，更加清晰专业：

```
vibe-engineering-playbook/
├── README.md                           # 项目主文档
├── DEPLOYMENT.md                       # 部署指南
├── docs/
│   ├── workflow/                       # 工作流程文档
│   │   ├── agent-protocol.md          # AI Agent 协议
│   │   ├── daily-todolist.md          # 每日工作清单模板
│   │   └── review-checklist.md        # 代码审查清单
│   ├── development/                    # 开发指南
│   │   ├── local-development.md       # 本地开发指南
│   │   ├── project-design.md          # 项目设计文档
│   │   └── backend-spec.md            # 后端技术规范
│   ├── templates/                      # 各类模板
│   │   └── pull-request-template.md   # PR 模板
│   └── examples/                       # 示例文档
│       └── example-issue.md           # Issue 示例
├── backend/                            # 后端代码及文档
└── frontend/                           # 前端代码及文档
```

### 核心文档链接
- **开始使用**: [本地开发指南](docs/development/local-development.md)
- **部署**: [部署指南](DEPLOYMENT.md)
- **工作流**: [AI Agent 协议](docs/workflow/agent-protocol.md)
- **代码审查**: [Review Checklist](docs/workflow/review-checklist.md)

---

## 🛠️ 安装与配置

### 1. 设置 Secrets
在你的 GitHub 仓库 `Settings` -> `Secrets and variables` -> `Actions` 中添加：
- `OPENROUTER_API_KEY`: 你的 OpenRouter API Key (推荐使用 Claude 3.5 Sonnet 模型)

### 2. 部署 Workflow
将本项目 `.github/workflows` 目录下的 YAML 文件复制到你的仓库中：
- `vibe-spec-guard.yml`: 处理 Issue 分析和 PR 审查。
- `vibe-codegen.yml`: 处理代码生成指令。

### 3. 权限设置
确保你的 Workflow 拥有读写权限。在 `.github/workflows` 文件中已配置：
```yaml
permissions:
  contents: write
  pull-requests: write
  issues: write
