import { YouTubeUrlInfo } from "@/types/video";

const YOUTUBE_REGEX = /^(?:https?:\/\/)?(?:www\.)?(?:youtube\.com\/watch\?v=|youtu\.be\/)([a-zA-Z0-9_-]{11})(?:\S+)?$/;

export function validateYouTubeUrl(url: string): YouTubeUrlInfo {
  const match = url.trim().match(YOUTUBE_REGEX);
  
  if (!match) {
    return {
      isValid: false,
      videoId: null,
      url: null,
    };
  }

  const videoId = match[1];
  const normalizedUrl = `https://www.youtube.com/watch?v=${videoId}`;

  return {
    isValid: true,
    videoId,
    url: normalizedUrl,
  };
}

export function extractVideoId(url: string): string | null {
  const info = validateYouTubeUrl(url);
  return info.videoId;
}

export function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  }

  return `${minutes}:${secs.toString().padStart(2, '0')}`;
}