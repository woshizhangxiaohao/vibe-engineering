/**
 * Chat API 服务
 */

import { apiClient } from "../client";
import { API_BASE_PATH } from "../config";
import {
  ChatHistoryResponse,
  SendChatMessageRequest,
  AnalyzeEntitiesResponse,
  SSEChunk,
} from "../types";

/**
 * Chat 服务类
 */
class ChatService {
  /**
   * 获取对话历史
   */
  async getChatHistory(insightId: number): Promise<ChatHistoryResponse> {
    return apiClient.get<ChatHistoryResponse>(
      `/v1/insights/${insightId}/chat`
    );
  }

  /**
   * 发送聊天消息（SSE 流式响应）
   * @returns EventSource 实例，用于接收流式响应
   */
  createChatStream(
    insightId: number,
    data: SendChatMessageRequest
  ): EventSource {
    // 构建查询参数
    const params = new URLSearchParams({
      message: data.message,
    });
    if (data.highlight_id) {
      params.append("highlight_id", String(data.highlight_id));
    }

    // 创建 EventSource
    const url = `${API_BASE_PATH}/v1/insights/${insightId}/chat?${params.toString()}`;
    return new EventSource(url);
  }

  /**
   * 分析内容中的实体
   */
  async analyzeEntities(insightId: number): Promise<AnalyzeEntitiesResponse> {
    return apiClient.post<AnalyzeEntitiesResponse>(
      `/v1/insights/${insightId}/analyze-entities`
    );
  }
}

// 导出单例实例
export const chatService = new ChatService();

// 导出类以便需要时可以创建新实例
export default ChatService;
