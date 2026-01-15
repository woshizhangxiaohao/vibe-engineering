import { apiClient } from "./client";
import {
  VideoMetadata,
  AnalysisResult,
  HistoryItem,
  YoutubeMetadata,
  PlaylistVideo,
  CaptionTrack,
  QuotaStatus,
  CaptionsResponse,
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
    apiClient.get<CaptionsResponse>(
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

/**
 * Insight API endpoints
 */
import type {
  InsightsListResponse,
  CreateInsightRequest,
  CreateInsightResponse,
  InsightDetailResponse,
} from "./types";

export const insightApi = {
  /**
   * Get insights list grouped by time
   * @param limit - Number of items to return (default: 50)
   * @param offset - Offset for pagination (default: 0)
   * @param search - Search keyword (optional)
   */
  getInsights: (params?: {
    limit?: number;
    offset?: number;
    search?: string;
  }) =>
    apiClient.get<InsightsListResponse>("/v1/insights", {
      params: {
        limit: params?.limit || 50,
        offset: params?.offset || 0,
        ...(params?.search && { search: params.search }),
      },
    }),

  /**
   * Create a new insight parsing task
   * @param data - Insight creation data
   */
  createInsight: (data: CreateInsightRequest) =>
    apiClient.post<CreateInsightResponse>("/v1/insights", data),

  /**
   * Get insight detail by ID
   * @param id - Insight ID
   */
  getInsightDetail: (id: number) =>
    apiClient.get<InsightDetailResponse>(`/v1/insights/${id}`),

  /**
   * Reprocess a failed insight
   * @param id - Insight ID
   */
  reprocessInsight: (id: number) =>
    apiClient.post<{ status: string; message: string }>(
      `/v1/insights/${id}/process`
    ),
};
