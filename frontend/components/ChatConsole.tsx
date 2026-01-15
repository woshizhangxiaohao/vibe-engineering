"use client";

import React, { useCallback, useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card } from "@/components/ui/card";
import { Loader2, Send, StopCircle, User, Bot, Sparkles } from "lucide-react";
import { useChat } from "@/hooks/use-chat";
import { ChatMessage } from "@/types/chat";
import { cn } from "@/lib/utils";

interface ChatConsoleProps {
  analysisId: number;
  className?: string;
}

/**
 * ChatMessage 组件 - 显示单条消息
 */
function MessageBubble({ message }: { message: ChatMessage }) {
  const isUser = message.role === "user";

  return (
    <div
      className={cn(
        "flex gap-3 mb-4",
        isUser ? "flex-row-reverse" : "flex-row"
      )}
    >
      {/* Avatar */}
      <div
        className={cn(
          "h-8 w-8 rounded-lg flex items-center justify-center shrink-0",
          isUser ? "bg-primary text-primary-foreground" : "bg-muted"
        )}
      >
        {isUser ? (
          <User className="h-4 w-4" />
        ) : (
          <Bot className="h-4 w-4" />
        )}
      </div>

      {/* Content */}
      <div
        className={cn(
          "max-w-[80%] rounded-xl px-4 py-3",
          isUser
            ? "bg-primary text-primary-foreground"
            : "bg-muted text-foreground"
        )}
      >
        <p className="text-sm whitespace-pre-wrap leading-relaxed">
          {message.content}
        </p>
      </div>
    </div>
  );
}

/**
 * 输入框组件
 */
function ChatInput({
  onSend,
  disabled,
  isStreaming,
  onStop,
}: {
  onSend: (message: string) => void;
  disabled: boolean;
  isStreaming: boolean;
  onStop: () => void;
}) {
  const [input, setInput] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSubmit = useCallback(() => {
    if (input.trim() && !disabled) {
      onSend(input.trim());
      setInput("");
    }
  }, [input, disabled, onSend]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  // 自动调整高度
  useEffect(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = "auto";
      textarea.style.height = Math.min(textarea.scrollHeight, 120) + "px";
    }
  }, [input]);

  return (
    <div className="border-t bg-background p-4">
      <div className="flex gap-2 items-end">
        <textarea
          ref={textareaRef}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Ask about the video..."
          disabled={disabled || isStreaming}
          rows={1}
          className={cn(
            "flex-1 resize-none rounded-xl border border-input bg-background px-4 py-3",
            "text-sm placeholder:text-muted-foreground",
            "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
            "disabled:cursor-not-allowed disabled:opacity-50",
            "min-h-[44px] max-h-[120px]"
          )}
        />
        {isStreaming ? (
          <Button
            onClick={onStop}
            size="icon"
            variant="destructive"
            className="h-11 w-11 rounded-xl shrink-0"
          >
            <StopCircle className="h-5 w-5" />
          </Button>
        ) : (
          <Button
            onClick={handleSubmit}
            disabled={!input.trim() || disabled}
            size="icon"
            className="h-11 w-11 rounded-xl shrink-0"
          >
            <Send className="h-5 w-5" />
          </Button>
        )}
      </div>
    </div>
  );
}

/**
 * ChatConsole 组件 - AI 对话面板
 */
export default function ChatConsole({ analysisId, className }: ChatConsoleProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const { messages, isLoading, isStreaming, sendMessage, stopStreaming } = useChat({
    analysisId,
    onError: (error) => {
      console.error("Chat error:", error);
    },
  });

  // 自动滚动到底部
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  return (
    <Card className={cn("flex flex-col h-full border-0 overflow-hidden", className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b shrink-0">
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-primary" />
          <h3 className="font-semibold text-sm">AI Assistant</h3>
        </div>
      </div>

      {/* Messages Area */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto p-4">
        {isLoading ? (
          <div className="flex items-center justify-center h-full">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center px-4">
            <div className="h-12 w-12 rounded-xl bg-primary/10 flex items-center justify-center mb-4">
              <Sparkles className="h-6 w-6 text-primary" />
            </div>
            <h4 className="font-semibold mb-2">Start a conversation</h4>
            <p className="text-sm text-muted-foreground max-w-xs">
              Ask questions about the video content, request summaries, or explore specific topics.
            </p>
          </div>
        ) : (
          messages.map((msg) => (
            <MessageBubble key={msg.id} message={msg} />
          ))
        )}

        {/* Streaming indicator */}
        {isStreaming && messages[messages.length - 1]?.role !== "assistant" && (
          <div className="flex gap-3 mb-4">
            <div className="h-8 w-8 rounded-lg bg-muted flex items-center justify-center">
              <Bot className="h-4 w-4" />
            </div>
            <div className="bg-muted rounded-xl px-4 py-3">
              <Loader2 className="h-4 w-4 animate-spin" />
            </div>
          </div>
        )}
      </div>

      {/* Input Area */}
      <ChatInput
        onSend={sendMessage}
        disabled={isLoading}
        isStreaming={isStreaming}
        onStop={stopStreaming}
      />
    </Card>
  );
}
