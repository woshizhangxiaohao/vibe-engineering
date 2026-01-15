/**
 * Chat Hook
 * 管理聊天状态和 SSE 流式响应
 */

import { useCallback, useEffect, useState } from "react";
import { useSSE } from "./use-sse";
import { chatApi } from "@/lib/api/endpoints";
import { ChatMessage, ChatStreamEvent } from "@/types/chat";

interface UseChatOptions {
  analysisId: number;
  onError?: (error: Error) => void;
}

interface UseChatReturn {
  messages: ChatMessage[];
  isLoading: boolean;
  isStreaming: boolean;
  error: Error | null;
  sendMessage: (message: string, highlightId?: number) => void;
  stopStreaming: () => void;
}

/**
 * useChat Hook - 管理 AI 对话
 */
export function useChat({ analysisId, onError }: UseChatOptions): UseChatReturn {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [pendingMessage, setPendingMessage] = useState<string>("");
  const { isConnecting, isStreaming, error, startStream, stopStream } = useSSE<ChatStreamEvent>();

  // 加载历史消息
  useEffect(() => {
    const loadHistory = async () => {
      try {
        setIsLoading(true);
        const response = await chatApi.getHistory(analysisId);
        setMessages(response.messages || []);
      } catch (err) {
        const error = err instanceof Error ? err : new Error(String(err));
        onError?.(error);
      } finally {
        setIsLoading(false);
      }
    };

    loadHistory();
  }, [analysisId, onError]);

  // 发送消息
  const sendMessage = useCallback(
    (message: string, highlightId?: number) => {
      if (!message.trim()) return;

      // 添加用户消息
      const userMessage: ChatMessage = {
        id: Date.now(),
        analysis_id: analysisId,
        role: "user",
        content: message,
        highlight_id: highlightId,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, userMessage]);
      setPendingMessage("");

      // 开始 SSE 流
      startStream(
        `/v1/insights/${analysisId}/chat`,
        { message, highlight_id: highlightId },
        {
          onMessage: (event) => {
            if (event.content) {
              setPendingMessage((prev) => prev + event.content);
            }
            if (event.done && event.message_id) {
              // 流结束，添加完整的助手消息
              setMessages((prev) => [
                ...prev,
                {
                  id: event.message_id!,
                  analysis_id: analysisId,
                  role: "assistant",
                  content: pendingMessage + event.content,
                  created_at: new Date().toISOString(),
                  updated_at: new Date().toISOString(),
                },
              ]);
              setPendingMessage("");
            }
          },
          onError: (err) => {
            setPendingMessage("");
            onError?.(err);
          },
          onComplete: () => {
            // 如果有未完成的消息，添加到列表
            setPendingMessage((current) => {
              if (current) {
                setMessages((prev) => [
                  ...prev,
                  {
                    id: Date.now(),
                    analysis_id: analysisId,
                    role: "assistant",
                    content: current,
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                  },
                ]);
              }
              return "";
            });
          },
        }
      );
    },
    [analysisId, startStream, onError, pendingMessage]
  );

  // 组合消息列表（包括正在流式传输的消息）
  const displayMessages = pendingMessage
    ? [
        ...messages,
        {
          id: -1,
          analysis_id: analysisId,
          role: "assistant" as const,
          content: pendingMessage,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ]
    : messages;

  return {
    messages: displayMessages,
    isLoading: isLoading || isConnecting,
    isStreaming,
    error,
    sendMessage,
    stopStreaming: stopStream,
  };
}
