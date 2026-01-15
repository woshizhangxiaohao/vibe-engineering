"use client";

import { forwardRef, useImperativeHandle, useRef, useState } from "react";
import { Play, User, Clock } from "lucide-react";
import { cn } from "@/lib/utils";

interface VideoPreviewProps {
  videoId: string;
  title: string;
  author?: string;
  duration?: number;
  thumbnailUrl?: string;
}

export interface VideoPreviewHandle {
  seekTo: (seconds: number) => void;
}

/**
 * Format duration in seconds to HH:MM:SS or MM:SS
 */
function formatDuration(seconds: number): string {
  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hrs > 0) {
    return `${hrs}:${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  }
  return `${mins}:${secs.toString().padStart(2, "0")}`;
}

/**
 * VideoPreview - Embedded YouTube player with metadata
 */
export const VideoPreview = forwardRef<VideoPreviewHandle, VideoPreviewProps>(
  function VideoPreview(
    { videoId, title, author, duration, thumbnailUrl },
    ref
  ) {
    const iframeRef = useRef<HTMLIFrameElement>(null);
    const [isPlaying, setIsPlaying] = useState(false);

    useImperativeHandle(ref, () => ({
      seekTo: (seconds: number) => {
        if (iframeRef.current && iframeRef.current.contentWindow) {
          // Use YouTube IFrame API postMessage
          iframeRef.current.contentWindow.postMessage(
            JSON.stringify({
              event: "command",
              func: "seekTo",
              args: [seconds, true],
            }),
            "*"
          );
        }
        // Also start playing if not already
        if (!isPlaying) {
          setIsPlaying(true);
        }
      },
    }));

    const handlePlayClick = () => {
      setIsPlaying(true);
    };

    return (
      <div className="space-y-4">
        {/* Video Player */}
        <div className="relative aspect-video rounded-xl overflow-hidden bg-black">
          {isPlaying ? (
            <iframe
              ref={iframeRef}
              src={`https://www.youtube.com/embed/${videoId}?autoplay=1&enablejsapi=1`}
              title={title}
              className="absolute inset-0 w-full h-full"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowFullScreen
            />
          ) : (
            <button
              onClick={handlePlayClick}
              className="absolute inset-0 w-full h-full group"
            >
              {thumbnailUrl ? (
                <img
                  src={thumbnailUrl}
                  alt={title}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full bg-muted flex items-center justify-center">
                  <span className="text-muted-foreground">视频预览</span>
                </div>
              )}
              <div className="absolute inset-0 bg-black/30 group-hover:bg-black/40 transition-colors flex items-center justify-center">
                <div className="w-16 h-16 rounded-full bg-white/90 flex items-center justify-center group-hover:scale-110 transition-transform">
                  <Play className="w-8 h-8 text-black fill-black ml-1" />
                </div>
              </div>
            </button>
          )}
        </div>

        {/* Metadata */}
        <div className="space-y-2">
          <h1 className="text-xl font-semibold leading-tight">{title}</h1>
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            {author && (
              <div className="flex items-center gap-1.5">
                <User className="w-4 h-4" />
                <span>{author}</span>
              </div>
            )}
            {duration && duration > 0 && (
              <div className="flex items-center gap-1.5">
                <Clock className="w-4 h-4" />
                <span>{formatDuration(duration)}</span>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }
);
