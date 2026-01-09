# 前端项目架构文档

> **📌 重要提示**：本文档是专门为 AI 代码生成工具设计的架构参考文档。它详细描述了项目的结构、规范、最佳实践和代码生成约束，旨在帮助 AI 理解项目架构并生成符合规范的代码。

本文档描述了前端项目的完整模块化架构，包括：

- 📁 项目结构和目录组织
- 🏗️ 模块功能说明和使用示例
- 📋 文件命名和代码规范
- 🔗 导入路径和模块导出规范
- 🧩 组件开发规范（客户端/服务端）
- 🔌 API 服务开发模式
- 🎣 Hook 开发规范
- 🎨 样式编写规范
- ⚠️ 错误处理模式
- 📝 类型定义规范
- 🚀 代码生成约束和检查清单
- 📖 快速参考表

**使用建议**：在生成代码前，请仔细阅读相关章节，确保生成的代码符合项目规范。

## 📁 项目结构

```
frontend/
├── app/                          # Next.js App Router
│   ├── layout.tsx               # 根布局
│   ├── page.tsx                 # 首页
│   ├── error.tsx                # 全局错误页面
│   ├── not-found.tsx            # 404 页面
│   ├── loading.tsx              # 全局加载页面
│   └── globals.css              # 全局样式
│
├── components/                   # React 组件
│   ├── ui/                      # shadcn/ui 组件库（53个组件）
│   │                            # ⚠️ 重要：此目录下的文件不允许修改，只能引用
│   │                            # 如需修改样式或行为，请在外部通过 className 或包装组件实现
│   ├── layout/                  # 布局组件
│   │   └── main-layout.tsx
│   ├── error-boundary.tsx       # 错误边界组件
│   └── loading.tsx             # 加载组件
│
├── hooks/                        # 自定义 Hooks
│   ├── use-debounce.ts          # 防抖 Hook
│   ├── use-local-storage.ts     # 本地存储 Hook
│   ├── use-media-query.ts       # 媒体查询 Hook
│   ├── use-click-outside.ts     # 点击外部区域 Hook
│   ├── use-mobile.tsx           # 移动端检测 Hook
│   └── index.ts                 # 统一导出
│
├── lib/                          # 核心库
│   ├── api/                     # API 服务层
│   │   ├── client.ts            # HTTP 客户端
│   │   ├── config.ts            # API 配置
│   │   ├── types.ts             # API 类型定义
│   │   ├── services/           # API 服务
│   │   │   └── pomodoro.service.ts
│   │   ├── hooks/               # API Hooks
│   │   │   └── use-pomodoro.ts
│   │   └── index.ts             # 统一导出
│   │
│   ├── config/                  # 配置模块
│   │   └── env.ts               # 环境配置
│   │
│   ├── constants/               # 常量定义
│   │   └── index.ts             # 应用常量
│   │
│   └── utils/                   # 工具函数
│       ├── utils.ts             # 基础工具（cn 函数等）
│       ├── toast.ts             # Toast 工具
│       ├── format.ts             # 格式化工具
│       ├── validation.ts         # 验证工具
│       ├── storage.ts            # 存储工具
│       ├── date.ts               # 日期工具
│       └── index.ts              # 统一导出
│
├── types/                        # 全局类型定义
│   └── index.ts
│
├── middleware.ts                  # Next.js 中间件
├── components.json              # shadcn/ui 配置
├── tailwind.config.ts           # Tailwind 配置
├── tsconfig.json                # TypeScript 配置
└── package.json                 # 依赖配置
```

## 🏗️ 模块说明

### 1. 核心框架 ✅
- ✅ Next.js 16 (App Router)
- ✅ React 19
- ✅ TypeScript
- ✅ Tailwind CSS v4
- ✅ shadcn/ui (53个组件)

### 2. API 服务层 (`lib/api/`) ✅

**功能**: 统一的 API 请求管理

**已完成模块**:
- ✅ HTTP 客户端 (`client.ts`) - 支持拦截器、错误处理、超时控制
- ✅ API 配置 (`config.ts`) - baseURL、端点等基础配置
- ✅ 类型定义 (`types.ts`) - API 相关类型定义
- ✅ Pomodoro 服务示例 (`services/pomodoro.service.ts`)
- ✅ API Hooks (`hooks/use-pomodoro.ts`)

**使用示例**:
```tsx
import { apiClient, pomodoroService } from "@/lib/api";
import { usePomodoros } from "@/lib/api/hooks";

// 使用服务层
const data = await pomodoroService.list();

// 使用 Hook
const { pomodoros, loading, error } = usePomodoros();
```

### 3. 工具函数库 (`lib/utils/`) ✅

**功能**: 通用工具函数集合

**已完成模块**:
- ✅ 基础工具 (`utils.ts`) - cn 函数等
- ✅ Toast 工具 (`toast.ts`) - 通知提示
- ✅ 格式化工具 (`format.ts`) - 文件大小、货币、百分比等
- ✅ 验证工具 (`validation.ts`) - 邮箱、手机号、密码等
- ✅ 存储工具 (`storage.ts`) - localStorage/sessionStorage
- ✅ 日期工具 (`date.ts`) - 日期格式化、相对时间等

**使用示例**:
```tsx
// 统一导入（推荐）
import { cn, formatFileSize, isValidEmail, setAuthToken, formatDate, toast } from "@/lib/utils";

// 或按需导入
import { formatFileSize, formatCurrency } from "@/lib/utils/format";
import { isValidEmail, isValidPhone } from "@/lib/utils/validation";
import { setAuthToken, getAuthToken } from "@/lib/utils/storage";
import { formatDate, getRelativeTime } from "@/lib/utils/date";
import { toast } from "@/lib/utils/toast";
```

### 4. 配置模块 ✅

**常量配置 (`lib/constants/`)** ✅
- ✅ 应用常量集中管理
- ✅ 包含：应用信息、路由路径、存储键名、HTTP 状态码、分页配置、日期格式、文件上传配置、验证规则、防抖延迟时间

