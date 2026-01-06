"use client";

import { useState, useCallback } from "react";
import { videoService } from "../services/video.service";
import { VideoMetadata } from "@/types/video";
import { toast } from "sonner";

interface UseVideoMetadataReturn {
  metadata: VideoMetadata | null;
  loading: boolean;
  error: string | null;
  fetchMetadata: (url: string) => Promise<void>;
  clearMetadata: () => void;
}

export function useVideoMetadata(): UseVideoMetadataReturn {
  const [metadata, setMetadata] = useState<VideoMetadata | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchMetadata = useCallback(async (url: string) => {
    try {
      setLoading(true);
      setError(null);
      const data = await videoService.parseYouTubeUrl(url);
      setMetadata(data);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "无法获取视频信息";
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  const clearMetadata = useCallback(() => {
    setMetadata(null);
    setError(null);
  }, []);

  return {
    metadata,
    loading,
    error,
    fetchMetadata,
    clearMetadata,
  };
}