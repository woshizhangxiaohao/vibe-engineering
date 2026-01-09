"use client";

import React, { useState } from 'react';
import Sidebar from "./Sidebar";
import SearchInputGroup from "./SearchInputGroup";
import MetadataCard from "./MetadataCard";
import AuthPanel from "./AuthPanel";
import { youtubeApi } from "@/lib/api/endpoints";
import { YoutubeMetadata, PlaylistVideo, CaptionTrack } from "@/types/video";
import { toast } from "@/lib/utils/toast";
import { Card, CardContent } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle, FileText, ListVideo } from "lucide-react";

export default function YoutubeDashboard() {
  const [activeTab, setActiveTab] = useState('video');
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const [videoData, setVideoData] = useState<(YoutubeMetadata & { cached: boolean }) | null>(null);
  const [playlistData, setPlaylistData] = useState<PlaylistVideo[]>([]);
  const [captionData, setCaptionData] = useState<CaptionTrack[]>([]);

  const handleSearch = async () => {
    if (!query.trim()) return;
    setLoading(true);
    setError(null);
    
    try {
      if (activeTab === 'video') {
        const data = await youtubeApi.getVideo(query);
        setVideoData(data);
      } else if (activeTab === 'playlist') {
        const data = await youtubeApi.getPlaylist(query);
        setPlaylistData(data.items);
      } else if (activeTab === 'captions') {
        const data = await youtubeApi.getCaptions(query);
        setCaptionData(data.captions);
      }
    } catch (e: any) {
      const msg = e.status === 404 ? "Resource not found" : e.status === 429 ? "API Quota exhausted" : "Failed to fetch data";
      setError(msg);
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen bg-[#f9f9f9]">
      <Sidebar activeTab={activeTab} onTabChange={setActiveTab} />
      
      <main className="flex-1 p-6 md:p-12 lg:p-16">
        <div className="max-w-5xl mx-auto space-y-12">
          {/* Header Section */}
          <div className="space-y-4">
            <h1 className="text-4xl font-black tracking-tighter uppercase">
              {activeTab === 'video' && "Video Intelligence"}
              {activeTab === 'playlist' && "Playlist Explorer"}
              {activeTab === 'captions' && "Caption Extractor"}
              {activeTab === 'auth' && "Security Center"}
            </h1>
            <p className="text-muted-foreground text-lg">
              {activeTab !== 'auth' ? "Extract structured data directly from YouTube Data API v3." : "Manage your API credentials and access levels."}
            </p>
          </div>

          {activeTab !== 'auth' ? (
            <div className="space-y-10">
              <SearchInputGroup
                value={query}
                onChange={setQuery}
                onSearch={handleSearch}
                loading={loading}
                error={!!error}
                placeholder={activeTab === 'playlist' ? "Enter Playlist ID..." : "Enter Video URL or ID..."}
              />

              {error && (
                <Alert variant="destructive" className="rounded-xl border-0 bg-destructive/10 text-destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertTitle>Error</AlertTitle>
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              <div className="space-y-6">
                {activeTab === 'video' && <MetadataCard data={videoData} loading={loading} />}
                
                {activeTab === 'playlist' && playlistData.length > 0 && (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {playlistData.map((item) => (
                      <Card key={item.videoId} className="border-0 rounded-xl bg-card hover:bg-muted/50 transition-colors">
                        <CardContent className="p-4 flex gap-4">
                          <img src={item.thumbnailUrl} className="w-24 aspect-video rounded-lg object-cover" alt="" />
                          <div className="flex-1 min-w-0">
                            <p className="font-semibold text-sm line-clamp-2">{item.title}</p>
                            <p className="text-xs text-muted-foreground mt-1">ID: {item.videoId}</p>
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                )}

                {activeTab === 'captions' && captionData.length > 0 && (
                  <div className="space-y-3">
                    {captionData.map((track) => (
                      <div key={track.languageCode} className="flex items-center justify-between p-4 bg-card rounded-xl">
                        <div className="flex items-center gap-3">
                          <FileText className="h-5 w-5 text-primary" />
                          <span className="font-medium">{track.languageName}</span>
                          {track.isAutoGenerated && <Badge variant="secondary" className="text-[10px] uppercase">Auto</Badge>}
                        </div>
                        <span className="text-xs text-muted-foreground font-mono">{track.languageCode}</span>
                      </div>
                    ))}
                  </div>
                )}

                {((activeTab === 'playlist' && playlistData.length === 0) || (activeTab === 'captions' && captionData.length === 0)) && !loading && !error && (
                  <div className="py-20 text-center bg-card rounded-2xl">
                    <ListVideo className="h-12 w-12 text-muted-foreground mx-auto mb-4 opacity-20" />
                    <p className="text-muted-foreground">No data extracted yet. Enter a valid ID to begin.</p>
                  </div>
                )}
              </div>
            </div>
          ) : (
            <AuthPanel />
          )}
        </div>
      </main>
    </div>
  );
}