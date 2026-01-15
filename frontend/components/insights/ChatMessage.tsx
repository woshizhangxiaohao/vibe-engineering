/**
 * ChatMessage - 单条消息组件
 *
 * 遵循 Base.org 设计规范：
 * - 无边框设计
 * - 极简色彩
 * - 背景色差异创建层次
 */

"use client";

import { ChatMessage as ChatMessageType } from "@/lib/api/types";
import { cn } from "@/lib/utils";
import ReactMarkdown from "react-markdown";
import { User, Bot } from "lucide-react";

export interface ChatMessageProps {
  message: ChatMessageType;
  isStreaming?: boolean;
  className?: string;
}

export function ChatMessage({
  message,
  isStreaming = false,
  className,
}: ChatMessageProps) {
  const isUser = message.role === "user";

  return (
    <div
      className={cn(
        "flex gap-3 p-4 rounded-xl transition-colors duration-200",
        isUser ? "bg-muted/50" : "bg-background",
        className
      )}
    >
      {/* 头像 */}
      <div
        className={cn(
          "flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center",
          isUser ? "bg-primary text-primary-foreground" : "bg-muted"
        )}
      >
        {isUser ? (
          <User className="w-4 h-4" />
        ) : (
          <Bot className="w-4 h-4 text-foreground" />
        )}
      </div>

      {/* 内容 */}
      <div className="flex-1 min-w-0">
        {/* 角色标签 */}
        <div className="text-sm font-medium text-foreground mb-1">
          {isUser ? "用户" : "AI 助手"}
        </div>

        {/* 消息内容 */}
        <div
          className={cn(
            "text-base leading-relaxed",
            isUser ? "text-foreground" : "text-foreground/90"
          )}
        >
          {isUser ? (
            // 用户消息：纯文本
            <p className="whitespace-pre-wrap break-words">{message.content}</p>
          ) : (
            // AI 消息：支持 Markdown
            <div className="prose prose-sm max-w-none">
              <ReactMarkdown
                components={{
                  // 自定义样式以符合 Base.org 设计
                  p: ({ children }) => (
                    <p className="mb-2 last:mb-0">{children}</p>
                  ),
                  ul: ({ children }) => (
                    <ul className="mb-2 ml-4 list-disc">{children}</ul>
                  ),
                  ol: ({ children }) => (
                    <ol className="mb-2 ml-4 list-decimal">{children}</ol>
                  ),
                  li: ({ children }) => <li className="mb-1">{children}</li>,
                  code: ({ children, className }) => {
                    const isInline = !className;
                    return isInline ? (
                      <code className="px-1.5 py-0.5 rounded-md bg-muted font-mono text-sm">
                        {children}
                      </code>
                    ) : (
                      <code className="block p-3 rounded-lg bg-muted font-mono text-sm overflow-x-auto">
                        {children}
                      </code>
                    );
                  },
                  strong: ({ children }) => (
                    <strong className="font-semibold">{children}</strong>
                  ),
                  em: ({ children }) => (
                    <em className="italic">{children}</em>
                  ),
                }}
              >
                {message.content}
              </ReactMarkdown>
            </div>
          )}

          {/* 流式加载指示器 */}
          {isStreaming && !isUser && (
            <span className="inline-block w-1 h-4 ml-1 bg-primary animate-pulse" />
          )}
        </div>

        {/* 时间戳 */}
        <div className="mt-2 text-xs text-muted-foreground">
          {new Date(message.created_at).toLocaleTimeString("zh-CN", {
            hour: "2-digit",
            minute: "2-digit",
          })}
        </div>
      </div>
    </div>
  );
}
