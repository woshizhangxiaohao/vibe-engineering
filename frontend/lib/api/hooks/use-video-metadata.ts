import { useState, useEffect, useRef } from 'react';
import { videoService } from '../services/video.service';
import type { VideoMetadata } from '@/types/video';

const CACHE_KEY_PREFIX = 'video_metadata_';
const ANALYTICS_KEY = 'video_analytics';

interface VideoAnalytics {
  successCount: number;
  errorCount: number;
  cancelCount: number;
  errors: Record<string, number>;
}

function getAnalytics(): VideoAnalytics {
  const data = sessionStorage.getItem(ANALYTICS_KEY);
  if (!data) {
    return {
      successCount: 0,
      errorCount: 0,
      cancelCount: 0,
      errors: {},
    };
  }
  return JSON.parse(data);
}

function updateAnalytics(type: 'success' | 'error' | 'cancel', errorType?: string) {
  const analytics = getAnalytics();
  
  if (type === 'success') {
    analytics.successCount++;
  } else if (type === 'error' && errorType) {
    analytics.errorCount++;
    analytics.errors[errorType] = (analytics.errors[errorType] || 0) + 1;
  } else if (type === 'cancel') {
    analytics.cancelCount++;
  }
  
  sessionStorage.setItem(ANALYTICS_KEY, JSON.stringify(analytics));
}

function getCachedMetadata(videoId: string): VideoMetadata | null {
  const cached = sessionStorage.getItem(CACHE_KEY_PREFIX + videoId);
  if (!cached) return null;
  
  try {
    return JSON.parse(cached);
  } catch {
    return null;
  }
}

function setCachedMetadata(videoId: string, metadata: VideoMetadata) {
  sessionStorage.setItem(CACHE_KEY_PREFIX + videoId, JSON.stringify(metadata));
}

export function useVideoMetadata(url: string, videoId: string | null) {
  const [metadata, setMetadata] = useState<VideoMetadata | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  useEffect(() => {
    if (!url || !videoId) {
      setMetadata(null);
      setError(null);
      return;
    }

    const cached = getCachedMetadata(videoId);
    if (cached) {
      setMetadata(cached);
      setError(null);
      return;
    }

    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      updateAnalytics('cancel');
    }

    const abortController = new AbortController();
    abortControllerRef.current = abortController;

    const fetchMetadata = async () => {
      setLoading(true);
      setError(null);

      try {
        const data = await videoService.getVideoMetadata({
          url,
          signal: abortController.signal,
        });
        
        if (!abortController.signal.aborted) {
          setMetadata(data);
          setCachedMetadata(videoId, data);
          updateAnalytics('success');
        }
      } catch (err: any) {
        if (!abortController.signal.aborted && err.message !== 'REQUEST_CANCELED') {
          setError(err.message);
          setMetadata(null);
          updateAnalytics('error', err.message);
        }
      } finally {
        if (!abortController.signal.aborted) {
          setLoading(false);
        }
      }
    };

    fetchMetadata();

    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [url, videoId]);

  const retry = () => {
    if (videoId) {
      sessionStorage.removeItem(CACHE_KEY_PREFIX + videoId);
      setMetadata(null);
      setError(null);
    }
  };

  return {
    metadata,
    loading,
    error,
    retry,
  };
}