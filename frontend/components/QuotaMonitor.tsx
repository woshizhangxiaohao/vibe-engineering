"use client";

import React, { useEffect, useState } from 'react';
import { Progress } from "@/components/ui/progress";
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card";
import { youtubeApi } from "@/lib/api/endpoints";
import { QuotaStatus } from "@/types/video";
import { Info } from "lucide-react";

export default function QuotaMonitor() {
  const [quota, setQuota] = useState<QuotaStatus | null>(null);

  useEffect(() => {
    const fetchQuota = async () => {
      try {
        const data = await youtubeApi.getQuota();
        setQuota(data);
      } catch (e) {
        console.error("Failed to fetch quota");
      }
    };
    fetchQuota();
    const interval = setInterval(fetchQuota, 60000);
    return () => clearInterval(interval);
  }, []);

  if (!quota) return null;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">API Quota</span>
        <HoverCard>
          <HoverCardTrigger asChild>
            <button className="text-muted-foreground hover:text-primary transition-colors">
              <Info className="h-3.5 w-3.5" />
            </button>
          </HoverCardTrigger>
          <HoverCardContent side="top" className="w-64 rounded-xl border-0 bg-card p-4">
            <div className="space-y-2">
              <p className="text-sm font-semibold">Google API Usage</p>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Remaining</span>
                <span className="font-mono">{quota.remaining.toLocaleString()}</span>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Daily Limit</span>
                <span className="font-mono">{quota.total.toLocaleString()}</span>
              </div>
            </div>
          </HoverCardContent>
        </HoverCard>
      </div>
      <Progress value={quota.percent} className="h-1.5 bg-secondary" />
      <p className="text-[10px] text-muted-foreground">{quota.percent}% of daily units used</p>
    </div>
  );
}