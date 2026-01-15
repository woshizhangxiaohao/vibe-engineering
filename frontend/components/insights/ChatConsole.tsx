/**
 * ChatConsole - 右侧对话面板主容器
 *
 * 遵循 Base.org 设计规范：
 * - 无边框设计
 * - 极简色彩
 * - 背景色差异创建层次
 */

"use client";

import { useEffect, useRef } from "react";
import { Card } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Bot, AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { useChat } from "@/lib/api/hooks/use-chat";
import { ChatMessage } from "./ChatMessage";
import { ChatInput } from "./ChatInput";
import { InsightCard } from "./InsightCard";

export interface ChatConsoleProps {
  insightId: number;
  prefilledMessage?: string;
  className?: string;
}

export function ChatConsole({
  insightId,
  prefilledMessage,
  className,
}: ChatConsoleProps) {
  const {
    messages,
    streamingContent,
    isStreaming,
    isLoading,
    error,
    sendMessage,
  } = useChat({ insightId });

  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, streamingContent]);

  const handleSendMessage = (message: string) => {
    sendMessage(message);
  };

  const handleSuggestionClick = (prompt: string) => {
    sendMessage(prompt);
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-background rounded-2xl",
        className
      )}
    >
      {/* 标题 */}
      <div className="flex items-center gap-2 p-4 border-b border-border/50">
        <Bot className="w-5 h-5 text-primary" />
        <h2 className="text-lg font-semibold">AI 助手</h2>
      </div>

      {/* 内容区域 */}
      <div className="flex-1 flex flex-col min-h-0">
        <ScrollArea className="flex-1 p-4" ref={scrollAreaRef}>
          <div className="space-y-4">
            {/* 智能洞察卡片 */}
            <InsightCard
              insightId={insightId}
              onSuggestionClick={handleSuggestionClick}
            />

            {/* 对话历史 */}
            {isLoading && messages.length === 0 ? (
              <div className="flex items-center justify-center py-8">
                <p className="text-sm text-muted-foreground">加载对话历史中...</p>
              </div>
            ) : messages.length === 0 && !isStreaming ? (
              <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
                <Bot className="w-12 h-12 text-muted-foreground/50 mb-4" />
                <p className="text-base text-foreground font-medium mb-2">
                  开始对话
                </p>
                <p className="text-sm text-muted-foreground max-w-sm">
                  针对当前笔记内容提问，AI 会基于内容为你提供智能洞察
                </p>
              </div>
            ) : (
              <>
                {messages.map((message) => (
                  <ChatMessage key={message.id} message={message} />
                ))}

                {/* 流式消息 */}
                {isStreaming && streamingContent && (
                  <ChatMessage
                    message={{
                      id: -1,
                      role: "assistant",
                      content: streamingContent,
                      created_at: new Date().toISOString(),
                    }}
                    isStreaming={true}
                  />
                )}
              </>
            )}

            {/* 错误提示 */}
            {error && (
              <Alert className="bg-destructive/10 border-0">
                <AlertCircle className="h-4 w-4 text-destructive" />
                <AlertDescription className="text-destructive">
                  {error.message}
                </AlertDescription>
              </Alert>
            )}

            {/* 滚动锚点 */}
            <div ref={messagesEndRef} />
          </div>
        </ScrollArea>

        {/* 输入框 */}
        <div className="p-4 border-t border-border/50">
          <ChatInput
            onSend={handleSendMessage}
            disabled={isStreaming}
            prefilledMessage={prefilledMessage}
          />
        </div>
      </div>
    </div>
  );
}
