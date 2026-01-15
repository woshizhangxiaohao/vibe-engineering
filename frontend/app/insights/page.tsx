"use client";

import { useState } from "react";
import { MemoryRail } from "@/components/insights/MemoryRail";
import { InsightCanvas } from "@/components/insights/InsightCanvas";
import { cn } from "@/lib/utils";

/**
 * Insights Page
 * Main page for InsightFlow - AI 灵感解析工作台
 * Features a three-column layout with Memory Rail on the left
 */
export default function InsightsPage() {
  const [selectedInsightId, setSelectedInsightId] = useState<number>();

  return (
    <div className="h-screen flex bg-background">
      {/* Left Sidebar - Memory Rail */}
      <aside
        className={cn(
          "w-80 border-r border-border/50 flex-shrink-0",
          "hidden md:block"
        )}
      >
        <MemoryRail
          onSelectInsight={setSelectedInsightId}
          selectedId={selectedInsightId}
        />
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col min-w-0">
        {selectedInsightId ? (
          <InsightCanvas insightId={selectedInsightId} />
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center space-y-4 max-w-md px-6">
              <h1 className="text-4xl font-bold tracking-tight">
                InsightFlow
              </h1>
              <p className="text-lg text-muted-foreground">
                AI 灵感解析工作台
              </p>
              <div className="mt-8 p-6 bg-muted/50 rounded-2xl">
                <p className="text-sm text-muted-foreground">
                  从左侧选择一条记录，或点击「新建解析」开始
                </p>
              </div>
            </div>
          </div>
        )}
      </main>

      {/* Right Sidebar - Reserved for future features */}
      <aside
        className={cn(
          "w-80 border-l border-border/50 flex-shrink-0",
          "hidden lg:block bg-muted/20"
        )}
      >
        <div className="h-full flex items-center justify-center p-6">
          <p className="text-sm text-muted-foreground text-center">
            侧边栏功能即将上线
          </p>
        </div>
      </aside>
    </div>
  );
}
