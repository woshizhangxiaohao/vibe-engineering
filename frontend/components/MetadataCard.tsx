"use client";

import React from 'react';
import { Card, CardContent } from "@/components/ui/card";
import { AspectRatio } from "@/components/ui/aspect-ratio";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { YoutubeMetadata } from "@/types/video";
import { Clock, User, Database } from "lucide-react";

interface MetadataCardProps {
  data: YoutubeMetadata & { cached?: boolean } | null;
  loading: boolean;
}

export default function MetadataCard({ data, loading }: MetadataCardProps) {
  if (loading) {
    return (
      <Card className="border-0 rounded-2xl bg-card overflow-hidden">
        <div className="p-6 space-y-6">
          <AspectRatio ratio={16 / 9}>
            <Skeleton className="w-full h-full rounded-xl" />
          </AspectRatio>
          <div className="space-y-3">
            <Skeleton className="h-8 w-3/4" />
            <Skeleton className="h-4 w-1/4" />
            <div className="space-y-2 pt-4">
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-2/3" />
            </div>
          </div>
        </div>
      </Card>
    );
  }

  if (!data) return null;

  return (
    <Card className="border-0 rounded-2xl bg-card overflow-hidden animate-in fade-in slide-in-from-bottom-4 duration-500">
      <CardContent className="p-6 md:p-8">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
          <div className="lg:col-span-5">
            <AspectRatio ratio={16 / 9} className="bg-muted rounded-xl overflow-hidden">
              <img
                src={data.thumbnailUrl}
                alt={data.title}
                className="object-cover w-full h-full"
              />
            </AspectRatio>
          </div>
          <div className="lg:col-span-7 space-y-4">
            <div className="flex items-start justify-between gap-4">
              <h2 className="text-2xl md:text-3xl font-bold tracking-tight leading-tight">
                {data.title}
              </h2>
              {data.cached && (
                <Badge variant="secondary" className="bg-secondary/50 text-muted-foreground border-0 rounded-md">
                  <Database className="h-3 w-3 mr-1" />
                  Cached
                </Badge>
              )}
            </div>

            <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-1.5">
                <User className="h-4 w-4" />
                <span className="font-medium">{data.author}</span>
              </div>
              <div className="flex items-center gap-1.5">
                <Clock className="h-4 w-4" />
                <span>{data.duration}s</span>
              </div>
            </div>

            <p className="text-muted-foreground leading-relaxed line-clamp-4 text-sm md:text-base">
              {data.description}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}