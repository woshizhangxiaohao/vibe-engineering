import { apiClient } from "./client";
import { VideoMetadata, AnalysisResult, HistoryItem, YoutubeMetadata, PlaylistVideo, CaptionTrack, QuotaStatus } from "@/types/video";

export const videoApi = {
  getMetadata: (url: string) => 
    apiClient.post<VideoMetadata>("/v1/videos/metadata", { url }),
  
  analyze: (videoId: string, targetLanguage: string) =>
    apiClient.post<{ jobId: string; status: string }>("/v1/videos/analyze", { videoId, targetLanguage }),
  
  getResult: (jobId: string) =>
    apiClient.get<AnalysisResult>(`/v1/videos/result/${jobId}`),
  
  getHistory: () =>
    apiClient.get<{ items: HistoryItem[] }>("/v1/history"),
  
  export: (videoId: string, format: 'pdf' | 'markdown') =>
    apiClient.post<{ downloadUrl: string; fileName: string }>("/v1/videos/export", { videoId, format }),
};

export const youtubeApi = {
  getVideo: (idOrUrl: string) =>
    apiClient.get<YoutubeMetadata & { cached: boolean }>("/v1/youtube/video", { params: { id: idOrUrl } }),
  
  getPlaylist: (playlistId: string) =>
    apiClient.get<{ items: PlaylistVideo[]; cached: boolean }>("/v1/youtube/playlist", { params: { playlistId } }),
  
  getCaptions: (videoId: string) =>
    apiClient.get<{ captions: CaptionTrack[]; cached: boolean }>("/v1/youtube/captions", { params: { videoId } }),
  
  getQuota: () =>
    apiClient.get<QuotaStatus>("/v1/youtube/quota"),
    
  getAuthUrl: () =>
    apiClient.get<{ url: string }>("/v1/youtube/auth/url"),
};

export const contentApi = {
  parseUrl: (url: string) =>
    apiClient.post<any>("/parse", { url }),
};