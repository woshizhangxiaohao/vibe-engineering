"use client";

import { useState, useCallback, useEffect } from "react";
import { VideoUrlInput } from "./video-url-input";
import { VideoMetadataPreview } from "./video-metadata-preview";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent } from "@/components/ui/card";
import { useVideoMetadata } from "@/lib/api/hooks/use-video-metadata";
import { useVideoNote } from "@/lib/api/hooks/use-video-note";
import { NoteFormData } from "@/types/video";

interface VideoNoteFormProps {
  initialData?: NoteFormData;
  onSuccess?: () => void;
}

export function VideoNoteForm({ initialData, onSuccess }: VideoNoteFormProps) {
  const { metadata, loading, fetchMetadata, clearMetadata } = useVideoMetadata();
  const { saving, saveNote } = useVideoNote();
  const [validatedUrl, setValidatedUrl] = useState<string | null>(null);
  const [formData, setFormData] = useState<NoteFormData>({
    title: initialData?.title || "",
    content: initialData?.content || "",
    videoURL: initialData?.videoURL,
    videoDuration: initialData?.videoDuration,
    videoThumbnail: initialData?.videoThumbnail,
    videoSource: initialData?.videoSource,
  });

  useEffect(() => {
    if (metadata && !formData.title) {
      setFormData((prev) => ({
        ...prev,
        title: metadata.title,
      }));
    }
  }, [metadata, formData.title]);

  const handleUrlValidated = useCallback((url: string) => {
    setValidatedUrl(url);
    fetchMetadata(url);
  }, [fetchMetadata]);

  const handleRemove = useCallback(() => {
    clearMetadata();
    setValidatedUrl(null);
    setFormData((prev) => ({
      ...prev,
      videoURL: undefined,
      videoDuration: undefined,
      videoThumbnail: undefined,
      videoSource: undefined,
    }));
  }, [clearMetadata]);

  const handleSave = useCallback(async () => {
    if (!metadata) return;

    if (!formData.title.trim()) {
      return;
    }

    await saveNote({
      videoUrl: metadata.url,
      videoId: metadata.videoId,
      title: formData.title,
      duration: metadata.duration,
      thumbnail: metadata.thumbnail,
      content: formData.content,
    });

    setFormData({
      title: "",
      content: "",
    });
    handleRemove();
    onSuccess?.();
  }, [metadata, formData, saveNote, handleRemove, onSuccess]);

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
          
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="title">标题</Label>
              <Input
                id="title"
                type="text"
                placeholder="笔记标题"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                disabled={saving}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="content">内容（可选）</Label>
              <Textarea
                id="content"
                placeholder="笔记内容"
                value={formData.content}
                onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                disabled={saving}
                rows={5}
              />
            </div>
          </div>

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
              disabled={saving || !formData.title.trim()}
            >
              {saving ? "保存中..." : "保存笔记"}
            </Button>
          </div>
        </div>
      )}

      {initialData?.videoURL && !metadata && (
        <div className="space-y-4">
          <Card className="bg-card">
            <CardContent className="p-4">
              <div className="flex gap-4">
                <div className="relative aspect-video w-40 shrink-0 overflow-hidden rounded-md bg-muted">
                  {initialData.videoThumbnail && (
                    <img
                      src={initialData.videoThumbnail}
                      alt={initialData.title}
                      className="h-full w-full object-cover"
                    />
                  )}
                </div>
                <div className="flex min-w-0 flex-1 flex-col gap-2">
                  <h3 className="line-clamp-2 text-sm font-medium text-foreground">
                    {initialData.title}
                  </h3>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleRemove}
                  >
                    删除视频关联
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="title">标题</Label>
              <Input
                id="title"
                type="text"
                placeholder="笔记标题"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                disabled={saving}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="content">内容（可选）</Label>
              <Textarea
                id="content"
                placeholder="笔记内容"
                value={formData.content}
                onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                disabled={saving}
                rows={5}
              />
            </div>
          </div>

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
              disabled={saving || !formData.title.trim()}
            >
              {saving ? "保存中..." : "保存笔记"}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}