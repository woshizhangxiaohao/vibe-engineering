"use client";

import React, { useEffect, useState, useRef, useCallback } from 'react';
import { Progress } from "@/components/ui/progress";
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card";
import { youtubeApi } from "@/lib/api/endpoints";
import { QuotaStatus } from "@/types/video";
import { Info } from "lucide-react";

// 简单的请求缓存，避免重复调用
const quotaCache = {
  data: null as QuotaStatus | null,
  timestamp: 0,
  CACHE_TTL: 30000, // 30秒缓存
  pendingRequest: null as Promise<QuotaStatus> | null,
};

export default function QuotaMonitor() {
  const [quota, setQuota] = useState<QuotaStatus | null>(quotaCache.data);
  const isMounted = useRef(true);

  const fetchQuota = useCallback(async () => {
    const now = Date.now();
    
    // 如果缓存有效，直接返回缓存数据
    if (quotaCache.data && now - quotaCache.timestamp < quotaCache.CACHE_TTL) {
      if (isMounted.current) {
        setQuota(quotaCache.data);
      }
      return;
    }

    // 如果有正在进行的请求，等待它完成（请求去重）
    if (quotaCache.pendingRequest) {
      try {
        const data = await quotaCache.pendingRequest;
        if (isMounted.current) {
          setQuota(data);
        }
      } catch (e) {
        // 静默处理等待中的请求错误，因为主请求会处理错误
        console.warn("Pending quota request failed:", e);
      }
      return;
    }

    // 发起新请求
    try {
      quotaCache.pendingRequest = youtubeApi.getQuota();
      const data = await quotaCache.pendingRequest;
      
      // 更新缓存
      quotaCache.data = data;
      quotaCache.timestamp = Date.now();
      
      if (isMounted.current) {
        setQuota(data);
      }
    } catch (e) {
      // 检查是否是认证错误
      const error = e as Error & { status?: number };
      if (error.status === 401 || error.message?.includes("No refresh token")) {
        // 认证失败，清除缓存，组件将不显示
        quotaCache.data = null;
        quotaCache.timestamp = 0;
        if (isMounted.current) {
          setQuota(null);
        }
        console.warn("Quota fetch failed: Authentication required");
      } else {
        console.error("Failed to fetch quota:", e);
      }
    } finally {
      quotaCache.pendingRequest = null;
    }
  }, []);

  useEffect(() => {
    isMounted.current = true;
    
    fetchQuota();
    const interval = setInterval(fetchQuota, 60000);
    
    return () => {
      isMounted.current = false;
      clearInterval(interval);
    };
  }, [fetchQuota]);

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