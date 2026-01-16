/**
 * InsightCard - 智能洞察卡片
 *
 * 遵循 Base.org 设计规范：
 * - 无边框设计
 * - 背景色差异创建层次
 * - 极简图标和排版
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Sparkles, ChevronRight, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { chatService } from "@/lib/api/services/chat.service";
import { AnalyzeEntitiesResponse, ApiError } from "@/lib/api/types";

export interface InsightCardProps {
  insightId: number;
  onSuggestionClick: (prompt: string) => void;
  className?: string;
}

export function InsightCard({
  insightId,
  onSuggestionClick,
  className,
}: InsightCardProps) {
  const [data, setData] = useState<AnalyzeEntitiesResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchEntities = async () => {
      try {
        setIsLoading(true);
        setError(null);
        const response = await chatService.analyzeEntities(insightId);
        setData(response);
      } catch (err) {
        if (err instanceof ApiError) {
          setError(err);
        } else {
          setError(new Error(err instanceof Error ? err.message : String(err)));
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchEntities();
  }, [insightId]);

  if (isLoading) {
    return (
      <Card className={cn("bg-muted/50 rounded-xl", className)}>
        <CardContent className="p-6 flex items-center justify-center">
          <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
          <span className="ml-2 text-sm text-muted-foreground">
            分析中...
          </span>
        </CardContent>
      </Card>
    );
  }

  if (error || !data) {
    return null; // 静默失败，不显示卡片
  }

  // 安全检查：确保 entities 和 suggestions 存在且为数组
  const entities = Array.isArray(data.entities) ? data.entities : [];
  const suggestions = Array.isArray(data.suggestions) ? data.suggestions : [];

  if (entities.length === 0 && suggestions.length === 0) {
    return null; // 没有洞察，不显示卡片
  }

  return (
    <Card
      className={cn(
        "bg-muted/50 rounded-xl transition-colors duration-200",
        className
      )}
    >
      <CardHeader className="p-4 pb-3">
        <CardTitle className="flex items-center gap-2 text-base font-semibold">
          <Sparkles className="w-4 h-4 text-primary" />
          智能洞察
        </CardTitle>
      </CardHeader>

      <CardContent className="p-4 pt-0 space-y-3">
        {/* 检测到的实体 */}
        {entities.length > 0 && (
          <div>
            <div className="text-sm text-muted-foreground mb-2">
              检测到讨论：
              <span className="ml-1 font-medium text-foreground">
                {entities.map((e) => e?.name || "").filter(Boolean).join(", ")}
              </span>
            </div>
          </div>
        )}

        {/* 建议列表 */}
        {suggestions.length > 0 && (
          <div className="space-y-2">
            {suggestions.map((suggestion, index) => (
              <Button
                key={index}
                variant="ghost"
                onClick={() => onSuggestionClick(suggestion.prompt)}
                className={cn(
                  "w-full justify-start text-left h-auto py-2 px-3",
                  "rounded-lg bg-background hover:bg-muted/80",
                  "transition-colors duration-200"
                )}
              >
                <ChevronRight className="w-4 h-4 mr-2 flex-shrink-0 text-muted-foreground" />
                <span className="text-sm text-foreground">
                  {suggestion.prompt}
                </span>
              </Button>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
