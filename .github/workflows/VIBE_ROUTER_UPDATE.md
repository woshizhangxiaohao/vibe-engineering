# Vibe Router 更新说明

## 问题描述

之前的 Vibe Router 只在 issue 创建时（`opened` 事件）触发，当用户手动给已存在的 issue 打上标签时，不会自动触发路由分析和 Agent。

相关 Issue: [#265](https://github.com/lessthanno/vibe-engineering-playbook/issues/265)

## 解决方案

**触发方式已优化**：现在 Vibe Router 只在手动添加 `needs-route` 标签时触发。

### 为什么改为手动触发？
- ✅ 避免误触发
- ✅ 让开发者先整理需求
- ✅ 不需要跳过标签了
- ✅ 更清晰的控制流程

### 如何使用
1. 创建 Issue（不会自动触发任何操作）
2. 慢慢完善需求描述
3. 准备好后，添加 `needs-route` 标签
4. Router 自动分析复杂度并路由到对应 Agent

## 智能防重复机制

为了避免重复触发，系统会自动跳过以下情况：

### 跳过条件 1：已有复杂度标签
如果 issue 已经有以下任何标签，说明已经被路由过了：
- `complexity:simple`
- `complexity:medium`
- `complexity:complex`
- 任何以 `complexity:` 开头的标签

### 跳过条件 2：已被 AI 处理过
如果 issue 已经有以下任何标签，说明已经在处理中或已处理：
- `ai:processing` - AI 正在处理中
- `ai:completed` - AI 已完成
- `ai:failed` - AI 处理失败

### 跳过条件 3：非触发标签
只有添加 `needs-route` 标签时才会触发。
添加其他标签（如 `backend`、`frontend`、`enhancement`、`bug` 等）不会触发路由。

## 使用场景

### 场景 1：Issue 失败后重新路由
```
1. Issue #265 之前处理失败，有标签: ❌ ai-failed, backend, frontend
2. 你手动移除 ❌ ai-failed 标签
3. 添加 needs-route 标签
4. Vibe Router 自动触发，重新分析复杂度并路由到对应的 Agent
```

### 场景 2：手动分类后触发
```
1. 创建新 issue，但不想立即自动处理
2. 你先手动分析，确定这是一个后端任务
3. 添加 backend 标签
4. Vibe Router 自动触发，分析复杂度并路由
```

### 场景 3：补充标签触发
```
1. Issue 创建时没有任何标签
2. 后来你补充了 frontend 标签
3. Vibe Router 自动触发分析
```

## 工作流程

```
Issue 打上标签
    ↓
检查是否是触发标签？
    ↓ 是
检查是否已有复杂度标签？
    ↓ 否
检查是否已被 AI 处理？
    ↓ 否
触发路由分析
    ↓
AI 分析复杂度 (S/M/L)
    ↓
添加复杂度标签
    ↓
触发对应的 Agent
```

## 日志示例

### 触发路由的日志
```
📋 检测到标签添加: backend
✅ 满足路由条件，将触发路由分析
```

### 跳过路由的日志
```
📋 检测到标签添加: backend
⏭️ 跳过: 已有复杂度标签，不重复路由
```

或

```
📋 检测到标签添加: enhancement
⏭️ 跳过: 标签 enhancement 不是路由触发标签
```

## 配置

触发标签配置已简化，只保留 `needs-route` 作为唯一触发标签。

`backend`、`frontend`、`database` 现在仅作为范围标签，由 Router 自动添加，不再触发路由。

## 注意事项

1. **不会重复路由**：一旦 issue 被路由过（有复杂度标签），即使再次添加触发标签也不会重新路由
2. **需要手动清理**：如果想重新路由，需要先手动移除复杂度标签和 AI 处理标签
3. **只响应特定标签**：不是所有标签都会触发，只有配置的触发标签才会触发
4. **安全机制**：防止误触发导致的资源浪费和重复处理

## 修改的文件

- `.github/workflows/vibe-router.yml`
  - 添加 `labeled` 事件支持
  - 添加智能判断逻辑，防止重复触发
  - 优化条件检查，确保只在必要时触发

## 测试建议

1. 创建一个测试 issue
2. 手动添加 `backend` 标签
3. 观察 Actions 运行情况
4. 检查是否正确添加了复杂度标签
5. 再次添加标签，确认不会重复触发
