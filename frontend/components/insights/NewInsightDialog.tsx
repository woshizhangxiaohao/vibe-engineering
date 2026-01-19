"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { insightApi } from "@/lib/api/endpoints";
import { toast } from "sonner";

interface NewInsightDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

/**
 * NewInsightDialog Component
 * Dialog for creating a new insight parsing task with automatic translation
 */
export function NewInsightDialog({
  open,
  onOpenChange,
  onSuccess,
}: NewInsightDialogProps) {
  const [sourceUrl, setSourceUrl] = useState("");
  const [targetLang, setTargetLang] = useState("zh");
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!sourceUrl.trim()) {
      toast.error("请输入内容链接");
      return;
    }

    try {
      setIsLoading(true);

      const response = await insightApi.createInsight({
        source_url: sourceUrl,
        target_lang: targetLang,
      });

      toast.success(response.message || "解析任务已创建，正在处理中...");
      setSourceUrl("");
      setTargetLang("zh");
      onOpenChange(false);
      onSuccess?.();
    } catch (error) {
      console.error("Failed to create insight:", error);
      toast.error("创建失败，请重试");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md rounded-2xl">
        <DialogHeader>
          <DialogTitle className="text-2xl font-bold">新建解析</DialogTitle>
          <DialogDescription className="text-muted-foreground">
            输入 YouTube、Twitter 或播客链接开始 AI 解析（自动生成摘要和翻译字幕）
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 pt-4">
          <div className="space-y-2">
            <Label htmlFor="source-url">内容链接</Label>
            <Input
              id="source-url"
              type="url"
              placeholder="https://youtube.com/watch?v=..."
              value={sourceUrl}
              onChange={(e) => setSourceUrl(e.target.value)}
              className="h-12 rounded-lg bg-muted focus:bg-background transition-colors"
              disabled={isLoading}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="target-lang">目标语言</Label>
            <select
              id="target-lang"
              value={targetLang}
              onChange={(e) => setTargetLang(e.target.value)}
              className="w-full h-12 rounded-lg bg-muted px-4 text-base focus:bg-background focus:outline-none transition-colors"
              disabled={isLoading}
            >
              <option value="zh">中文 (Chinese)</option>
              <option value="en">English (英文)</option>
              <option value="ja">日本語 (Japanese)</option>
              <option value="ko">한국어 (Korean)</option>
            </select>
          </div>

          {/* 功能说明 */}
          <div className="p-3 rounded-lg bg-muted/50 text-sm text-muted-foreground">
            系统将自动：
            <ul className="list-disc list-inside mt-1 space-y-1">
              <li>提取视频内容和字幕</li>
              <li>生成 AI 摘要和关键点</li>
              <li>翻译字幕到目标语言</li>
              <li>提供中英对照显示</li>
            </ul>
          </div>

          <DialogFooter className="pt-4">
            <Button
              type="button"
              variant="secondary"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
              className="rounded-lg"
            >
              取消
            </Button>
            <Button
              type="submit"
              disabled={isLoading}
              className="rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
            >
              {isLoading ? "创建中..." : "开始解析"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
