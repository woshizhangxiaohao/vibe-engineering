export interface YouTubeUrlInfo {
  isValid: boolean;
  videoId: string | null;
  url: string;
}

const YOUTUBE_REGEX = /^(?:https?:\/\/)?(?:www\.)?(?:youtube\.com\/watch\?v=|youtu\.be\/)([a-zA-Z0-9_-]{11})(?:[&?].*)?$/;

export function parseYouTubeUrl(url: string): YouTubeUrlInfo {
  const trimmedUrl = url.trim();
  
  if (!trimmedUrl) {
    return {
      isValid: false,
      videoId: null,
      url: trimmedUrl,
    };
  }

  const match = trimmedUrl.match(YOUTUBE_REGEX);
  
  if (!match) {
    return {
      isValid: false,
      videoId: null,
      url: trimmedUrl,
    };
  }

  return {
    isValid: true,
    videoId: match[1],
    url: trimmedUrl,
  };
}

export function extractVideoId(url: string): string | null {
  return parseYouTubeUrl(url).videoId;
}

export function isValidYouTubeUrl(url: string): boolean {
  return parseYouTubeUrl(url).isValid;
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