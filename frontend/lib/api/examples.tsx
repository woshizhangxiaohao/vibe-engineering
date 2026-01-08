/**
 * API 使用示例
 * 这些示例展示了如何使用 API 模块
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  pomodoroService,
  ApiError,
  usePomodoros,
  useCreatePomodoro,
  useUpdatePomodoro,
  useDeletePomodoro,
} from "@/lib/api";

/**
 * 示例 1: 使用服务层直接调用
 */
export function ServiceExample() {
  const [result, setResult] = useState<string>("");

  const handleCreate = async () => {
    try {
      const pomodoro = await pomodoroService.create({
        start_time: new Date().toISOString(),
        end_time: new Date(Date.now() + 25 * 60 * 1000).toISOString(),
        duration: 25,
        is_completed: false,
      });
      setResult(`Created: ${JSON.stringify(pomodoro, null, 2)}`);
    } catch (error) {
      if (error instanceof ApiError) {
        setResult(`Error: ${error.status} - ${error.message}`);
      } else {
        setResult(`Unknown error: ${error}`);
      }
    }
  };

  const handleList = async () => {
    try {
      const list = await pomodoroService.list();
      setResult(`List: ${JSON.stringify(list, null, 2)}`);
    } catch (error) {
      if (error instanceof ApiError) {
        setResult(`Error: ${error.status} - ${error.message}`);
      } else {
        setResult(`Unknown error: ${error}`);
      }
    }
  };

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-bold">服务层示例</h2>
      <div className="flex gap-2">
        <Button onClick={handleCreate}>创建 Pomodoro</Button>
        <Button onClick={handleList} variant="secondary">
          获取列表
        </Button>
      </div>
      {result && (
        <pre className="rounded-lg bg-muted p-4 text-sm">{result}</pre>
      )}
    </div>
  );
}

/**
 * 示例 2: 使用 React Hooks
 */
export function HooksExample() {
  const { pomodoros, loading, error, refetch } = usePomodoros();
  const { createPomodoro, loading: creating } = useCreatePomodoro();
  const { updatePomodoro, loading: updating } = useUpdatePomodoro();
  const { deletePomodoro, loading: deleting } = useDeletePomodoro();

  const handleCreate = async () => {
    try {
      await createPomodoro({
        start_time: new Date().toISOString(),
        end_time: new Date(Date.now() + 25 * 60 * 1000).toISOString(),
        duration: 25,
      });
      refetch();
    } catch (error) {
      console.error("Failed to create:", error);
    }
  };

  const handleToggle = async (id: number, current: boolean) => {
    try {
      await updatePomodoro(id, { is_completed: !current });
      refetch();
    } catch (error) {
      console.error("Failed to update:", error);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deletePomodoro(id);
      refetch();
    } catch (error) {
      console.error("Failed to delete:", error);
    }
  };

  if (loading) {
    return <div>加载中...</div>;
  }

  if (error) {
    return <div className="text-destructive">错误: {error.message}</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">Hooks 示例</h2>
        <Button onClick={handleCreate} disabled={creating}>
          {creating ? "创建中..." : "创建 Pomodoro"}
        </Button>
      </div>

      <div className="space-y-2">
        {pomodoros.length === 0 ? (
          <p className="text-muted-foreground">暂无数据</p>
        ) : (
          pomodoros.map((pomodoro) => (
            <div
              key={pomodoro.id}
              className="flex items-center justify-between rounded-lg border p-4"
            >
              <div>
                <p className="font-medium">
                  {pomodoro.duration} 分钟
                  {pomodoro.is_completed && " ✓"}
                </p>
                <p className="text-sm text-muted-foreground">
                  {new Date(pomodoro.start_time).toLocaleString()}
                </p>
              </div>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleToggle(pomodoro.id, pomodoro.is_completed)}
                  disabled={updating}
                >
                  {pomodoro.is_completed ? "标记未完成" : "标记完成"}
                </Button>
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => handleDelete(pomodoro.id)}
                  disabled={deleting}
                >
                  删除
                </Button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