**环境配置 (`lib/config/env.ts`)** ✅
- ✅ 类型安全的环境变量访问
- ✅ 环境验证
- ✅ 开发/生产环境区分

**使用示例**:
```tsx
import { ROUTES, STORAGE_KEYS, PAGINATION } from "@/lib/constants";
import { env } from "@/lib/config/env";

const apiUrl = env.API_URL;
const isDev = env.IS_DEV;
```

### 5. 类型定义 (`types/`) ✅

**功能**: 全局 TypeScript 类型定义

**包含**:
- ✅ API 响应类型
- ✅ 分页类型
- ✅ 用户类型
- ✅ 文件上传类型
- ✅ 选择项类型
- ✅ 表格列配置类型
- ✅ 菜单项类型

**使用示例**:
```tsx
import type { User, ApiResponse, PaginatedResponse } from "@/types";
```

### 6. Hooks (`hooks/`) ✅

**功能**: 可复用的 React Hooks

**已完成 Hooks**:
- ✅ `useDebounce` - 防抖
- ✅ `useLocalStorage` - 本地存储
- ✅ `useMediaQuery` - 媒体查询
- ✅ `useClickOutside` - 点击外部区域
- ✅ `useIsMobile` - 移动端检测
- ✅ `useIsTablet` - 平板检测
- ✅ `useIsDesktop` - 桌面端检测

**使用示例**:
```tsx
import { useDebounce, useLocalStorage, useIsMobile } from "@/hooks";

const [value, setValue] = useLocalStorage("key", "default");
const debouncedValue = useDebounce(value, 300);
const isMobile = useIsMobile();
```

### 7. 组件库 (`components/`) ✅

**功能**: React 组件库

**已完成组件**:
- ✅ shadcn/ui 组件 (53个)
- ✅ 布局组件 (`components/layout/main-layout.tsx`)
- ✅ 错误边界 (`components/error-boundary.tsx`)
- ✅ 加载组件 (`components/loading.tsx`)

**⚠️ 重要约束**:
- **`components/ui/` 目录下的所有文件不允许修改**
- 这些文件是 shadcn/ui 组件库的核心文件，只能引用使用
- 如需修改样式：通过 `className` prop 在外部覆盖样式
- 如需修改行为：创建包装组件在外部扩展功能
- 如需更新组件：使用 `npx shadcn-ui@latest add [component]` 命令重新生成

**使用示例**:
```tsx
import { Button } from "@/components/ui/button";
import { ErrorBoundary } from "@/components/error-boundary";
import { Loading } from "@/components/loading";

// ✅ 正确 - 通过 className 修改样式
<Button className="w-full bg-custom-color">Click</Button>

// ✅ 正确 - 创建包装组件扩展功能
function CustomButton({ children, ...props }) {
  return (
    <Button {...props} className="custom-styles">
      {children}
    </Button>
  );
}
```

### 8. Next.js 特殊页面 (`app/`) ✅

**功能**: Next.js App Router 特殊页面

**已完成页面**:
- ✅ 全局错误页面 (`app/error.tsx`)
- ✅ 404 页面 (`app/not-found.tsx`)
- ✅ 全局加载页面 (`app/loading.tsx`)

### 9. 中间件 (`middleware.ts`) ✅

**功能**: Next.js 中间件，用于请求拦截、认证、重定向等

## 🔧 核心特性

### ✅ 类型安全
- 完整的 TypeScript 类型定义
- 类型安全的 API 调用
- 类型安全的工具函数

### ✅ 模块化设计
- 清晰的模块划分
- 统一的导出接口
- 易于扩展和维护
- 每个模块都有统一的导出接口
- 模块职责单一，易于测试和维护

### ✅ 错误处理
- 统一的错误处理机制
- 错误边界组件
- 全局错误页面

### ✅ 性能优化
- 防抖/节流 Hooks
- 代码分割
- 懒加载支持

### ✅ 开发体验
- 完整的类型提示
- 统一的代码风格
- 完善的工具函数

### ✅ 代码复用
- 丰富的工具函数库
- 可复用的 React Hooks
- 统一的组件库

## 📦 依赖管理

### 核心依赖
- **Next.js 16** - React 框架
- **React 19** - UI 库
- **TypeScript** - 类型系统
- **Tailwind CSS v4** - 样式框架
- **shadcn/ui** - 组件库

### 工具库
- **date-fns** - 日期处理
- **zod** - 数据验证
- **sonner** - Toast 通知
- **react-hook-form** - 表单处理

## 🚀 快速开始

### 1. 模块导出

所有模块都提供统一的导出接口，支持按需导入：

```tsx
// API 模块
import { apiClient, pomodoroService } from "@/lib/api";
import { usePomodoros } from "@/lib/api/hooks";

// 工具函数（统一导出）
import { cn, formatFileSize, isValidEmail, setAuthToken, formatDate, toast } from "@/lib/utils";

// Hooks
import { useDebounce, useLocalStorage, useIsMobile } from "@/hooks";

// 组件
import { Button } from "@/components/ui/button";
import { ErrorBoundary } from "@/components/error-boundary";
import { Loading } from "@/components/loading";

// 常量
import { ROUTES, STORAGE_KEYS, PAGINATION } from "@/lib/constants";

// 类型
import type { User, ApiResponse, PaginatedResponse } from "@/types";
```

### 2. 使用示例

```tsx
"use client";

import { useState } from "react";
import { useDebounce, useIsMobile } from "@/hooks";
import { pomodoroService } from "@/lib/api";
import { toast, cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

export default function MyComponent() {
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebounce(search, 300);
  const isMobile = useIsMobile();

  const handleClick = async () => {
    try {
      const data = await pomodoroService.list();
      toast.success("加载成功");
    } catch (error) {
      toast.error("加载失败");
    }
  };

  return (
    <div className={cn("container", isMobile && "px-4")}>
      <input 
        value={search} 
        onChange={(e) => setSearch(e.target.value)} 
      />
      <Button onClick={handleClick}>提交</Button>
    </div>
  );
}
```

## 📝 最佳实践

