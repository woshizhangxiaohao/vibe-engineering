# 工作流问题总结与修复记录

## 发现的问题

### 1. ❌ 依赖链正则表达式匹配失败 (已修复)

**问题描述**：
`dependency-chain-trigger.yml` 中的正则表达式无法匹配 markdown 格式的依赖声明。

**原始正则**：
```javascript
body.match(/前置依赖[：:]\s*((?:#\d+(?:,\s*)?)+)/i)
```

**实际格式**：
```markdown
**前置依赖:** #222
```

**根因**：正则要求 `前置依赖` 后直接跟 `：` 或 `:`，但实际有 `**` markdown 加粗符号。

**修复**：
```javascript
body.match(/\*{0,2}前置依赖\*{0,2}[：:]\s*((?:#\d+(?:,\s*)?)+)/i)
```

**提交**：`08e6960b`

---

### 2. ❌ Agent 完成后未自动关闭 Issue (已修复)

**问题描述**：
`agent-medium.yml` 和 `agent-simple.yml` 成功完成后只添加 `✅ ai-completed` 标签，但不关闭 Issue。

**影响**：依赖链触发器依赖 Issue 的 `closed` 事件，导致后续任务无法自动触发。

**修复**：在 "Update Issue on Success" 步骤添加：
```javascript
await github.rest.issues.update({
  owner: context.repo.owner,
  repo: context.repo.repo,
  issue_number: issueNumber,
  state: 'closed',
  state_reason: 'completed'
});
```

**提交**：`a779132a`

---

### 3. ❌ 依赖链触发评论缺少 Actions 链接 (已修复)

**问题描述**：触发下一个任务时的评论只有通用的 Actions 页面链接，没有具体的运行链接。

**修复**：在触发 workflow 后等待 2 秒，然后查询最新的 workflow run 获取具体链接。

**提交**：`398b40fe`

---

### 4. ❌ 失败提示信息不准确 (已修复)

**问题描述**：Agent 失败时显示通用的"需求描述不够清晰"提示，没有显示真实的错误原因。

**实际可能原因**：
- Edit 工具的 `old_string` 匹配失败
- 依赖包缺失
- 构建错误
- API 调用失败

**修复方案**：
1. 添加 `continue-on-error: true` 捕获 Claude 执行结果
2. 检查 Claude 执行输出文件提取错误信息
3. 根据错误类型智能分析并给出针对性建议
4. 在评论中显示错误详情（折叠区域）

**提交**：本次提交

---

### 5. ⚠️ 并行任务未同时触发 (设计限制)

**问题描述**：当多个任务的依赖同时满足时，依赖链触发器已经支持并行触发。

**当前状态**：代码已正确遍历所有满足条件的任务并触发，无需额外修复。

---

## 工作流执行路径

```
Issue 创建
    ↓
/agent-complex (任务拆分)
    ↓
创建子 Issues (带 sub-issue 标签)
    ↓
自动触发第一个无依赖的任务
    ↓
Agent 执行 (simple/medium)
    ↓
成功 → 添加标签 + 关闭 Issue → 触发 dependency-chain-trigger
    ↓
检查依赖满足的下一个任务 → 并行触发
    ↓
循环直到所有任务完成
```

## 修复清单

| # | 问题 | 状态 | 提交 |
|---|------|------|------|
| 1 | 正则表达式匹配 | ✅ 已修复 | `08e6960b` |
| 2 | 自动关闭 Issue | ✅ 已修复 | `a779132a` |
| 3 | Actions 链接 | ✅ 已修复 | `398b40fe` |
| 4 | 失败提示信息 | ✅ 已修复 | 本次提交 |
| 5 | 并行触发 | ✅ 无需修复 | - |

---

*最后更新: 2026-01-16*
