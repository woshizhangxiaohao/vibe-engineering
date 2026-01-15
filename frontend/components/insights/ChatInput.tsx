/**
 * ChatInput - 输入框组件
 *
 * 遵循 Base.org 设计规范：
 * - 无边框设计
 * - 背景色变化创建焦点状态
 * - 极简按钮样式
 */

"use client";

import { useState, KeyboardEvent } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Send } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ChatInputProps {
  onSend: (message: string) => void;
  disabled?: boolean;
  placeholder?: string;
  prefilledMessage?: string;
  className?: string;
}

export function ChatInput({
  onSend,
  disabled = false,
  placeholder = "💬 输入你的问题...",
  prefilledMessage,
  className,
}: ChatInputProps) {
  const [message, setMessage] = useState(prefilledMessage || "");

  const handleSend = () => {
    const trimmedMessage = message.trim();
    if (!trimmedMessage || disabled) return;

    onSend(trimmedMessage);
    setMessage("");
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    // Cmd/Ctrl + Enter 发送
    if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className={cn("flex gap-2", className)}>
      {/* 输入框 - 无边框设计 */}
      <Textarea
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        disabled={disabled}
        rows={1}
        className={cn(
          "flex-1 min-h-[48px] max-h-[120px] resize-none",
          "rounded-xl border-0",
          "bg-muted px-4 py-3",
          "text-base placeholder:text-muted-foreground",
          "focus:bg-background focus:outline-none",
          "transition-all duration-200"
        )}
      />

      {/* 发送按钮 - 无边框设计 */}
      <Button
        onClick={handleSend}
        disabled={disabled || !message.trim()}
        size="icon"
        className={cn(
          "h-12 w-12 rounded-xl",
          "bg-primary text-primary-foreground",
          "hover:bg-primary/90",
          "disabled:opacity-50 disabled:cursor-not-allowed",
          "transition-all duration-200"
        )}
      >
        <Send className="w-5 h-5" />
        <span className="sr-only">发送</span>
      </Button>
    </div>
  );
}
