/**
 * useChat Hook - 对话状态管理
 */

"use client";

import { useState, useCallback, useEffect } from "react";
import { chatService } from "../services/chat.service";
import { ChatMessage, SendChatMessageRequest, SSEChunk } from "../types";
import { useSSE } from "./use-sse";
import { ApiError } from "../types";

export interface UseChatOptions {
  insightId: number;
  onMessageComplete?: (messageId: number) => void;
}

export interface UseChatReturn {
  messages: ChatMessage[];
  streamingContent: string;
  isStreaming: boolean;
  isLoading: boolean;
  error: Error | null;
  sendMessage: (message: string, highlightId?: number) => Promise<void>;
  refetch: () => Promise<void>;
}

/**
 * Chat 状态管理 Hook
 */
export function useChat(options: UseChatOptions): UseChatReturn {
  const { insightId, onMessageComplete } = options;

  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [streamingContent, setStreamingContent] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  // SSE 流式响应处理
  const { isStreaming, connect, disconnect } = useSSE({
    onMessage: (chunk: SSEChunk) => {
      if (!chunk.done) {
        // 累积流式内容
        setStreamingContent((prev) => prev + chunk.content);
      }
    },
    onComplete: (messageId) => {
      // 流式完成，刷新消息列表
      refetch();
      setStreamingContent("");
      if (messageId) {
        onMessageComplete?.(messageId);
      }
    },
    onError: (err) => {
      setError(err);
      setStreamingContent("");
    },
  });

  // 获取对话历史
  const refetch = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await chatService.getChatHistory(insightId);
      setMessages(response.messages);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err);
      } else {
        setError(new Error(err instanceof Error ? err.message : String(err)));
      }
    } finally {
      setIsLoading(false);
    }
  }, [insightId]);

  // 发送消息
  const sendMessage = useCallback(
    async (message: string, highlightId?: number) => {
      try {
        setError(null);
        setStreamingContent("");

        // 立即添加用户消息到列表（乐观更新）
        const userMessage: ChatMessage = {
          id: Date.now(), // 临时 ID
          role: "user",
          content: message,
          created_at: new Date().toISOString(),
        };
        setMessages((prev) => [...prev, userMessage]);

        // 创建 SSE 连接
        const eventSource = chatService.createChatStream(insightId, {
          message,
          highlight_id: highlightId,
        });

        // 连接 SSE
        connect(eventSource);
      } catch (err) {
        if (err instanceof ApiError) {
          setError(err);
        } else {
          setError(new Error(err instanceof Error ? err.message : String(err)));
        }
        // 移除乐观添加的消息
        setMessages((prev) => prev.slice(0, -1));
      }
    },
    [insightId, connect]
  );

  // 初始加载
  useEffect(() => {
    refetch();
  }, [refetch]);

  // 清理
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    messages,
    streamingContent,
    isStreaming,
    isLoading,
    error,
    sendMessage,
    refetch,
  };
}
