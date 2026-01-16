## 任务类型

这是一个 **简单任务**，请直接实现，不需要复杂分析。

## Issue 信息

- 仓库: {{repository}}
- Issue: #{{issue_number}}
- 标题: {{title}}

## 需求描述

{{body}}

## ⚠️ 重要：你必须完成实际的代码修改

你的任务是 **实际编写代码并创建 PR**，而不仅仅是分析。如果你只做分析而没有写代码，任务就是失败的。

## 执行流程

### 步骤 1: 创建分支

**必须先创建新分支再开始编码：**

```bash
git checkout -b claude/issue-{{issue_number}}-$(date +%Y%m%d-%H%M)
```

### 步骤 2: 直接开始编码

1. **不需要分析**，直接开始写代码
2. 修改尽量最小化，只做必要的改动
3. 如果涉及前端，使用 shadcn/ui 组件
4. 如果涉及后端，遵循现有的 handler/service/repository 模式

### 步骤 3: 提交代码

```bash
git add .
git commit -m "feat(issue-{{issue_number}}): 简短描述

Co-Authored-By: Claude <noreply@anthropic.com>"
```

### 步骤 4: 推送并创建 PR

```bash
git push -u origin HEAD
gh pr create --title "feat: {{title}}" --body "## Summary
- 完成了 Issue #{{issue_number}} 的需求

## Changes
- 列出主要修改

Closes #{{issue_number}}"
```

## 项目结构

- 前端: `frontend/` (Next.js + TypeScript + shadcn/ui)
- 后端: `backend/` (Go + Gin + GORM)
- 后端模型: `backend/internal/model/`
- 后端 Handler: `backend/internal/handler/`
- 类型定义: `frontend/lib/api/types.ts`
- API 客户端: `frontend/lib/api/client.ts`

## 注意事项

- 不要过度设计
- 不要添加不必要的功能
- 如果需求不明确，做最简单的实现

## 成功标准

任务成功的标志是：
1. ✅ 创建了新分支
2. ✅ 编写了实现需求的代码
3. ✅ 代码已提交并推送
4. ✅ 已创建 PR
5. ✅ **更新了 Issue body 中的验收标准 checkbox**

**如果没有创建 PR，任务就是失败的。**

## ⚠️ 必须更新验收标准 Checkbox

在完成任务后，你必须更新 Issue body，将验收标准的 checkbox 从 `- [ ]` 改为 `- [x]`：

```bash
# 获取当前 Issue body，更新 checkbox，然后保存
CURRENT_BODY=$(gh issue view {{issue_number}} --json body --jq '.body')
UPDATED_BODY=$(echo "$CURRENT_BODY" | sed 's/- \[ \] /- [x] /g')
gh issue edit {{issue_number}} --body "$UPDATED_BODY"
```

**不更新 checkbox 等于任务未完成！**

## 处理"代码已存在"的情况

如果你发现代码已经存在：
1. **验证代码是否满足所有验收标准**
2. **必须更新 Issue body 中的 checkbox**
3. 不需要创建 PR，但必须在评论中说明原因
