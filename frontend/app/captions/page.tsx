"use client";

import React, { useState } from "react";
import { DashboardLayout } from "@/components/layout/dashboard-layout";
import SearchInputGroup from "@/components/SearchInputGroup";
import { youtubeApi } from "@/lib/api/endpoints";
import {
  CaptionsResponse,
  TranscriptSegment,
  TranscriptData,
} from "@/types/video";
import { toast } from "@/lib/utils/toast";
import { extractVideoId } from "@/lib/utils/youtube";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  AlertCircle,
  FileText,
  Clock,
  Languages,
  Copy,
  Check,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export default function CaptionsPage() {
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [captionData, setCaptionData] = useState<CaptionsResponse | null>(null);
  const [selectedLang, setSelectedLang] = useState<string>("");
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  const handleSearch = async () => {
    if (!query.trim()) return;
    setLoading(true);
    setError(null);
    setCaptionData(null);

    try {
      const videoId = extractVideoId(query);
      if (!videoId) {
        setError(
          "Invalid video URL or ID. Please enter a valid YouTube video link."
        );
        toast.error("Invalid video URL or ID");
        setLoading(false);
        return;
      }
      const data = await youtubeApi.getCaptions(videoId);
      setCaptionData(data);
      // Auto select first language
      if (data.language_code && data.language_code.length > 0) {
        setSelectedLang(data.language_code[0].code);
      }
    } catch (e: any) {
      // Extract error message from ApiError or use default
      let errorMsg = "Failed to fetch data";

      if (e instanceof Error) {
        // Use the error message directly (ApiError sets it from backend message)
        errorMsg = e.message;
      } else if (e?.message) {
        errorMsg = e.message;
      } else if (e?.data && typeof e.data === "object" && "message" in e.data) {
        errorMsg = String(e.data.message);
      } else {
        // Fallback to status-based messages
        if (e?.status === 401) {
          errorMsg = "Authorization required. Please authenticate with Google.";
        } else if (e?.status === 404) {
          errorMsg = "No captions available for this video";
        } else if (e?.status === 429) {
          errorMsg = "API Quota exhausted";
        }
      }

      setError(errorMsg);
      toast.error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  // Get segments from TranscriptData (priority: default > auto > custom)
  const getSegments = (
    data: TranscriptData | undefined
  ): TranscriptSegment[] => {
    if (!data) return [];
    if (data.default && data.default.length > 0) return data.default;
    if (data.auto && data.auto.length > 0) return data.auto;
    if (data.custom && data.custom.length > 0) return data.custom;
    return [];
  };

  const copyToClipboard = async (text: string, index: number) => {
    await navigator.clipboard.writeText(text);
    setCopiedIndex(index);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  const copyAllTranscript = async () => {
    if (!captionData || !selectedLang) return;
    const segments = getSegments(captionData.transcripts?.[selectedLang]);
    const fullText = segments.map((s) => s.text).join(" ");
    await navigator.clipboard.writeText(fullText);
    toast.success("Transcript copied to clipboard!");
  };

  const formatDuration = (seconds: string) => {
    const sec = parseInt(seconds, 10);
    if (isNaN(sec)) return seconds;
    const h = Math.floor(sec / 3600);
    const m = Math.floor((sec % 3600) / 60);
    const s = sec % 60;
    if (h > 0)
      return `${h}:${m.toString().padStart(2, "0")}:${s
        .toString()
        .padStart(2, "0")}`;
    return `${m}:${s.toString().padStart(2, "0")}`;
  };

  const currentTranscripts = getSegments(
    captionData?.transcripts?.[selectedLang]
  );

  return (
    <DashboardLayout
      title="Caption Extractor"
      description="Extract structured transcripts from YouTube videos."
    >
      <div className="space-y-6">
        <SearchInputGroup
          value={query}
          onChange={setQuery}
          onSearch={handleSearch}
          loading={loading}
          error={!!error}
          placeholder="Enter video URL or ID (e.g., g5hw1HZfwTc)"
        />

        {error && (
          <Alert
            variant="destructive"
            className="rounded-xl border-0 bg-destructive/10 text-destructive"
          >
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {captionData && (
          <div className="space-y-6">
            {/* Video Info Card */}
            <Card className="border-0 rounded-2xl bg-card overflow-hidden">
              <div className="flex flex-col md:flex-row">
                <div className="md:w-80 flex-shrink-0">
                  {/* #region agent log */}
                  {(() => {
                    fetch(
                      "http://127.0.0.1:7243/ingest/a127609d-0110-4a4e-83ea-2be1242c90c3",
                      {
                        method: "POST",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify({
                          location: "app/captions/page.tsx:123",
                          message: "Before accessing thumbnailUrl",
                          data: {
                            hasVideoInfo: !!captionData.videoInfo,
                            hasThumbnailUrl:
                              !!captionData.videoInfo?.thumbnailUrl,
                            thumbnailUrlType:
                              typeof captionData.videoInfo?.thumbnailUrl,
                          },
                          timestamp: Date.now(),
                          sessionId: "debug-session",
                          runId: "run1",
                          hypothesisId: "A,B,C,D",
                        }),
                      }
                    ).catch(() => {});
                    return null;
                  })()}
                  {/* #endregion */}
                  {(() => {
                    const thumbnailUrl =
                      captionData.videoInfo?.thumbnailUrl?.maxresdefault ||
                      captionData.videoInfo?.thumbnailUrl?.hqdefault;
                    return thumbnailUrl ? (
                      <img
                        src={thumbnailUrl}
                        alt={captionData.videoInfo?.name || "Video thumbnail"}
                        className="w-full aspect-video object-cover"
                      />
                    ) : (
                      <div className="w-full aspect-video bg-muted flex items-center justify-center">
                        <FileText className="h-12 w-12 text-muted-foreground opacity-20" />
                      </div>
                    );
                  })()}
                </div>
                <CardContent className="flex-1 p-5">
                  {/* #region agent log */}
                  {(() => {
                    fetch(
                      "http://127.0.0.1:7243/ingest/a127609d-0110-4a4e-83ea-2be1242c90c3",
                      {
                        method: "POST",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify({
                          location: "app/captions/page.tsx:135",
                          message:
                            "Accessing videoInfo.name with optional chaining",
                          data: {
                            hasVideoInfo: !!captionData.videoInfo,
                            hasName: !!captionData.videoInfo?.name,
                          },
                          timestamp: Date.now(),
                          sessionId: "debug-session",
                          runId: "run2",
                          hypothesisId: "VERIFY",
                        }),
                      }
                    ).catch(() => {});
                    return null;
                  })()}
                  {/* #endregion */}
                  <h3 className="font-semibold text-lg line-clamp-2 mb-3">
                    {captionData.videoInfo?.name || "Untitled Video"}
                  </h3>
                  <div className="flex flex-wrap gap-3 text-sm text-muted-foreground">
                    {captionData.videoInfo?.author && (
                      <span>{captionData.videoInfo?.author}</span>
                    )}
                    {captionData.videoInfo?.duration && (
                      <span className="flex items-center gap-1">
                        <Clock className="h-3.5 w-3.5" />
                        {formatDuration(captionData.videoInfo?.duration || "")}
                      </span>
                    )}
                    <span className="flex items-center gap-1">
                      <Languages className="h-3.5 w-3.5" />
                      {captionData.language_code?.length || 0} language(s)
                    </span>
                  </div>
                  <div className="flex flex-wrap gap-2 mt-4">
                    {captionData.language_code?.map((lang) => (
                      <Badge
                        key={lang.code}
                        variant={
                          selectedLang === lang.code ? "default" : "secondary"
                        }
                        className="cursor-pointer transition-colors"
                        onClick={() => setSelectedLang(lang.code)}
                      >
                        {lang.name} ({lang.code})
                      </Badge>
                    ))}
                  </div>
                </CardContent>
              </div>
            </Card>

            {/* Transcript Display */}
            {captionData.language_code &&
              captionData.language_code.length > 0 && (
                <Card className="border-0 rounded-2xl bg-card">
                  <CardHeader className="flex flex-row items-center justify-between pb-4">
                    <CardTitle className="text-base font-medium flex items-center gap-2">
                      <FileText className="h-4 w-4 text-primary" />
                      Transcript
                      <Badge variant="outline" className="ml-2">
                        {currentTranscripts.length} segments
                      </Badge>
                    </CardTitle>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={copyAllTranscript}
                      className="gap-2"
                    >
                      <Copy className="h-3.5 w-3.5" />
                      Copy All
                    </Button>
                  </CardHeader>
                  <CardContent>
                    {captionData.language_code &&
                      captionData.language_code.length > 1 && (
                        <Tabs
                          value={selectedLang}
                          onValueChange={setSelectedLang}
                          className="mb-4"
                        >
                          <TabsList className="bg-muted/50">
                            {captionData.language_code?.map((lang) => (
                              <TabsTrigger key={lang.code} value={lang.code}>
                                {lang.name}
                              </TabsTrigger>
                            ))}
                          </TabsList>
                        </Tabs>
                      )}

                    <div className="max-h-[500px] overflow-y-auto space-y-1 pr-2">
                      {currentTranscripts.map((segment, index) => (
                        <div
                          key={index}
                          className={cn(
                            "group flex gap-3 p-3 rounded-lg hover:bg-muted/50 transition-colors",
                            index % 2 === 0 ? "bg-muted/20" : ""
                          )}
                        >
                          <span className="text-xs text-muted-foreground font-mono min-w-[70px] pt-0.5">
                            {segment.start}
                          </span>
                          <p className="flex-1 text-sm leading-relaxed">
                            {segment.text}
                          </p>
                          <button
                            onClick={() => copyToClipboard(segment.text, index)}
                            className="opacity-0 group-hover:opacity-100 transition-opacity p-1 hover:bg-muted rounded"
                          >
                            {copiedIndex === index ? (
                              <Check className="h-3.5 w-3.5 text-green-500" />
                            ) : (
                              <Copy className="h-3.5 w-3.5 text-muted-foreground" />
                            )}
                          </button>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              )}
          </div>
        )}

        {!captionData && !loading && !error && (
          <div className="py-20 text-center bg-card rounded-2xl">
            <FileText className="h-12 w-12 text-muted-foreground mx-auto mb-4 opacity-20" />
            <p className="text-muted-foreground">
              No data extracted yet. Enter a valid video ID to begin.
            </p>
          </div>
        )}

        {loading && (
          <div className="py-20 text-center bg-card rounded-2xl">
            <div className="animate-spin h-8 w-8 border-2 border-primary border-t-transparent rounded-full mx-auto mb-4" />
            <p className="text-muted-foreground">Fetching captions...</p>
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
