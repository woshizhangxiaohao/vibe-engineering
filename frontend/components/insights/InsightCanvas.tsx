"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { insightApi } from "@/lib/api/endpoints";
import type { InsightDetailResponse } from "@/lib/api/types";
import { VideoPreview } from "./VideoPreview";
import { SummarySection } from "./SummarySection";
import { TranscriptView } from "./TranscriptView";
import { Loader2, AlertCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";

interface InsightCanvasProps {
  insightId: number;
}

type DisplayMode = "zh" | "en" | "bilingual";

/**
 * InsightCanvas - Main content display component for Insight details
 * Shows video preview, AI summary, key points, and transcript
 */
export function InsightCanvas({ insightId }: InsightCanvasProps) {
  const [insight, setInsight] = useState<InsightDetailResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [displayMode, setDisplayMode] = useState<DisplayMode>("zh");
  const [isReprocessing, setIsReprocessing] = useState(false);
  const playerRef = useRef<{ seekTo: (seconds: number) => void } | null>(null);
  const pollingRef = useRef<NodeJS.Timeout | null>(null);

  const fetchInsight = useCallback(async () => {
    try {
      setError(null);
      const response = await insightApi.getInsightDetail(insightId);
      setInsight(response);

      // Start polling if still processing
      if (
        response.status === "processing" ||
        response.status === "pending"
      ) {
        if (!pollingRef.current) {
          pollingRef.current = setInterval(() => {
            fetchInsight();
          }, 3000);
        }
      } else {
        // Stop polling when completed or failed
        if (pollingRef.current) {
          clearInterval(pollingRef.current);
          pollingRef.current = null;
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load insight");
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
    } finally {
      setLoading(false);
    }
  }, [insightId]);

  useEffect(() => {
    setLoading(true);
    fetchInsight();

    return () => {
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
    };
  }, [fetchInsight]);

  const handleTimestampClick = useCallback((seconds: number) => {
    playerRef.current?.seekTo(seconds);
  }, []);

  const handleReprocess = async () => {
    if (!insight) return;
    setIsReprocessing(true);
    try {
      await insightApi.reprocessInsight(insight.id);
      fetchInsight();
    } catch (err) {
      setError(err instanceof Error ? err.message : "重新处理失败");
    } finally {
      setIsReprocessing(false);
    }
  };

  // Loading state
  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center space-y-4">
          <Loader2 className="h-8 w-8 animate-spin mx-auto text-muted-foreground" />
          <p className="text-sm text-muted-foreground">加载中...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center space-y-4 max-w-md px-6">
          <AlertCircle className="h-12 w-12 mx-auto text-red-500" />
          <h2 className="text-lg font-semibold">加载失败</h2>
          <p className="text-sm text-muted-foreground">{error}</p>
          <Button variant="outline" onClick={() => fetchInsight()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            重试
          </Button>
        </div>
      </div>
    );
  }

  if (!insight) {
    return null;
  }

  // Processing state
  if (insight.status === "processing" || insight.status === "pending") {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center space-y-4 max-w-md px-6">
          <Loader2 className="h-12 w-12 animate-spin mx-auto text-primary" />
          <h2 className="text-lg font-semibold">
            {insight.status === "pending" ? "等待处理..." : "处理中..."}
          </h2>
          <p className="text-sm text-muted-foreground">
            正在分析视频内容，这可能需要几分钟时间
          </p>
          {insight.title && (
            <p className="text-sm font-medium mt-4">{insight.title}</p>
          )}
        </div>
      </div>
    );
  }

  // Failed state
  if (insight.status === "failed") {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center space-y-4 max-w-md px-6">
          <AlertCircle className="h-12 w-12 mx-auto text-red-500" />
          <h2 className="text-lg font-semibold">处理失败</h2>
          <p className="text-sm text-muted-foreground">
            {insight.error_message || "处理过程中发生错误"}
          </p>
          <Button
            variant="outline"
            onClick={handleReprocess}
            disabled={isReprocessing}
          >
            {isReprocessing ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <RefreshCw className="h-4 w-4 mr-2" />
            )}
            重新处理
          </Button>
        </div>
      </div>
    );
  }

  // Completed state - show full content
  return (
    <ScrollArea className="flex-1">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        {/* Video Preview */}
        {insight.source_type === "youtube" && insight.source_id && (
          <VideoPreview
            ref={playerRef}
            videoId={insight.source_id}
            title={insight.title}
            author={insight.author}
            duration={insight.duration}
            thumbnailUrl={insight.thumbnail_url}
          />
        )}

        {/* Summary and Key Points */}
        <SummarySection
          summary={insight.summary}
          keyPoints={insight.key_points}
        />

        {/* Transcript */}
        {insight.transcripts && insight.transcripts.length > 0 && (
          <TranscriptView
            transcripts={insight.transcripts}
            displayMode={displayMode}
            onDisplayModeChange={setDisplayMode}
            onTimestampClick={handleTimestampClick}
          />
        )}
      </div>
    </ScrollArea>
  );
}
