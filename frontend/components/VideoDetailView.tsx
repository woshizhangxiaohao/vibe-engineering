"use client";

import React, { useState, useEffect, useRef } from 'react';
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card } from "@/components/ui/card";
import { ChevronLeft, Download, Loader2, Share2, Play, FileText, MessageSquare, Sparkles } from "lucide-react";
import { VideoMetadata, AnalysisResult } from "@/types/video";
import { videoApi } from "@/lib/api/endpoints";
import { toast } from "@/lib/utils/toast";
import SummaryPanel from "./SummaryPanel";
import TranscriptionPanel from "./TranscriptionPanel";
import ChatConsole from "./ChatConsole";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";

interface VideoDetailViewProps {
  metadata: VideoMetadata;
  jobId: string;
  onBack: () => void;
}

export default function VideoDetailView({ metadata, jobId, onBack }: VideoDetailViewProps) {
  const [result, setResult] = useState<AnalysisResult | null>(null);
  const [currentTime, setCurrentTime] = useState(0);
  const [isExporting, setIsExporting] = useState(false);
  const playerRef = useRef<any>(null);

  useEffect(() => {
    const pollResult = async () => {
      try {
        const data = await videoApi.getResult(jobId);
        setResult(data);
        if (data.status === 'pending' || data.status === 'processing') {
          setTimeout(pollResult, 3000);
        } else if (data.status === 'failed') {
          toast.error("Analysis failed. Please try again.");
        }
      } catch (error) {
        console.error("Polling error:", error);
      }
    };

    pollResult();
  }, [jobId]);

  // Mock time tracking (in real app, use YouTube IFrame API events)
  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentTime(prev => prev + 1);
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  const handleSeek = (seconds: number) => {
    // In real implementation: playerRef.current.seekTo(seconds)
    setCurrentTime(seconds);
    toast.info(`Jumping to ${seconds}s`);
  };

  const handleExport = async (format: 'pdf' | 'markdown') => {
    if (!result || result.status !== 'completed') return;
    setIsExporting(true);
    try {
      const { downloadUrl } = await videoApi.export(metadata.videoId, format);
      window.open(downloadUrl, '_blank');
      toast.success(`Exported as ${format.toUpperCase()}`);
    } catch (error) {
      toast.error("Export failed");
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Fixed Header */}
      <header className="fixed top-0 left-0 right-0 z-50 bg-background border-b border-border/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Back Button & Title */}
            <div className="flex items-center gap-3 min-w-0 flex-1">
              <Button 
                variant="ghost" 
                size="icon" 
                onClick={onBack} 
                className="shrink-0 rounded-lg"
              >
                <ChevronLeft className="h-5 w-5" />
              </Button>
              <div className="min-w-0">
                <h1 className="text-base md:text-lg font-semibold tracking-tight truncate">
                  {metadata.title}
                </h1>
                <p className="text-xs text-muted-foreground truncate">
                  by {metadata.author}
                </p>
              </div>
            </div>
            
            {/* Actions */}
            <div className="flex items-center gap-2 shrink-0 ml-4">
              <Button 
                variant="ghost" 
                size="icon" 
                className="rounded-lg"
              >
                <Share2 className="h-4 w-4" />
              </Button>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button 
                    className="rounded-lg h-10 px-4 bg-primary text-primary-foreground hover:bg-primary/90 transition-colors duration-200"
                    disabled={!result || result.status !== 'completed' || isExporting}
                  >
                    {isExporting ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <>
                        <Download className="h-4 w-4 md:mr-2" />
                        <span className="hidden md:inline">Export</span>
                      </>
                    )}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="rounded-lg border-0 bg-background">
                  <DropdownMenuItem 
                    onClick={() => handleExport('pdf')} 
                    className="rounded cursor-pointer"
                  >
                    <FileText className="h-4 w-4 mr-2" />
                    PDF Document
                  </DropdownMenuItem>
                  <DropdownMenuItem 
                    onClick={() => handleExport('markdown')} 
                    className="rounded cursor-pointer"
                  >
                    <MessageSquare className="h-4 w-4 mr-2" />
                    Markdown File
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="pt-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 md:py-8">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 lg:gap-8">
            
            {/* Left: Player Area */}
            <div className="lg:col-span-7 space-y-4 md:space-y-6">
              {/* Video Player */}
              <Card className="overflow-hidden border-0 rounded-xl bg-muted aspect-video">
                <iframe
                  width="100%"
                  height="100%"
                  src={`https://www.youtube.com/embed/${metadata.videoId}?enablejsapi=1`}
                  title="YouTube video player"
                  frameBorder="0"
                  allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                  allowFullScreen
                  className="w-full h-full"
                />
              </Card>
              
              {/* Author Info Card */}
              <Card className="border-0 rounded-xl p-5 bg-card">
                <div className="flex items-center gap-4">
                  <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary font-bold text-lg">
                    {metadata.author.charAt(0).toUpperCase()}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h4 className="font-semibold truncate">{metadata.author}</h4>
                    <p className="text-sm text-muted-foreground">Video Creator</p>
                  </div>
                  <Button 
                    variant="ghost" 
                    size="sm" 
                    className="shrink-0 rounded-lg"
                  >
                    View Channel
                  </Button>
                </div>
              </Card>
            </div>

            {/* Right: Content Area */}
            <div className="lg:col-span-5 lg:h-[calc(100vh-120px)] lg:sticky lg:top-24">
              <Card className="h-full border-0 rounded-xl overflow-hidden bg-card">
                <Tabs defaultValue="summary" className="h-full flex flex-col">
                  {/* Tab Header */}
                  <div className="p-4 border-b border-border/50">
                    <TabsList className="grid w-full grid-cols-3 rounded-lg p-1 bg-muted h-11">
                      <TabsTrigger
                        value="summary"
                        className="rounded data-[state=active]:bg-background font-medium"
                      >
                        <Play className="h-4 w-4 mr-2" />
                        Summary
                      </TabsTrigger>
                      <TabsTrigger
                        value="transcription"
                        className="rounded data-[state=active]:bg-background font-medium"
                      >
                        <FileText className="h-4 w-4 mr-2" />
                        Transcript
                      </TabsTrigger>
                      <TabsTrigger
                        value="chat"
                        className="rounded data-[state=active]:bg-background font-medium"
                      >
                        <Sparkles className="h-4 w-4 mr-2" />
                        Chat
                      </TabsTrigger>
                    </TabsList>
                  </div>

                  {/* Tab Content */}
                  <div className="flex-1 overflow-hidden">
                    {!result || result.status === 'processing' || result.status === 'pending' ? (
                      /* Loading State */
                      <div className="flex flex-col items-center justify-center h-full p-8 text-center">
                        <div className="h-16 w-16 rounded-xl bg-muted flex items-center justify-center mb-6">
                          <Loader2 className="h-8 w-8 animate-spin text-primary" />
                        </div>
                        <h3 className="font-semibold text-lg mb-2">AI is analyzing...</h3>
                        <p className="text-sm text-muted-foreground max-w-xs leading-relaxed">
                          We're processing the audio and generating your insights. This usually takes less than 30 seconds.
                        </p>
                        {/* Progress Indicator */}
                        <div className="mt-6 w-full max-w-xs">
                          <div className="h-1 bg-muted rounded-full overflow-hidden">
                            <div className="h-full bg-primary rounded-full animate-pulse" style={{ width: '60%' }} />
                          </div>
                        </div>
                      </div>
                    ) : result.status === 'failed' ? (
                      /* Error State */
                      <div className="flex flex-col items-center justify-center h-full p-8 text-center">
                        <div className="h-16 w-16 rounded-xl bg-destructive/10 flex items-center justify-center mb-6">
                          <span className="text-2xl">⚠️</span>
                        </div>
                        <h3 className="font-semibold text-lg mb-2">Analysis Failed</h3>
                        <p className="text-sm text-muted-foreground max-w-xs leading-relaxed mb-4">
                          Something went wrong. Please try again or choose a different video.
                        </p>
                        <Button 
                          variant="outline" 
                          onClick={onBack}
                          className="rounded-lg"
                        >
                          Go Back
                        </Button>
                      </div>
                    ) : (
                      /* Content */
                      <>
                        <TabsContent value="summary" className="h-full mt-0 overflow-y-auto no-scrollbar p-4">
                          <SummaryPanel result={result} onSeek={handleSeek} />
                        </TabsContent>
                        <TabsContent value="transcription" className="h-full mt-0 overflow-hidden">
                          <TranscriptionPanel
                            result={result}
                            currentTime={currentTime}
                            onSeek={handleSeek}
                          />
                        </TabsContent>
                        <TabsContent value="chat" className="h-full mt-0 overflow-hidden">
                          {result.analysisId && (
                            <ChatConsole
                              analysisId={result.analysisId}
                              className="h-full rounded-none"
                            />
                          )}
                        </TabsContent>
                      </>
                    )}
                  </div>
                </Tabs>
              </Card>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