1. **统一导入**: 使用统一的导出接口，避免深层导入
2. **类型安全**: 始终使用 TypeScript 类型
3. **错误处理**: 使用统一的错误处理机制
4. **代码复用**: 优先使用已有的工具函数和 Hooks
5. **模块化**: 保持模块职责单一，易于测试和维护

## 🔄 扩展指南

### 添加新的 API 服务

1. 在 `lib/api/services/` 创建服务文件
2. 在 `lib/api/config.ts` 添加端点配置
3. 在 `lib/api/services/index.ts` 导出
4. 可选：创建对应的 Hook

### 添加新的工具函数

1. 在 `lib/utils/` 创建工具文件
2. 在 `lib/utils/index.ts` 导出
3. 添加类型定义和文档

### 添加新的 Hook

1. 在 `hooks/` 创建 Hook 文件
2. 在 `hooks/index.ts` 导出
3. 添加使用示例和文档

## 📋 文件命名规范

### 文件命名规则

1. **组件文件**
   - 使用 kebab-case：`user-profile.tsx`, `data-table.tsx`
   - UI 组件（shadcn/ui）：`button.tsx`, `input.tsx`, `dialog.tsx`
   - 布局组件：`main-layout.tsx`, `sidebar-layout.tsx`
   - 页面组件（app/）：`page.tsx`, `layout.tsx`, `error.tsx`, `loading.tsx`

2. **工具文件**
   - 使用 kebab-case：`format.ts`, `validation.ts`, `date.ts`
   - 服务文件：`*.service.ts`（如 `pomodoro.service.ts`）
   - Hook 文件：`use-*.ts` 或 `use-*.tsx`（如 `use-debounce.ts`, `use-mobile.tsx`）

3. **类型文件**
   - 统一使用 `index.ts` 作为类型导出文件
   - 类型定义在对应模块的 `types.ts` 或 `index.ts` 中

4. **配置文件**
   - 使用 kebab-case：`env.ts`, `config.ts`
   - 统一导出文件：`index.ts`

### 目录命名规则

- 使用 kebab-case：`lib/`, `components/`, `hooks/`
- 子目录：`lib/api/services/`, `components/ui/`, `components/layout/`

## 🔗 导入路径规范

### 路径别名配置

项目使用 `@/` 作为根路径别名（配置在 `tsconfig.json` 中）：
```json
{
  "paths": {
    "@/*": ["./*"]
  }
}
```

### 导入规则

1. **统一使用路径别名**
   ```tsx
   // ✅ 正确
   import { Button } from "@/components/ui/button";
   import { useDebounce } from "@/hooks";
   import { apiClient } from "@/lib/api";
   
   // ❌ 错误 - 不要使用相对路径
   import { Button } from "../../components/ui/button";
   ```

2. **优先使用统一导出**
   ```tsx
   // ✅ 推荐 - 使用统一导出
   import { cn, toast, formatDate } from "@/lib/utils";
   import { useDebounce, useIsMobile } from "@/hooks";
   
   // ⚠️ 允许 - 但优先使用统一导出
   import { cn } from "@/lib/utils/utils";
   import { toast } from "@/lib/utils/toast";
   ```

3. **类型导入使用 `type` 关键字**
   ```tsx
   // ✅ 正确
   import type { User, ApiResponse } from "@/types";
   import type { Pomodoro } from "@/lib/api/services/pomodoro.service";
   
   // ✅ 也正确 - 混合导入
   import { apiClient, type ApiError } from "@/lib/api";
   ```

4. **组件导入顺序**
   ```tsx
   // 1. React 和 Next.js
   import { useState, useEffect } from "react";
   import { useRouter } from "next/navigation";
   
   // 2. 第三方库
   import { format } from "date-fns";
   
   // 3. 项目模块（按字母顺序）
   import { Button } from "@/components/ui/button";
   import { useDebounce } from "@/hooks";
   import { apiClient } from "@/lib/api";
   import { toast } from "@/lib/utils";
   ```

## 🧩 组件开发规范

### 客户端组件 vs 服务端组件

1. **服务端组件（默认）**
   - Next.js App Router 默认所有组件都是服务端组件
   - 不需要 `"use client"` 指令
   - 可以直接访问数据库、API 等
   - 不能使用浏览器 API（如 `useState`, `useEffect`）

2. **客户端组件（需要交互时）**
   - 必须添加 `"use client"` 指令在文件顶部
   - 可以使用所有 React Hooks
   - 可以使用浏览器 API
   - 示例：
   ```tsx
   "use client";
   
   import { useState } from "react";
   import { Button } from "@/components/ui/button";
   
   export function Counter() {
     const [count, setCount] = useState(0);
     return <Button onClick={() => setCount(count + 1)}>{count}</Button>;
   }
   ```

### 组件结构规范

```tsx
/**
 * 组件文件头注释（可选但推荐）
 */

"use client"; // 如果需要客户端组件

// 1. 导入顺序：React → Next.js → 第三方 → 项目模块
import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { format } from "date-fns";
import { Button } from "@/components/ui/button";
import { useDebounce } from "@/hooks";
import { toast } from "@/lib/utils";

// 2. 类型定义
interface ComponentProps {
  title: string;
  optional?: boolean;
}

// 3. 组件实现
export function Component({ title, optional = false }: ComponentProps) {
  // Hooks
  const [state, setState] = useState("");
  const debouncedValue = useDebounce(state, 300);
  
  // 事件处理函数
  const handleClick = () => {
    toast.success("Clicked!");
  };
  
  // 渲染
  return (
    <div>
      <h1>{title}</h1>
      <Button onClick={handleClick}>Click</Button>
    </div>
  );
}

// 4. 默认导出（如果使用 default export）
export default Component;
```

### 组件命名规范

1. **组件名称使用 PascalCase**
   ```tsx
   // ✅ 正确
   export function UserProfile() {}
   export function DataTable() {}
   
   // ❌ 错误
   export function userProfile() {}
   export function data_table() {}
   ```

