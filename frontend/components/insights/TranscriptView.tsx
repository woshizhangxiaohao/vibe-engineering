"use client";

import { useState } from "react";
import { FileText } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { TranscriptItem } from "@/lib/api/types";

type DisplayMode = "zh" | "en" | "bilingual";

interface TranscriptViewProps {
  transcripts: TranscriptItem[];
  displayMode: DisplayMode;
  onDisplayModeChange: (mode: DisplayMode) => void;
  onTimestampClick: (seconds: number) => void;
}

const displayModeLabels: Record<DisplayMode, string> = {
  zh: "中文",
  en: "原文",
  bilingual: "中英对照",
};

/**
 * TranscriptView - Display transcript with timestamp navigation
 */
export function TranscriptView({
  transcripts,
  displayMode,
  onDisplayModeChange,
  onTimestampClick,
}: TranscriptViewProps) {
  if (!transcripts || transcripts.length === 0) {
    return null;
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <FileText className="h-4 w-4 text-primary" />
            内容全文
          </CardTitle>
          {/* Language Toggle */}
          <div className="flex items-center gap-1 bg-muted rounded-lg p-1">
            {(Object.keys(displayModeLabels) as DisplayMode[]).map((mode) => (
              <Button
                key={mode}
                variant="ghost"
                size="sm"
                className={cn(
                  "h-7 px-3 text-xs",
                  displayMode === mode && "bg-background shadow-sm"
                )}
                onClick={() => onDisplayModeChange(mode)}
              >
                {displayModeLabels[mode]}
              </Button>
            ))}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-3 max-h-[500px] overflow-y-auto pr-2">
          {transcripts.map((item, index) => (
            <TranscriptLine
              key={index}
              item={item}
              displayMode={displayMode}
              onTimestampClick={onTimestampClick}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

interface TranscriptLineProps {
  item: TranscriptItem;
  displayMode: DisplayMode;
  onTimestampClick: (seconds: number) => void;
}

function TranscriptLine({
  item,
  displayMode,
  onTimestampClick,
}: TranscriptLineProps) {
  // Determine which text to display based on mode
  const hasTranslation = item.translated_text && item.translated_text.trim() !== "";

  return (
    <div className="flex gap-3 group">
      {/* Timestamp Button */}
      <button
        onClick={() => onTimestampClick(item.seconds)}
        className="flex-shrink-0 text-xs font-mono text-primary hover:text-primary/80 hover:underline transition-colors pt-0.5"
      >
        [{item.timestamp}]
      </button>

      {/* Text Content */}
      <div className="flex-1 text-sm text-muted-foreground leading-relaxed">
        {displayMode === "zh" && (
          <p>{hasTranslation ? item.translated_text : item.text}</p>
        )}
        {displayMode === "en" && <p>{item.text}</p>}
        {displayMode === "bilingual" && (
          <div className="space-y-1.5">
            <p className="text-foreground">{item.text}</p>
            {hasTranslation && (
              <p className="text-sm text-primary/90 font-medium">
                {item.translated_text}
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
