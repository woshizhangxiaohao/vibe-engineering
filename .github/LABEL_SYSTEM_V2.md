# 标签系统 V2 - 优化后版本

> 更新时间：2026-01-17  
> 从 23 个标签优化到 10 个核心标签

---

## ✅ 核心改进

1. **触发方式改变**：从自动触发改为手动触发（只有 `needs-route`）
2. **移除冗余**：删除 `agent:xxx` 标签（与 `complexity:xxx` 重复）
3. **去除 emoji**：状态标签更专业（`ai:processing` 代替 `🤖 ai-processing`）
4. **合并临时状态**：统一使用 `needs-review`，在评论中说明具体原因
5. **职责分明**：触发、范围、状态标签职责清晰

---

## 📋 当前标签系统（10 个核心标签）

### 1️⃣ 复杂度标签（3个）
```
complexity:simple   - 简单任务（单文件，< 2小时）
complexity:medium   - 中等任务（2-5文件，2-8小时）
complexity:complex  - 复杂任务（> 5文件，> 1天）
```

**用途**：由 Vibe Router 自动添加，用于判断使用哪个 Agent

### 2️⃣ AI 状态标签（3个）
```
ai:processing  - AI 正在处理中
ai:completed   - AI 已完成
ai:failed      - AI 处理失败
```

**用途**：跟踪 Agent 执行状态

### 3️⃣ 范围标签（3个）
```
backend   - 涉及后端代码
frontend  - 涉及前端代码
database  - 涉及数据库
```

**用途**：由 Vibe Router 自动添加，标识 Issue 影响范围

### 4️⃣ 控制标签（1个）
```
needs-route  - 需要路由分析
```

**用途**：**唯一的触发标签**，手动添加后会触发 Vibe Router

### 5️⃣ 其他必要标签
```
needs-review    - 需要人工审查（合并了 no-pr, ci-failed, ci-pending）
ui-spec-ready   - UI Spec 已生成
feature:xxx     - 功能分支（动态标签）
```

---

## 🔄 工作流程

### 创建新需求
```
1. 创建 Issue
   ↓
2. 完善需求描述（不会触发任何自动化）
   ↓
3. 添加 needs-route 标签
   ↓
4. Router 自动分析：
   - 添加 complexity:xxx 标签
   - 添加范围标签（backend/frontend/database）
   - 移除 needs-route 标签
   ↓
5. 触发对应的 Agent（Simple/Medium/Complex）
   ↓
6. 添加 ai:processing 标签
   ↓
7. Agent 执行完成：
   - 成功：ai:completed
   - 失败：ai:failed + needs-review
```

### 标签状态流转
```
创建 Issue
    ↓
手动添加 needs-route
    ↓
complexity:medium + backend + ai:processing
    ↓
ai:completed  或  ai:failed + needs-review
```

---

## 🚫 已移除的标签

### 冗余标签（-3）
```
❌ agent:simple    （与 complexity:simple 重复）
❌ agent:medium    （与 complexity:medium 重复）
❌ agent:complex   （与 complexity:complex 重复）
```

### 跳过标签（-4）
```
❌ skip-vibe      （不需要了，默认不触发）
❌ manual         （不需要了，默认不触发）
❌ question       （不需要了，默认不触发）
❌ discussion     （不需要了，默认不触发）
```

### 临时状态标签（-3，合并为 needs-review）
```
❌ ⚠️ no-pr       → needs-review + 评论说明
❌ ⚠️ ci-failed   → needs-review + 评论说明
❌ ⚠️ ci-pending  → needs-review + 评论说明
```

### Emoji 状态标签（-3，改为无 emoji 版本）
```
❌ 🤖 ai-processing  → ai:processing
❌ ✅ ai-completed   → ai:completed
❌ ❌ ai-failed      → ai:failed
```

---

## 💡 使用指南

### 场景 1：创建新开发需求
```bash
1. 创建 Issue，填写需求
2. 慢慢完善（不会触发）
3. 准备好了，添加 needs-route
4. 等待 Router 分析并自动路由
```

### 场景 2：重新路由失败的任务
```bash
1. Issue 有 ai:failed + needs-review 标签
2. 移除 ai:failed 和 complexity:xxx 标签
3. 添加 needs-route 标签
4. 重新触发路由
```

### 场景 3：查看 Issue 状态
```bash
- 无复杂度标签：未路由
- complexity:xxx + ai:processing：正在处理
- complexity:xxx + ai:completed：已完成
- complexity:xxx + ai:failed + needs-review：失败，需审查
- needs-review：需要人工处理（查看评论了解具体原因）
```

---

## 📊 对比

| 项目 | V1（旧） | V2（新） | 改进 |
|------|---------|---------|------|
| 标签总数 | 23 | 10 | **-57%** |
| 触发方式 | 自动（需跳过标签） | 手动（needs-route） | **更可控** |
| 状态标签 | 带 emoji | 无 emoji | **更专业** |
| Agent 标签 | 有（重复） | 无 | **无冗余** |
| 临时状态 | 3个独立标签 | 1个 + 评论 | **更简洁** |

---

## 🎯 优势

1. **标签数量减少 57%**：从 23 个减少到 10 个
2. **更易理解**：职责清晰，不重复
3. **更好控制**：默认不触发，需要时才触发
4. **更专业**：去除 emoji，更适合程序化处理
5. **更灵活**：通过评论说明具体情况，而非创建大量标签

---

## 📝 迁移说明

### 自动迁移
- 所有 workflow 已更新
- 新创建的 Issue 会使用新标签系统

### 旧 Issue 处理
- 旧标签不影响使用
- 可以逐步手动清理旧标签
- 重新路由时会使用新标签

---

**查看完整优化方案**：`.github/LABEL_OPTIMIZATION.md`