2. **Props 接口命名**
   ```tsx
   // ✅ 推荐 - 组件名 + Props
   interface UserProfileProps {}
   interface DataTableProps {}
   
   // ✅ 也接受 - 直接使用组件名
   interface UserProfile {}
   ```

3. **导出方式**
   ```tsx
   // ✅ 推荐 - 命名导出
   export function UserProfile() {}
   
   // ✅ 也接受 - 默认导出（页面组件常用）
   export default function UserProfile() {}
   ```

## 🔌 API 服务开发规范

### 服务类结构

```tsx
/**
 * Service 文件头注释
 */

import { apiClient } from "../client";
import { API_ENDPOINTS } from "../config";
import { ApiResponse, PaginatedResponse } from "../types";

// 1. 类型定义
export interface Entity {
  id: number;
  name: string;
  created_at: string;
}

export interface CreateEntityRequest {
  name: string;
}

export interface UpdateEntityRequest {
  name?: string;
}

// 2. 服务类
class EntityService {
  /**
   * 创建实体
   */
  async create(data: CreateEntityRequest): Promise<Entity> {
    const response = await apiClient.post<ApiResponse<Entity>>(
      API_ENDPOINTS.ENTITY.CREATE,
      data
    );
    return response.data || (response as unknown as Entity);
  }

  /**
   * 获取列表
   */
  async list(params?: {
    page?: number;
    pageSize?: number;
  }): Promise<Entity[] | PaginatedResponse<Entity>> {
    const response = await apiClient.get<ApiResponse<Entity[]>>(
      API_ENDPOINTS.ENTITY.LIST,
      { params }
    );
    return response.data || (response as unknown as Entity[]);
  }

  /**
   * 根据 ID 获取
   */
  async getById(id: string | number): Promise<Entity> {
    const response = await apiClient.get<ApiResponse<Entity>>(
      API_ENDPOINTS.ENTITY.GET(id)
    );
    return response.data || (response as unknown as Entity);
  }

  /**
   * 更新
   */
  async update(
    id: string | number,
    data: UpdateEntityRequest
  ): Promise<Entity> {
    const response = await apiClient.put<ApiResponse<Entity>>(
      API_ENDPOINTS.ENTITY.UPDATE(id),
      data
    );
    return response.data || (response as unknown as Entity);
  }

  /**
   * 删除
   */
  async delete(id: string | number): Promise<void> {
    await apiClient.delete(API_ENDPOINTS.ENTITY.DELETE(id));
  }
}

// 3. 导出单例实例
export const entityService = new EntityService();

// 4. 导出类（可选）
export default EntityService;
```

### 端点配置规范

在 `lib/api/config.ts` 中添加端点：

```tsx
export const API_ENDPOINTS = {
  // 现有端点...
  
  // 新实体端点
  ENTITY: {
    BASE: "/entities",
    CREATE: "/entities",
    LIST: "/entities",
    GET: (id: string | number) => `/entities/${id}`,
    UPDATE: (id: string | number) => `/entities/${id}`,
    DELETE: (id: string | number) => `/entities/${id}`,
  },
} as const;
```

### API Hook 开发规范

```tsx
/**
 * Entity React Hook
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { entityService, ApiError } from "../index";
import type { Entity, CreateEntityRequest, UpdateEntityRequest } from "../services/entity.service";

/**
 * 使用 Entity 列表的 Hook
 */
export function useEntities() {
  const [entities, setEntities] = useState<Entity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchEntities = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await entityService.list();
      setEntities(Array.isArray(data) ? data : data.data || []);
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Failed to fetch entities"));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchEntities();
  }, [fetchEntities]);

  return {
    entities,
    loading,
    error,
    refetch: fetchEntities,
  };
}

/**
 * 使用单个 Entity 的 Hook
 */
export function useEntity(id: string | number | null) {
  const [entity, setEntity] = useState<Entity | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchEntity = useCallback(async () => {
    if (!id) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const data = await entityService.getById(id);
      setEntity(data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Failed to fetch entity"));
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchEntity();
  }, [fetchEntity]);

  return {
    entity,
    loading,
    error,
    refetch: fetchEntity,
  };
}

/**
 * 创建 Entity 的 Hook
 */
export function useCreateEntity() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const createEntity = useCallback(async (data: CreateEntityRequest) => {
    try {
      setLoading(true);
      setError(null);
      const result = await entityService.create(data);
      return result;
    } catch (err) {
      const apiError = err instanceof ApiError ? err : new Error("Failed to create entity");
      setError(apiError);
      throw apiError;
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    createEntity,
    loading,
    error,
  };
}

// 类似地实现 useUpdateEntity 和 useDeleteEntity...
```

## 🎣 Hook 开发规范

### Hook 命名规范

1. **必须以 `use` 开头**
   ```tsx
   // ✅ 正确
   export function useDebounce() {}
   export function useLocalStorage() {}
   
   // ❌ 错误
   export function debounce() {}
   export function getLocalStorage() {}
   ```

2. **文件命名**
   - 使用 kebab-case：`use-debounce.ts`, `use-local-storage.ts`
   - 如果包含 JSX，使用 `.tsx`：`use-mobile.tsx`

### Hook 结构规范

```tsx
/**
 * Hook 文件头注释
 */

"use client"; // 如果需要客户端功能

import { useState, useEffect, useCallback } from "react";

/**
 * Hook 功能描述
 * 
 * @param value - 参数描述
 * @param delay - 延迟时间（毫秒）
 * @returns 返回值描述
 */
export function useCustomHook<T>(value: T, delay: number = 300) {
  // 1. 状态定义
  const [state, setState] = useState<T>(value);
  
  // 2. 副作用
  useEffect(() => {
    // 副作用逻辑
  }, [dependencies]);
  
  // 3. 回调函数
  const handleAction = useCallback(() => {
    // 回调逻辑
  }, [dependencies]);
  
  // 4. 返回值
  return {
    state,
    handleAction,
  };
}
```

### Hook 导出规范

在 `hooks/index.ts` 中统一导出：

```tsx
/**
 * Hooks 统一导出
 */

export * from "./use-debounce";
export * from "./use-local-storage";
export * from "./use-custom-hook";
```

