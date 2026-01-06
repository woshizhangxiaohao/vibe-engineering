import { apiClient } from '../client';
import type { VideoMetadata, VideoNote } from '@/types/video';

export interface GetVideoMetadataParams {
  url: string;
  signal?: AbortSignal;
}

export interface SaveVideoNoteParams {
  url: string;
  title: string;
  duration: number;
  thumbnail: string;
}

class VideoService {
  async getVideoMetadata({ url, signal }: GetVideoMetadataParams): Promise<VideoMetadata> {
    try {
      const response = await apiClient.get<VideoMetadata>('/api/video/metadata', {
        params: { url },
        signal,
      });
      return response.data;
    } catch (error: any) {
      if (error.name === 'AbortError' || error.code === 'ERR_CANCELED') {
        throw new Error('REQUEST_CANCELED');
      }
      
      if (error.response?.status === 400) {
        throw new Error(error.response.data?.message || '无效的 YouTube 链接');
      }
      
      if (error.response?.status === 500) {
        throw new Error('服务异常，请稍后重试');
      }
      
      if (error.code === 'ECONNABORTED' || error.message.includes('timeout')) {
        throw new Error('请求超时，请重试');
      }
      
      if (!navigator.onLine) {
        throw new Error('网络连接失败');
      }
      
      throw new Error(error.response?.data?.message || '获取视频信息失败');
    }
  }

  async saveVideoNote(params: SaveVideoNoteParams): Promise<VideoNote> {
    try {
      const response = await apiClient.post<VideoNote>('/api/video/note', params);
      return response.data;
    } catch (error: any) {
      if (error.response?.status === 400) {
        throw new Error(error.response.data?.message || '保存失败，请检查输入');
      }
      
      if (error.response?.status === 500) {
        throw new Error('服务异常，请稍后重试');
      }
      
      throw new Error(error.response?.data?.message || '保存笔记失败');
    }
  }
}

export const videoService = new VideoService();