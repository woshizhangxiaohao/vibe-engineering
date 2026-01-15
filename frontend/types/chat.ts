/**
 * Chat 相关类型定义
 */

export interface ChatMessage {
  id: number;
  analysis_id: number;
  role: 'user' | 'assistant';
  content: string;
  highlight_id?: number;
  created_at: string;
  updated_at: string;
}

export interface ChatRequest {
  message: string;
  highlight_id?: number;
}

export interface ChatStreamEvent {
  role: string;
  content: string;
  done: boolean;
  message_id?: number;
}

export interface ChatHistoryResponse {
  messages: ChatMessage[];
}

export interface Entity {
  name: string;
  type: string;
  description: string;
  first_mention?: string;
}

export interface EntityAnalysisResult {
  entities: Entity[];
}