## 🎨 样式编写规范

### Tailwind CSS 使用规范

1. **优先使用 Tailwind 工具类**
   ```tsx
   // ✅ 推荐
   <div className="flex items-center justify-between p-4 bg-white rounded-lg shadow-md">
   
   // ❌ 避免内联样式
   <div style={{ display: 'flex', padding: '16px' }}>
   ```

2. **使用 `cn` 函数合并类名**
   ```tsx
   import { cn } from "@/lib/utils";
   
   <div className={cn(
     "base-classes",
     condition && "conditional-classes",
     className // 支持外部传入的 className
   )}>
   ```

3. **响应式设计**
   ```tsx
   <div className="
     w-full 
     md:w-1/2 
     lg:w-1/3
     p-4 
     md:p-6 
     lg:p-8
   ">
   ```

4. **暗色模式支持**
   ```tsx
   <div className="bg-white dark:bg-gray-900 text-gray-900 dark:text-white">
   ```

### 组件样式规范

1. **UI 组件样式**
   - shadcn/ui 组件已包含基础样式
   - **⚠️ 禁止直接修改 `components/ui/` 目录下的组件文件**
   - 通过 `className` prop 在外部扩展样式
   ```tsx
   // ✅ 正确 - 通过 className 扩展样式
   <Button className="w-full md:w-auto">Click</Button>
   
   // ✅ 正确 - 使用 cn 函数合并样式
   import { cn } from "@/lib/utils";
   <Button className={cn("base-classes", customClasses)}>Click</Button>
   ```

2. **布局组件样式**
   - 使用 Tailwind 的布局工具类
   - 保持响应式设计
   ```tsx
   <div className="container mx-auto px-4 md:px-6 lg:px-8">
   ```

## ⚠️ 错误处理规范

### API 错误处理

```tsx
import { apiClient, ApiError } from "@/lib/api";
import { toast } from "@/lib/utils";

try {
  const data = await apiClient.get("/endpoint");
  toast.success("操作成功");
} catch (error) {
  if (error instanceof ApiError) {
    // API 错误
    if (error.status === 401) {
      toast.error("未授权，请重新登录");
      // 跳转到登录页
    } else if (error.status === 404) {
      toast.error("资源不存在");
    } else {
      toast.error(error.message || "操作失败");
    }
  } else {
    // 网络错误或其他错误
    toast.error("网络错误，请稍后重试");
  }
}
```

### 组件错误处理

```tsx
"use client";

import { ErrorBoundary } from "@/components/error-boundary";
import { toast } from "@/lib/utils";

function MyComponent() {
  const handleAction = async () => {
    try {
      await someAsyncOperation();
    } catch (error) {
      toast.error("操作失败");
      console.error(error);
    }
  };
  
  return <button onClick={handleAction}>Action</button>;
}

// 使用错误边界包裹
export default function Page() {
  return (
    <ErrorBoundary>
      <MyComponent />
    </ErrorBoundary>
  );
}
```

## 📝 类型定义规范

### 类型文件组织

1. **全局类型** → `types/index.ts`
   ```tsx
   // types/index.ts
   export interface User {
     id: number;
     name: string;
     email: string;
   }
   
   export interface ApiResponse<T> {
     data: T;
     message?: string;
     code?: number;
   }
   ```

2. **模块类型** → 在对应模块中定义
   ```tsx
   // lib/api/services/entity.service.ts
   export interface Entity {
     id: number;
     name: string;
   }
   ```

3. **组件 Props 类型** → 在组件文件中定义
   ```tsx
   // components/user-profile.tsx
   interface UserProfileProps {
     userId: number;
     showEmail?: boolean;
   }
   ```

### 类型命名规范

1. **接口使用 PascalCase**
   ```tsx
   interface UserProfile {}
   interface ApiResponse<T> {}
   ```

2. **类型别名使用 PascalCase**
   ```tsx
   type UserId = number;
   type Status = "pending" | "completed" | "failed";
   ```

3. **泛型参数使用单个大写字母**
   ```tsx
   interface ApiResponse<T> {}
   function useEntity<T extends Entity>() {}
   ```

## 🔐 环境变量使用规范

### 环境变量命名

1. **客户端环境变量必须以 `NEXT_PUBLIC_` 开头**
   ```bash
   # .env.local
   NEXT_PUBLIC_API_URL=http://localhost:8080
   NEXT_PUBLIC_APP_NAME=VibeFlow
   ```

2. **服务端环境变量不需要前缀**
   ```bash
   # .env.local
   DATABASE_URL=postgresql://...
   SECRET_KEY=...
   ```

### 环境变量访问

使用 `lib/config/env.ts` 统一访问：

```tsx
import { env } from "@/lib/config/env";

// ✅ 正确
const apiUrl = env.API_URL;
const isDev = env.IS_DEV;

// ❌ 错误 - 不要直接访问 process.env
const apiUrl = process.env.NEXT_PUBLIC_API_URL;
```

## 🤖 AI 代码生成工作流

### 生成代码前的准备

1. **理解需求**
   - 明确要生成的功能模块类型（组件/Hook/API 服务/工具函数）
   - 确定文件应该放在哪个目录
   - 了解是否需要客户端组件（需要交互）

2. **查阅相关章节**
   - 组件开发 → 查看"组件开发规范"
   - API 服务 → 查看"API 服务开发规范"
   - Hook → 查看"Hook 开发规范"
   - 工具函数 → 查看"工具函数库"章节

3. **参考示例代码**
   - 查看文档中的代码示例
   - 参考系统文件列表中的实际文件

4. **遵循命名规范**
   - 文件命名：kebab-case
   - 组件命名：PascalCase
   - Hook 命名：useXxx

5. **使用正确的导入路径**
   - 统一使用 `@/` 别名
   - 优先使用统一导出接口

### 代码生成步骤

1. **创建文件**
   - 确定文件路径和名称
   - 创建文件并添加必要的注释

2. **添加导入**
   - 按照导入顺序添加必要的导入
   - 使用路径别名 `@/`

