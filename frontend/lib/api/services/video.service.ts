import { apiClient } from "../client";
import { VideoMetadata, VideoNote } from "@/types/video";

export const videoService = {
  async parseYouTubeUrl(url: string): Promise<VideoMetadata> {
    const response = await apiClient.post<VideoMetadata>('/api/v1/notes/parse-youtube', {
      url,
    });
    return response.data;
  },

  async saveVideoNote(note: Omit<VideoNote, 'id' | 'createdAt' | 'updatedAt'>): Promise<VideoNote> {
    const response = await apiClient.post<VideoNote>('/api/v1/notes/video', note);
    return response.data;
  },
};