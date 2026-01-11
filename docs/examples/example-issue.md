# 完整的测试 Issue 示例

**标题**：`[Agent] 创建一个简单的 Python 脚本：生成每日工作清单`

**正文**（直接复制到 GitHub Issue）：

```markdown
## 目标（一句话）
创建一个简单的 Python 脚本，根据模板生成每日工作清单的 Markdown 文件。

## 用户价值（为什么值得做）
- 让团队成员能快速生成当天的工作清单
- 验证整个 Agent 自动化流程是否正常工作
- 产出可运行的代码，展示完整的开发流程

## 范围（做什么 / 不做什么）
### 做什么
- 创建 `scripts/generate_todo.py` 脚本
- 脚本功能：
  - 读取 `DAILY_TODOLIST.md` 作为模板
  - 生成一个新的 Markdown 文件，文件名格式：`daily-YYYY-MM-DD.md`
  - 在生成的文件顶部自动填充日期（今天的日期）
- 添加简单的使用说明到 `README.md`

### 不做什么
- 不需要复杂的命令行参数（可以硬编码输出路径）
- 不需要 GUI 或交互式界面
- 不需要引入额外的第三方依赖（只用 Python 标准库）

## 验收标准（通过/不通过）
1. ✅ 脚本 `scripts/generate_todo.py` 存在且可以运行
2. ✅ 运行 `python3 scripts/generate_todo.py` 后，生成 `daily-2025-12-XX.md` 文件
3. ✅ 生成的文件包含日期信息（今天的日期）
4. ✅ `README.md` 中有使用说明

## 边界与反例（至少 2 条）
- 如果 `DAILY_TODOLIST.md` 不存在：脚本应该给出清晰的错误提示并退出
- 如果输出目录不存在：脚本应该自动创建目录

## 依赖与资源
- Python 3.11+（GitHub Actions 已安装）
- 标准库即可，无需额外依赖

## 风险与假设（最多 3 条）
1. 假设 `DAILY_TODOLIST.md` 文件存在于仓库根目录
2. 假设脚本运行环境有写入权限

## 交付物（可见成果）
- [ ] `scripts/generate_todo.py` 脚本文件
- [ ] 生成的示例文件 `daily-YYYY-MM-DD.md`
- [ ] `README.md` 中的使用说明

## 给 AI 队友的上下文（可选但推荐）
- 必须遵守 `AGENT_PROTOCOL.md`：先写计划到 `EXEC_PLAN.md`，列出要改哪些文件
- 代码变更要小且可审，必须产生非空 diff
- 在 PR 描述中引用本 Issue，并提供验证证据（命令+输出）
```

---

## 使用方法

1. 在 GitHub 仓库中点击 "Issues" → "New Issue"
2. 标题填写：`[Agent] 创建一个简单的 Python 脚本：生成每日工作清单`
3. 将上面的正文内容复制到 Issue 描述中
4. 添加 `agent-task` label（或标题已包含 `[Agent]` 前缀）
5. 提交 Issue，等待 GitHub Actions 自动运行并创建 PR
6. 检查 PR 中的代码改动，确认脚本可以运行

## 预期结果

- GitHub Actions 会自动运行
- Codex 会读取 Issue 和协议文件
- 生成 `EXEC_PLAN.md` 规划文件
- 创建 `scripts/generate_todo.py` 脚本
- 更新 `README.md` 添加使用说明
- 自动创建 PR，包含所有代码改动
- PR 可以通过 Review Checklist 进行代码审查