3. **实现功能**
   - 遵循模块的代码结构规范
   - 添加类型定义
   - 实现核心逻辑

4. **添加导出**
   - 如果是新模块，在对应的 `index.ts` 中添加导出
   - 确保导出符合统一导出规范

5. **检查清单**
   - 运行代码生成检查清单
   - 确保没有常见错误

### 代码生成示例流程

**场景**：需要创建一个新的用户管理 API 服务

1. **确定位置**：`lib/api/services/user.service.ts`

2. **参考模式**：参考 `pomodoro.service.ts` 的结构

3. **添加端点配置**：在 `lib/api/config.ts` 中添加 `USER` 端点

4. **创建服务类**：
   ```tsx
   // lib/api/services/user.service.ts
   import { apiClient } from "../client";
   import { API_ENDPOINTS } from "../config";
   import { ApiResponse } from "../types";
   
   export interface User {
     id: number;
     name: string;
     email: string;
   }
   
   class UserService {
     async list(): Promise<User[]> {
       const response = await apiClient.get<ApiResponse<User[]>>(
         API_ENDPOINTS.USER.LIST
       );
       return response.data || [];
     }
     // ... 其他方法
   }
   
   export const userService = new UserService();
   ```

5. **导出服务**：在 `lib/api/services/index.ts` 中添加导出

6. **创建 Hook（可选）**：在 `lib/api/hooks/use-user.ts` 中创建对应的 Hook

7. **检查**：确保符合所有规范

## 🚀 代码生成约束

### 系统文件列表

以下文件定义了项目的核心架构模式，代码生成时必须参考：

<!-- AGENT_SYSTEM_FILES_START -->
```
frontend/app/globals.css
frontend/app/layout.tsx
frontend/app/page.tsx
frontend/components.json
frontend/components/AppContainer.tsx
frontend/components/AuthPanel.tsx
frontend/components/MetadataCard.tsx
frontend/components/QuotaMonitor.tsx
frontend/components/SearchInputGroup.tsx
frontend/components/Sidebar.tsx
frontend/components/SummaryPanel.tsx
frontend/components/TranscriptionPanel.tsx
frontend/components/VideoDetailView.tsx
frontend/components/YoutubeDashboard.tsx
frontend/components/ui/button.tsx
frontend/hooks/index.ts
frontend/lib/api/client.ts
frontend/lib/api/config.ts
frontend/lib/api/endpoints.ts
frontend/lib/api/types.ts
frontend/lib/config/env.ts
frontend/lib/constants/index.ts
frontend/lib/utils.ts
frontend/lib/utils/index.ts
frontend/lib/utils/toast.ts
frontend/middleware.ts
frontend/package.json
frontend/tailwind.config.ts
frontend/tsconfig.json
frontend/types/index.ts
frontend/types/video.ts
```
<!-- AGENT_SYSTEM_FILES_END -->

### 代码生成规则

1. **遵循模块化架构**
   - API 服务层 → `lib/api/`
   - 工具函数 → `lib/utils/`
   - 组件 → `components/`
   - Hooks → `hooks/`
   - 类型定义 → `types/`

2. **路径约束**
   - 所有文件必须在 `frontend/` 目录下
   - 遵循现有目录结构
   - 不允许在 `frontend/` 目录外创建文件

3. **代码风格**
   - 使用 TypeScript 严格模式
   - 遵循现有代码风格
   - 使用统一的导入路径别名 `@/`

4. **组件规范**
   - 使用 shadcn/ui 组件库
   - **⚠️ 禁止修改 `components/ui/` 目录下的任何文件**
   - 只能引用使用，如需修改请在外部通过 `className` 或包装组件实现
   - 遵循组件命名规范
   - 使用统一的样式方案（Tailwind CSS）

5. **API 调用**
   - 使用统一的 API 客户端
   - 遵循 API 服务层模式
   - 使用类型安全的 API 调用

### 代码生成检查清单

生成代码时，请确保：

- [ ] 文件命名符合规范（kebab-case）
- [ ] 导入路径使用 `@/` 别名
- [ ] 优先使用统一导出接口
- [ ] 类型导入使用 `type` 关键字
- [ ] 客户端组件添加 `"use client"` 指令
- [ ] 组件使用 PascalCase 命名
- [ ] API 服务遵循服务类模式
- [ ] Hook 以 `use` 开头
- [ ] 错误处理完善
- [ ] 样式使用 Tailwind CSS
- [ ] 环境变量通过 `env` 对象访问
- [ ] 添加必要的类型定义
- [ ] 添加必要的注释和文档
- [ ] **禁止修改 `components/ui/` 目录下的任何文件**

### 常见错误避免

1. **❌ 不要使用相对路径导入**
   ```tsx
   // ❌ 错误
   import { Button } from "../../components/ui/button";
   
   // ✅ 正确
   import { Button } from "@/components/ui/button";
   ```

2. **❌ 不要在服务端组件中使用客户端 API**
   ```tsx
   // ❌ 错误 - 服务端组件不能使用 useState
   export function ServerComponent() {
     const [state, setState] = useState(0); // 错误！
   }
   
   // ✅ 正确 - 添加 "use client"
   "use client";
   export function ClientComponent() {
     const [state, setState] = useState(0);
   }
   ```

3. **❌ 不要直接访问 process.env**
   ```tsx
   // ❌ 错误
   const apiUrl = process.env.NEXT_PUBLIC_API_URL;
   
   // ✅ 正确
   import { env } from "@/lib/config/env";
   const apiUrl = env.API_URL;
   ```

4. **❌ 不要创建重复的工具函数**
   ```tsx
   // ❌ 错误 - 已有 cn 函数
   function mergeClasses(...classes: string[]) {}
   
   // ✅ 正确 - 使用已有工具
   import { cn } from "@/lib/utils";
   ```

5. **❌ 不要绕过 API 服务层**
   ```tsx
   // ❌ 错误 - 直接使用 fetch
   const data = await fetch("/api/endpoint");
   
   // ✅ 正确 - 使用 API 客户端
   import { apiClient } from "@/lib/api";
   const data = await apiClient.get("/endpoint");
   ```

