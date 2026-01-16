## 任务类型

这是一个 **中等复杂度任务**，需要先分析再实现。

## Issue 信息

- 仓库: {{repository}}
- Issue: #{{issue_number}}
- 标题: {{title}}

## 需求描述

{{body}}

## 执行流程

### 阶段 1: 需求分析

1. 阅读并理解需求
2. 分析影响范围（涉及哪些文件/模块）
3. 评估技术方案
4. 识别可能的风险点

### 阶段 2: 开发实现

1. 遵循项目代码风格（参考 CLAUDE.md 和 STYLE_GUIDE.md）
2. 按模块逐步实现
3. 确保代码质量
4. 完成后创建 PR

## 项目结构

- 前端: frontend/ (Next.js + TypeScript + shadcn/ui)
- 后端: backend/ (Go + Gin)
- 类型定义: frontend/lib/api/types.ts
- API 客户端: frontend/lib/api/client.ts

## 质量要求

- 代码清晰易懂
- 遵循现有的设计模式
- 处理边界情况
- 添加必要的错误处理

## 注意事项

- 如果需求不明确，在评论中提出问题
- 复杂逻辑要添加注释
- 确保与现有代码风格一致
