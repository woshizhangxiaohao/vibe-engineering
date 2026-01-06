"use client";

import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { X } from "lucide-react";
import { VideoMetadata } from "@/types/video";
import { formatDuration } from "@/lib/utils/video";
import Image from "next/image";

interface VideoMetadataPreviewProps {
  metadata: VideoMetadata;
  onRemove: () => void;
}

export function VideoMetadataPreview({ metadata, onRemove }: VideoMetadataPreviewProps) {
  return (
    <Card className="bg-card">
      <CardContent className="p-4">
        <div className="flex gap-4">
          <div className="relative aspect-video w-40 shrink-0 overflow-hidden rounded-md bg-muted">
            <Image
              src={metadata.thumbnail}
              alt={metadata.title}
              fill
              className="object-cover"
              unoptimized
            />
          </div>
          
          <div className="flex min-w-0 flex-1 flex-col gap-2">
            <div className="flex items-start justify-between gap-2">
              <h3 className="line-clamp-2 text-sm font-medium text-foreground">
                {metadata.title}
              </h3>
              <Button
                variant="ghost"
                size="icon"
                onClick={onRemove}
                className="h-6 w-6 shrink-0"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
            
            <div className="flex items-center gap-2">
              <Badge variant="secondary" className="font-mono text-xs">
                {formatDuration(metadata.duration)}
              </Badge>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}