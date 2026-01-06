export interface VideoMetadata {
  title: string;
  duration: number;
  thumbnail: string;
  videoId: string;
  url: string;
}

export interface VideoNote {
  id?: string;
  videoUrl: string;
  videoId: string;
  title: string;
  duration: number;
  thumbnail: string;
  content?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface YouTubeUrlInfo {
  isValid: boolean;
  videoId: string | null;
  url: string | null;
}

export interface NoteFormData {
  title: string;
  content: string;
  videoURL?: string;
  videoDuration?: number;
  videoThumbnail?: string;
  videoSource?: string;
}