6. **❌ 不要修改 `components/ui/` 目录下的文件**
   ```tsx
   // ❌ 错误 - 直接修改 UI 组件库文件
   // 修改 components/ui/button.tsx
   
   // ✅ 正确 - 通过 className 修改样式
   import { Button } from "@/components/ui/button";
   <Button className="w-full bg-custom-color">Click</Button>
   
   // ✅ 正确 - 创建包装组件扩展功能
   function CustomButton({ children, ...props }) {
     return (
       <Button {...props} className="custom-styles">
         {children}
       </Button>
     );
   }
   ```

## 🔄 状态管理规范

### 本地状态管理

1. **使用 React Hooks**
   ```tsx
   // ✅ 简单状态
   const [count, setCount] = useState(0);
   
   // ✅ 复杂状态
   const [state, setState] = useState({
     name: "",
     email: "",
   });
   
   // ✅ 使用 useReducer 处理复杂状态逻辑
   const [state, dispatch] = useReducer(reducer, initialState);
   ```

2. **使用自定义 Hooks 封装状态逻辑**
   ```tsx
   // hooks/use-form.ts
   export function useForm<T>(initialValues: T) {
     const [values, setValues] = useState<T>(initialValues);
     const [errors, setErrors] = useState<Partial<Record<keyof T, string>>>({});
     
     const setValue = (key: keyof T, value: T[keyof T]) => {
       setValues(prev => ({ ...prev, [key]: value }));
     };
     
     return { values, errors, setValue };
   }
   ```

### 服务端状态管理

使用 React Query 或 SWR（如果项目引入）：
```tsx
// 如果使用 React Query
import { useQuery, useMutation } from "@tanstack/react-query";
import { pomodoroService } from "@/lib/api";

export function usePomodoros() {
  return useQuery({
    queryKey: ["pomodoros"],
    queryFn: () => pomodoroService.list(),
  });
}
```

### 全局状态管理

如果项目需要全局状态，考虑：
- Context API（简单场景）
- Zustand（轻量级）
- Redux（复杂场景）

## 📝 表单处理规范

### 使用 react-hook-form

```tsx
"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "@/lib/utils";

// 1. 定义验证 schema
const formSchema = z.object({
  name: z.string().min(3, "名称至少3个字符"),
  email: z.string().email("无效的邮箱地址"),
});

type FormData = z.infer<typeof formSchema>;

// 2. 组件实现
export function UserForm() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
  });

  const onSubmit = async (data: FormData) => {
    try {
      await submitForm(data);
      toast.success("提交成功");
    } catch (error) {
      toast.error("提交失败");
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Input
        {...register("name")}
        placeholder="姓名"
        error={errors.name?.message}
      />
      <Input
        {...register("email")}
        type="email"
        placeholder="邮箱"
        error={errors.email?.message}
      />
      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? "提交中..." : "提交"}
      </Button>
    </form>
  );
}
```

## 🛣️ 路由使用规范

### Next.js App Router

1. **页面路由**
   ```tsx
   // app/users/page.tsx - /users
   export default function UsersPage() {
     return <div>Users Page</div>;
   }
   
   // app/users/[id]/page.tsx - /users/:id
   export default function UserDetailPage({ params }: { params: { id: string } }) {
     return <div>User {params.id}</div>;
   }
   ```

2. **路由跳转**
   ```tsx
   "use client";
   
   import { useRouter } from "next/navigation";
   import { Button } from "@/components/ui/button";
   
   export function NavigationButton() {
     const router = useRouter();
     
     const handleClick = () => {
       router.push("/users");
       // 或 router.replace("/users")
     };
     
     return <Button onClick={handleClick}>Go to Users</Button>;
   }
   ```

3. **路由常量使用**
   ```tsx
   import { ROUTES } from "@/lib/constants";
   
   router.push(ROUTES.USERS);
   ```

## ⚡ 性能优化规范

### 1. 代码分割

```tsx
// ✅ 使用动态导入
import dynamic from "next/dynamic";

const HeavyComponent = dynamic(() => import("@/components/heavy-component"), {
  loading: () => <div>Loading...</div>,
  ssr: false, // 如果需要禁用 SSR
});
```

### 2. 图片优化

```tsx
import Image from "next/image";

<Image
  src="/image.jpg"
  alt="Description"
  width={500}
  height={300}
  priority // 首屏图片
  placeholder="blur" // 模糊占位符
/>
```

### 3. 列表渲染优化

```tsx
import { useMemo } from "react";

function DataList({ items }: { items: Item[] }) {
  // ✅ 使用 useMemo 缓存计算结果
  const filteredItems = useMemo(() => {
    return items.filter(item => item.active);
  }, [items]);
  
  return (
    <ul>
      {filteredItems.map(item => (
        <li key={item.id}>{item.name}</li>
      ))}
    </ul>
  );
}
```

### 4. 防抖和节流

```tsx
import { useDebounce } from "@/hooks";

function SearchInput() {
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebounce(search, 300);
  
  useEffect(() => {
    if (debouncedSearch) {
      performSearch(debouncedSearch);
    }
  }, [debouncedSearch]);
  
  return <input value={search} onChange={(e) => setSearch(e.target.value)} />;
}
```

## 🧪 测试规范（如果项目包含测试）

### 测试文件组织

```
components/
  ├── user-profile.tsx
  └── __tests__/
      └── user-profile.test.tsx
```

### 测试示例

```tsx
import { render, screen } from "@testing-library/react";
import { UserProfile } from "@/components/user-profile";

describe("UserProfile", () => {
  it("renders user name", () => {
    render(<UserProfile name="John" />);
    expect(screen.getByText("John")).toBeInTheDocument();
  });
});
```

## 📦 依赖管理规范

### 添加新依赖

1. **安装依赖**
   ```bash
   npm install package-name
   # 或
   npm install -D package-name # 开发依赖
   ```

2. **更新 package.json**
   - 自动更新（推荐）
   - 手动更新版本号（如需要特定版本）

