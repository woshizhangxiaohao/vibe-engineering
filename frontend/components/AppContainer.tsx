"use client";

import React, { useState, useEffect } from 'react';
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Search, Loader2, History, Clock, ArrowRight, Sparkles, Database } from "lucide-react";
import { videoApi } from '@/lib/api/endpoints';
import { VideoMetadata, HistoryItem } from '@/types/video';
import { toast } from '@/lib/utils/toast';
import VideoDetailView from './VideoDetailView';
import { Card, CardContent } from "@/components/ui/card";
import Header from './Header';
import YoutubeDashboard from './YoutubeDashboard';

export default function AppContainer() {
  const [url, setUrl] = useState('');
  const [language, setLanguage] = useState('en');
  const [loading, setLoading] = useState(false);
  const [history, setHistory] = useState<HistoryItem[]>([]);
  const [activeView, setActiveView] = useState<{ metadata: VideoMetadata; jobId: string } | null>(null);
  const [showDashboard, setShowDashboard] = useState(false);

  useEffect(() => {
    const loadHistory = async () => {
      try {
        const data = await videoApi.getHistory();
        setHistory(data.items || []);
      } catch (error) {
        console.error("Failed to load history");
      }
    };
    loadHistory();
  }, []);

  const handleStartAnalysis = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url.trim()) return;

    setLoading(true);
    try {
      const metadata = await videoApi.getMetadata(url);
      const { jobId } = await videoApi.analyze(metadata.videoId, language);
      setActiveView({ metadata, jobId });
      toast.success("Analysis started!");
    } catch (error: any) {
      toast.error(error.message || "Please enter a valid public YouTube link");
    } finally {
      setLoading(false);
    }
  };

  if (showDashboard) {
    return <YoutubeDashboard />;
  }

  if (activeView) {
    return <VideoDetailView 
      metadata={activeView.metadata} 
      jobId={activeView.jobId} 
      onBack={() => setActiveView(null)} 
    />;
  }

  return (
    <div className="min-h-screen bg-background">
      <Header />

      <main className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 pt-32 pb-16 md:pt-40 md:pb-24">
        <div className="text-center mb-16 md:mb-20 space-y-6">
          <div 
            className="inline-flex items-center rounded-full px-4 py-2 text-sm font-medium bg-secondary animate-fade-in-up"
            style={{ animationDelay: '0ms' }}
          >
            <Sparkles className="h-4 w-4 mr-2 text-primary" />
            <span>YouTube AI Intelligence</span>
            <span className="ml-2 flex h-2 w-2 rounded-full bg-primary animate-pulse" />
          </div>
          
          <h1 
            className="text-5xl sm:text-6xl md:text-7xl lg:text-8xl font-black tracking-tighter leading-none animate-fade-in-up"
            style={{ animationDelay: '100ms' }}
          >
            VIBE
            <br />
            <span className="text-primary">INTELLIGENCE.</span>
          </h1>
          
          <p 
            className="mx-auto max-w-xl text-lg md:text-xl text-muted-foreground leading-relaxed animate-fade-in-up"
            style={{ animationDelay: '200ms' }}
          >
            Transform any YouTube video into structured insights, summaries, and searchable transcripts in seconds.
          </p>

          <div className="pt-4 animate-fade-in-up" style={{ animationDelay: '250ms' }}>
            <Button 
              variant="secondary" 
              onClick={() => setShowDashboard(true)}
              className="rounded-full px-8 h-12 border-0 bg-secondary hover:bg-secondary/80 text-primary font-semibold"
            >
              <Database className="h-4 w-4 mr-2" />
              Open Data Dashboard
            </Button>
          </div>
        </div>

        <form 
          onSubmit={handleStartAnalysis} 
          className="space-y-4 mb-20 md:mb-28 animate-fade-in-up"
          style={{ animationDelay: '300ms' }}
        >
          <div className="relative flex items-center">
            <Search className="absolute left-5 text-muted-foreground h-5 w-5 transition-colors pointer-events-none z-10" />
            <Input
              type="url"
              placeholder="Paste YouTube link here..."
              className="pl-14 pr-36 md:pr-44 h-14 md:h-16 text-base md:text-lg rounded-lg border-0 bg-muted focus:bg-background focus:outline-none transition-colors duration-200"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              disabled={loading}
            />
            <Button 
              type="submit" 
              className="absolute right-2 rounded-lg h-10 md:h-12 px-5 md:px-8 text-sm md:text-base font-medium border-0 bg-primary text-primary-foreground hover:bg-primary/90 focus:bg-primary/90 focus:outline-none active:scale-[0.98] transition-all duration-200"
              disabled={loading || !url}
            >
              {loading ? (
                <Loader2 className="h-5 w-5 animate-spin" />
              ) : (
                <>
                  Analyze
                  <ArrowRight className="h-4 w-4 ml-2 hidden md:block" />
                </>
              )}
            </Button>
          </div>
          
          <div className="flex justify-center">
            <Select value={language} onValueChange={setLanguage}>
              <SelectTrigger className="w-[180px] md:w-[200px] rounded-lg h-10 border-0 bg-muted hover:bg-muted/80 focus:bg-background focus:outline-none transition-colors">
                <SelectValue placeholder="Target Language" />
              </SelectTrigger>
              <SelectContent className="rounded-lg border-0 bg-background">
                <SelectItem value="en" className="rounded cursor-pointer focus:bg-muted">English</SelectItem>
                <SelectItem value="zh" className="rounded cursor-pointer focus:bg-muted">Chinese (Simplified)</SelectItem>
                <SelectItem value="ja" className="rounded cursor-pointer focus:bg-muted">Japanese</SelectItem>
                <SelectItem value="ko" className="rounded cursor-pointer focus:bg-muted">Korean</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </form>

        <div 
          className="space-y-6 md:space-y-8 animate-fade-in-up"
          style={{ animationDelay: '400ms' }}
        >
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-lg bg-muted flex items-center justify-center">
              <History className="h-5 w-5 text-muted-foreground" />
            </div>
            <div>
              <h2 className="text-xl font-bold tracking-tight text-foreground">Recent Analysis</h2>
              <p className="text-sm text-muted-foreground">Your previously analyzed videos</p>
            </div>
          </div>

          {history.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6">
              {history.map((item, index) => (
                <Card 
                  key={`${item.videoId}-${index}`} 
                  className="group overflow-hidden border-0 hover:bg-muted/50 transition-colors duration-200 rounded-xl bg-card cursor-pointer"
                  onClick={() => setUrl(`https://youtube.com/watch?v=${item.videoId}`)}
                  style={{ animationDelay: `${500 + index * 100}ms` }}
                >
                  <div className="aspect-video relative overflow-hidden bg-muted">
                    {item.thumbnailUrl ? (
                      <img 
                        src={item.thumbnailUrl} 
                        alt={item.title} 
                        className="w-full h-full object-cover transition-transform duration-300 ease-out group-hover:scale-105"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <span className="text-muted-foreground text-sm">No thumbnail</span>
                      </div>
                    )}
                    <div className="absolute inset-0 bg-primary/90 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex items-center justify-center">
                      <div className="h-12 w-12 rounded-full bg-white flex items-center justify-center">
                        <Search className="text-primary h-5 w-5" />
                      </div>
                    </div>
                  </div>
                  
                  <CardContent className="p-4 md:p-5">
                    <h3 className="font-semibold line-clamp-2 mb-3 group-hover:text-primary transition-colors duration-200">
                      {item.title}
                    </h3>
                    <div className="flex items-center text-xs text-muted-foreground uppercase tracking-wide font-medium">
                      <Clock className="h-3 w-3 mr-1.5" />
                      {new Date(item.createdAt).toLocaleDateString('en-US', {
                        month: 'short',
                        day: 'numeric',
                        year: 'numeric'
                      })}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : (
            <div className="text-center py-16 md:py-24 rounded-xl border-0 bg-muted/30">
              <div className="h-16 w-16 rounded-xl bg-muted mx-auto mb-4 flex items-center justify-center">
                <History className="h-8 w-8 text-muted-foreground" />
              </div>
              <h3 className="font-semibold text-lg mb-2">No recent records</h3>
              <p className="text-muted-foreground">Start your first analysis above!</p>
            </div>
          )}
        </div>
      </main>

      <footer className="border-t border-border/50 py-8 mt-16">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <div className="h-6 w-6 rounded bg-primary" />
            <span className="font-semibold">VIBE</span>
          </div>
          <p className="text-sm text-muted-foreground">
            Built on Base. © {new Date().getFullYear()} All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}