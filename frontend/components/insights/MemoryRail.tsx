"use client";

import { useState, useEffect } from "react";
import { cn } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { InsightItem } from "./InsightItem";
import { NewInsightDialog } from "./NewInsightDialog";
import { insightApi } from "@/lib/api/endpoints";
import type { GroupedInsights } from "@/lib/api/types";
import { Plus, Search, ChevronDown, ChevronRight } from "lucide-react";
import { toast } from "sonner";

interface MemoryRailProps {
  onSelectInsight: (id: number) => void;
  selectedId?: number;
  className?: string;
}

/**
 * MemoryRail Component
 * Left sidebar displaying time-grouped insights
 */
export function MemoryRail({
  onSelectInsight,
  selectedId,
  className,
}: MemoryRailProps) {
  const [insights, setInsights] = useState<GroupedInsights>({
    today: [],
    yesterday: [],
    previous: [],
  });
  const [searchQuery, setSearchQuery] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [showNewDialog, setShowNewDialog] = useState(false);
  const [openSections, setOpenSections] = useState({
    today: true,
    yesterday: false,
    previous: false,
  });

  const fetchInsights = async () => {
    try {
      setIsLoading(true);
      const response = await insightApi.getInsights({
        limit: 50,
        offset: 0,
        search: searchQuery || undefined,
      });
      setInsights(response.data);
    } catch (error) {
      console.error("Failed to fetch insights:", error);
      toast.error("加载失败，请重试");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchInsights();
  }, [searchQuery]);

  const toggleSection = (section: keyof typeof openSections) => {
    setOpenSections((prev) => ({ ...prev, [section]: !prev[section] }));
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-background",
        className
      )}
    >
      {/* Header - New Insight Button */}
      <div className="p-4 border-b border-border/50">
        <Button
          onClick={() => setShowNewDialog(true)}
          className="w-full rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Plus className="w-4 h-4 mr-2" />
          新建解析
        </Button>
      </div>

      {/* Search Bar */}
      <div className="p-4 border-b border-border/50">
        <div className="relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="搜索笔记..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-12 pl-11 rounded-lg bg-muted focus:bg-background transition-colors"
          />
        </div>
      </div>

      {/* Insights List */}
      <ScrollArea className="flex-1">
        <div className="p-2 space-y-1">
          {/* Today Section */}
          <Collapsible
            open={openSections.today}
            onOpenChange={() => toggleSection("today")}
          >
            <CollapsibleTrigger className="flex items-center w-full px-4 py-2 hover:bg-muted/50 rounded-lg transition-colors">
              {openSections.today ? (
                <ChevronDown className="w-4 h-4 mr-2 text-muted-foreground" />
              ) : (
                <ChevronRight className="w-4 h-4 mr-2 text-muted-foreground" />
              )}
              <span className="text-sm font-medium text-foreground">
                今日 (Today)
              </span>
              <span className="ml-auto text-xs text-muted-foreground">
                {insights.today.length}
              </span>
            </CollapsibleTrigger>
            <CollapsibleContent className="pt-1 space-y-1">
              {insights.today.map((insight) => (
                <InsightItem
                  key={insight.id}
                  insight={insight}
                  isSelected={selectedId === insight.id}
                  onSelect={onSelectInsight}
                />
              ))}
              {insights.today.length === 0 && (
                <p className="px-4 py-2 text-xs text-muted-foreground">
                  暂无记录
                </p>
              )}
            </CollapsibleContent>
          </Collapsible>

          {/* Yesterday Section */}
          <Collapsible
            open={openSections.yesterday}
            onOpenChange={() => toggleSection("yesterday")}
          >
            <CollapsibleTrigger className="flex items-center w-full px-4 py-2 hover:bg-muted/50 rounded-lg transition-colors">
              {openSections.yesterday ? (
                <ChevronDown className="w-4 h-4 mr-2 text-muted-foreground" />
              ) : (
                <ChevronRight className="w-4 h-4 mr-2 text-muted-foreground" />
              )}
              <span className="text-sm font-medium text-foreground">
                昨日 (Yesterday)
              </span>
              <span className="ml-auto text-xs text-muted-foreground">
                {insights.yesterday.length}
              </span>
            </CollapsibleTrigger>
            <CollapsibleContent className="pt-1 space-y-1">
              {insights.yesterday.map((insight) => (
                <InsightItem
                  key={insight.id}
                  insight={insight}
                  isSelected={selectedId === insight.id}
                  onSelect={onSelectInsight}
                />
              ))}
              {insights.yesterday.length === 0 && (
                <p className="px-4 py-2 text-xs text-muted-foreground">
                  暂无记录
                </p>
              )}
            </CollapsibleContent>
          </Collapsible>

          {/* Previous Section */}
          <Collapsible
            open={openSections.previous}
            onOpenChange={() => toggleSection("previous")}
          >
            <CollapsibleTrigger className="flex items-center w-full px-4 py-2 hover:bg-muted/50 rounded-lg transition-colors">
              {openSections.previous ? (
                <ChevronDown className="w-4 h-4 mr-2 text-muted-foreground" />
              ) : (
                <ChevronRight className="w-4 h-4 mr-2 text-muted-foreground" />
              )}
              <span className="text-sm font-medium text-foreground">
                更早 (Previous)
              </span>
              <span className="ml-auto text-xs text-muted-foreground">
                {insights.previous.length}
              </span>
            </CollapsibleTrigger>
            <CollapsibleContent className="pt-1 space-y-1">
              {insights.previous.map((insight) => (
                <InsightItem
                  key={insight.id}
                  insight={insight}
                  isSelected={selectedId === insight.id}
                  onSelect={onSelectInsight}
                />
              ))}
              {insights.previous.length === 0 && (
                <p className="px-4 py-2 text-xs text-muted-foreground">
                  暂无记录
                </p>
              )}
            </CollapsibleContent>
          </Collapsible>
        </div>
      </ScrollArea>

      {/* Loading State */}
      {isLoading && (
        <div className="absolute inset-0 bg-background/50 backdrop-blur-sm flex items-center justify-center">
          <div className="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin" />
        </div>
      )}

      {/* New Insight Dialog */}
      <NewInsightDialog
        open={showNewDialog}
        onOpenChange={setShowNewDialog}
        onSuccess={fetchInsights}
      />
    </div>
  );
}
