import { apiClient } from "./client";
import {
  VideoMetadata,
  AnalysisResult,
  HistoryItem,
  YoutubeMetadata,
  PlaylistVideo,
  CaptionTrack,
  QuotaStatus,
} from "@/types/video";

export const videoApi = {
  getMetadata: (url: string) =>
    apiClient.post<VideoMetadata>("/v1/videos/metadata", { url }),

  analyze: (videoId: string, targetLanguage: string) =>
    apiClient.post<{ jobId: string; status: string }>("/v1/videos/analyze", {
      videoId,
      targetLanguage,
    }),

  getResult: (jobId: string) =>
    apiClient.get<AnalysisResult>(`/v1/videos/result/${jobId}`),

  getHistory: () => apiClient.get<{ items: HistoryItem[] }>("/v1/history"),

  export: (videoId: string, format: "pdf" | "markdown") =>
    apiClient.post<{ downloadUrl: string; fileName: string }>(
      "/v1/videos/export",
      { videoId, format }
    ),
};

export const youtubeApi = {
  getVideo: (idOrUrl: string) =>
    apiClient.get<YoutubeMetadata & { cached: boolean }>("/v1/youtube/video", {
      params: { input: idOrUrl },
    }),

  getPlaylist: (playlistId: string) =>
    apiClient.get<{ items: PlaylistVideo[]; cached: boolean }>(
      "/v1/youtube/playlist",
      { params: { playlistId } }
    ),

  getCaptions: (videoId: string) =>
    apiClient.get<{ captions: CaptionTrack[]; cached: boolean }>(
      "/v1/youtube/captions",
      { params: { videoId } }
    ),

  // Get transcript using yt-dlp (no OAuth required)
  getTranscript: (idOrUrl: string) =>
    apiClient.post<{
      videoId: string;
      title: string;
      author: string;
      duration: string;
      transcripts: Array<{
        start: string;
        end: string;
        text: string;
      }>;
    }>("/v1/transcript", { input: idOrUrl }),

  getQuota: () => apiClient.get<QuotaStatus>("/v1/system/quota"),

  getAuthUrl: async () => {
    const response = await apiClient.get<{ authUrl: string; url?: string }>(
      "/v1/auth/google/url"
    );
    return { url: response.authUrl || response.url || "" };
  },

  handleCallback: (code: string, state?: string) =>
    apiClient.post<{
      accessToken: string;
      refreshToken: string;
      tokenType: string;
      expiry: string;
      tokenJSON: string;
    }>("/v1/auth/google/callback", { code, state }),
};

export const contentApi = {
  parseUrl: (url: string) => apiClient.post<any>("/parse", { url }),
};