3. **类型定义**
   ```bash
   # 如果包没有类型定义
   npm install -D @types/package-name
   ```

### 依赖分类

- **核心依赖**：Next.js, React, TypeScript
- **UI 库**：shadcn/ui, Tailwind CSS
- **工具库**：date-fns, zod, react-hook-form
- **开发工具**：ESLint, TypeScript

## 🔍 调试规范

### 开发环境调试

1. **使用 console.log（开发环境）**
   ```tsx
   if (process.env.NODE_ENV === "development") {
     console.log("Debug info:", data);
   }
   ```

2. **使用环境变量控制调试**
   ```tsx
   import { env } from "@/lib/config/env";
   
   if (env.ENABLE_DEBUG) {
     console.log("Debug info:", data);
   }
   ```

3. **API 客户端调试日志**
   - API 客户端已在开发环境自动输出请求日志
   - 查看浏览器控制台或终端

### 错误追踪

```tsx
try {
  await someOperation();
} catch (error) {
  // 记录错误（可以集成错误追踪服务）
  console.error("Operation failed:", error);
  
  // 用户友好的错误提示
  toast.error("操作失败，请稍后重试");
}
```

## 📖 快速参考表

### 导入路径速查

| 模块类型 | 导入示例 |
|---------|---------|
| UI 组件 | `import { Button } from "@/components/ui/button"` |
| 布局组件 | `import { MainLayout } from "@/components/layout/main-layout"` |
| Hooks | `import { useDebounce, useIsMobile } from "@/hooks"` |
| API 服务 | `import { apiClient, pomodoroService } from "@/lib/api"` |
| API Hooks | `import { usePomodoros } from "@/lib/api/hooks"` |
| 工具函数 | `import { cn, toast, formatDate } from "@/lib/utils"` |
| 常量 | `import { ROUTES, STORAGE_KEYS } from "@/lib/constants"` |
| 环境变量 | `import { env } from "@/lib/config/env"` |
| 类型定义 | `import type { User, ApiResponse } from "@/types"` |

### 文件命名速查

| 文件类型 | 命名规范 | 示例 |
|---------|---------|------|
| 组件文件 | kebab-case | `user-profile.tsx`, `data-table.tsx` |
| Hook 文件 | use-kebab-case | `use-debounce.ts`, `use-mobile.tsx` |
| 服务文件 | *.service.ts | `pomodoro.service.ts`, `user.service.ts` |
| 工具文件 | kebab-case | `format.ts`, `validation.ts`, `date.ts` |
| 类型文件 | index.ts | `types/index.ts` |
| 配置文件 | kebab-case | `config.ts`, `env.ts` |

### 组件类型速查

| 组件类型 | 是否需要 "use client" | 示例 |
|---------|---------------------|------|
| 页面组件 | 视情况而定 | `app/page.tsx` |
| 布局组件 | 视情况而定 | `app/layout.tsx` |
| 交互组件 | ✅ 需要 | `components/counter.tsx` |
| UI 组件 | ✅ 需要 | `components/ui/button.tsx` |
| 服务端组件 | ❌ 不需要 | 默认所有组件 |

### API 调用模式速查

| 场景 | 代码示例 |
|------|---------|
| GET 请求 | `await apiClient.get<Response>("/endpoint")` |
| POST 请求 | `await apiClient.post<Response>("/endpoint", data)` |
| PUT 请求 | `await apiClient.put<Response>("/endpoint", data)` |
| DELETE 请求 | `await apiClient.delete("/endpoint")` |
| 带参数 | `await apiClient.get("/endpoint", { params: { page: 1 } })` |
| 使用服务 | `await pomodoroService.list()` |
| 使用 Hook | `const { data, loading, error } = usePomodoros()` |

### 错误处理模式速查

| 场景 | 代码示例 |
|------|---------|
| API 错误 | `catch (error) { if (error instanceof ApiError) { ... } }` |
| 通用错误 | `catch (error) { toast.error("操作失败") }` |
| 错误边界 | `<ErrorBoundary><Component /></ErrorBoundary>` |

### 样式类名速查

| 功能 | Tailwind 类名 |
|------|-------------|
| 容器 | `container mx-auto px-4` |
| Flex 布局 | `flex items-center justify-between` |
| Grid 布局 | `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4` |
| 响应式 | `w-full md:w-1/2 lg:w-1/3` |
| 间距 | `p-4 m-2 gap-4` |
| 圆角 | `rounded-lg` |
| 阴影 | `shadow-md` |
| 暗色模式 | `dark:bg-gray-900 dark:text-white` |

### 常用 Hooks 速查

| Hook | 用途 | 导入 |
|------|------|------|
| useDebounce | 防抖 | `import { useDebounce } from "@/hooks"` |
| useLocalStorage | 本地存储 | `import { useLocalStorage } from "@/hooks"` |
| useIsMobile | 移动端检测 | `import { useIsMobile } from "@/hooks"` |
| useClickOutside | 点击外部 | `import { useClickOutside } from "@/hooks"` |

### 常用工具函数速查

| 函数 | 用途 | 导入 |
|------|------|------|
| cn | 合并类名 | `import { cn } from "@/lib/utils"` |
| toast | 通知提示 | `import { toast } from "@/lib/utils"` |
| formatDate | 日期格式化 | `import { formatDate } from "@/lib/utils"` |
| isValidEmail | 邮箱验证 | `import { isValidEmail } from "@/lib/utils"` |
| setAuthToken | 设置 token | `import { setAuthToken } from "@/lib/utils"` |

## 📚 相关文档

- [模块清单](./MODULES.md) - 完整的模块列表和状态
- [API 模块文档](./lib/api/README.md)
- [组件列表](./components/ui/COMPONENTS.md)
- [Next.js 文档](https://nextjs.org/docs)
- [shadcn/ui 文档](https://ui.shadcn.com)
- [React 文档](https://react.dev)
- [TypeScript 文档](https://www.typescriptlang.org/docs)
- [Tailwind CSS 文档](https://tailwindcss.com/docs)

