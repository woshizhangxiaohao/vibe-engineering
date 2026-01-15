/**
 * useSSE Hook - SSE 流式响应处理
 */

"use client";

import { useState, useCallback, useRef } from "react";
import { SSEChunk } from "../types";

export interface UseSSEOptions {
  onMessage?: (chunk: SSEChunk) => void;
  onComplete?: (messageId?: number) => void;
  onError?: (error: Error) => void;
}

export interface UseSSEReturn {
  isStreaming: boolean;
  error: Error | null;
  connect: (eventSource: EventSource) => void;
  disconnect: () => void;
}

/**
 * SSE 流式响应处理 Hook
 */
export function useSSE(options: UseSSEOptions = {}): UseSSEReturn {
  const { onMessage, onComplete, onError } = options;

  const [isStreaming, setIsStreaming] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
      setIsStreaming(false);
    }
  }, []);

  const connect = useCallback(
    (eventSource: EventSource) => {
      // 关闭之前的连接
      disconnect();

      // 保存新连接
      eventSourceRef.current = eventSource;
      setIsStreaming(true);
      setError(null);

      // 监听消息
      eventSource.onmessage = (event) => {
        try {
          const chunk: SSEChunk = JSON.parse(event.data);

          // 调用回调
          onMessage?.(chunk);

          // 如果完成，关闭连接
          if (chunk.done) {
            disconnect();
            onComplete?.(chunk.message_id);
          }
        } catch (err) {
          const parseError = new Error(
            `Failed to parse SSE message: ${err instanceof Error ? err.message : String(err)}`
          );
          setError(parseError);
          onError?.(parseError);
          disconnect();
        }
      };

      // 监听错误
      eventSource.onerror = (event) => {
        const sseError = new Error("SSE connection error");
        setError(sseError);
        onError?.(sseError);
        disconnect();
      };
    },
    [disconnect, onMessage, onComplete, onError]
  );

  return {
    isStreaming,
    error,
    connect,
    disconnect,
  };
}
