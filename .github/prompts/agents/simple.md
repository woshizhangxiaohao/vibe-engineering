## 任务类型

这是一个 **简单任务**，请直接实现，不需要复杂分析。

## Issue 信息

- 仓库: {{repository}}
- Issue: #{{issue_number}}
- 标题: {{title}}

## 需求描述

{{body}}

## 执行要求

1. **直接开始编码**，不需要先分析或规划
2. 遵循项目现有代码风格（参考 CLAUDE.md 和 STYLE_GUIDE.md）
3. 修改尽量最小化，只做必要的改动
4. 如果涉及前端，使用 shadcn/ui 组件
5. 如果涉及后端，遵循现有的 handler/service/repository 模式
6. 完成后创建 PR

## 项目结构提示

- 前端: frontend/ (Next.js + TypeScript + shadcn/ui)
- 后端: backend/ (Go + Gin)
- 类型定义: frontend/lib/api/types.ts
- API 客户端: frontend/lib/api/client.ts

## 注意事项

- 不要过度设计
- 不要添加不必要的功能
- 如果需求不明确，做最简单的实现
