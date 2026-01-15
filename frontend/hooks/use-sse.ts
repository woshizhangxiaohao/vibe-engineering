/**
 * SSE (Server-Sent Events) Hook
 * 用于处理流式响应
 */

import { useCallback, useRef, useState } from "react";
import { API_BASE_URL } from "@/lib/api/config";
import { getAuthToken } from "@/lib/utils/storage";

interface SSEOptions<T> {
  onMessage?: (data: T) => void;
  onError?: (error: Error) => void;
  onComplete?: () => void;
}

interface SSEState {
  isConnecting: boolean;
  isStreaming: boolean;
  error: Error | null;
}

/**
 * SSE Hook - 处理 Server-Sent Events 流式响应
 */
export function useSSE<T = unknown>() {
  const [state, setState] = useState<SSEState>({
    isConnecting: false,
    isStreaming: false,
    error: null,
  });
  const abortControllerRef = useRef<AbortController | null>(null);

  /**
   * 开始 SSE 流式请求
   */
  const startStream = useCallback(
    async (
      endpoint: string,
      body: unknown,
      options: SSEOptions<T> = {}
    ): Promise<void> => {
      const { onMessage, onError, onComplete } = options;

      // 取消之前的请求
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }

      const controller = new AbortController();
      abortControllerRef.current = controller;

      setState({ isConnecting: true, isStreaming: false, error: null });

      try {
        const url = `${API_BASE_URL}/api${endpoint}`;
        const headers: Record<string, string> = {
          "Content-Type": "application/json",
          Accept: "text/event-stream",
        };

        const token = getAuthToken();
        if (token) {
          headers["Authorization"] = `Bearer ${token}`;
        }

        const response = await fetch(url, {
          method: "POST",
          headers,
          body: JSON.stringify(body),
          signal: controller.signal,
        });

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}));
          throw new Error(errorData.error || `HTTP ${response.status}`);
        }

        setState({ isConnecting: false, isStreaming: true, error: null });

        const reader = response.body?.getReader();
        if (!reader) {
          throw new Error("No response body");
        }

        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          buffer = lines.pop() || "";

          for (const line of lines) {
            if (line.startsWith("data:")) {
              const dataStr = line.slice(5).trim();
              if (dataStr) {
                try {
                  const data = JSON.parse(dataStr) as T;
                  onMessage?.(data);
                } catch {
                  // 忽略解析错误
                }
              }
            }
          }
        }

        setState({ isConnecting: false, isStreaming: false, error: null });
        onComplete?.();
      } catch (error) {
        if (error instanceof Error && error.name === "AbortError") {
          setState({ isConnecting: false, isStreaming: false, error: null });
          return;
        }

        const err = error instanceof Error ? error : new Error(String(error));
        setState({ isConnecting: false, isStreaming: false, error: err });
        onError?.(err);
      }
    },
    []
  );

  /**
   * 停止 SSE 流
   */
  const stopStream = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setState({ isConnecting: false, isStreaming: false, error: null });
  }, []);

  return {
    ...state,
    startStream,
    stopStream,
  };
}
