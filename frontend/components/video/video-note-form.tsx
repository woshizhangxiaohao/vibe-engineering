"use client";

import { useState, useCallback } from "react";
import { VideoUrlInput } from "./video-url-input";
import { VideoMetadataPreview } from "./video-metadata-preview";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent } from "@/components/ui/card";
import { useVideoMetadata } from "@/lib/api/hooks/use-video-metadata";
import { useVideoNote } from "@/lib/api/hooks/use-video-note";

export function VideoNoteForm() {
  const { metadata, loading, fetchMetadata, clearMetadata } = useVideoMetadata();
  const { saving, saveNote } = useVideoNote();
  const [validatedUrl, setValidatedUrl] = useState<string | null>(null);

  const handleUrlValidated = useCallback((url: string) => {
    setValidatedUrl(url);
    fetchMetadata(url);
  }, [fetchMetadata]);

  const handleRemove = useCallback(() => {
    clearMetadata();
    setValidatedUrl(null);
  }, [clearMetadata]);

  const handleSave = useCallback(async () => {
    if (!metadata) return;

    await saveNote({
      videoUrl: metadata.url,
      videoId: metadata.videoId,
      title: metadata.title,
      duration: metadata.duration,
      thumbnail: metadata.thumbnail,
    });

    handleRemove();
  }, [metadata, saveNote, handleRemove]);

  return (
    <div className="space-y-4">
      <VideoUrlInput
        onUrlValidated={handleUrlValidated}
        disabled={loading || !!metadata}
      />

      {loading && (
        <Card className="bg-card">
          <CardContent className="p-4">
            <div className="flex gap-4">
              <Skeleton className="aspect-video w-40 shrink-0" />
              <div className="flex min-w-0 flex-1 flex-col gap-2">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-4 w-1/2" />
                <Skeleton className="h-6 w-16" />
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {metadata && !loading && (
        <div className="space-y-4">
          <VideoMetadataPreview
            metadata={metadata}
            onRemove={handleRemove}
          />
          
          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={handleRemove}
              disabled={saving}
            >
              取消
            </Button>
            <Button
              onClick={handleSave}
              disabled={saving}
            >
              {saving ? "保存中..." : "保存笔记"}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}