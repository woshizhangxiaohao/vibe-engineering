/**
 * API 类型定义
 */

/**
 * API 响应基础结构
 */
export interface ApiResponse<T = unknown> {
  data?: T;
  error?: string;
  message?: string;
}

/**
 * 分页响应结构
 */
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

/**
 * 请求配置
 */
export interface RequestConfig extends RequestInit {
  timeout?: number;
  skipAuth?: boolean;
}

/**
 * HTTP 错误类
 */
export class ApiError extends Error {
  constructor(
    public status: number,
    public statusText: string,
    public data?: unknown,
    message?: string
  ) {
    super(message || statusText);
    this.name = "ApiError";
  }
}

/**
 * 请求选项
 */
export interface RequestOptions {
  method?: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
  body?: unknown;
  params?: Record<string, string | number | boolean | null | undefined>;
  headers?: Record<string, string>;
  timeout?: number;
  signal?: AbortSignal;
}

/**
 * Insight 相关类型定义
 */

export type InsightSourceType = "youtube" | "twitter" | "podcast";
export type InsightStatus = "pending" | "processing" | "completed" | "failed";

export interface Insight {
  id: number;
  source_type: InsightSourceType;
  source_url: string;
  title: string;
  author: string;
  thumbnail_url: string;
  status: InsightStatus;
  target_lang: string;
  created_at: string;
  updated_at: string;
}

export interface GroupedInsights {
  today: Insight[];
  yesterday: Insight[];
  previous: Insight[];
}

export interface InsightsListResponse {
  data: GroupedInsights;
  total: number;
}

export interface CreateInsightRequest {
  source_url: string;
  target_lang?: string;
}

export interface CreateInsightResponse {
  id: number;
  status: InsightStatus;
  message: string;
}

/**
 * Transcript item with timestamp
 */
export interface TranscriptItem {
  timestamp: string; // e.g., "05:12"
  seconds: number; // time in seconds
  text: string; // original transcript text
  translated_text?: string; // translated text (optional)
}

/**
 * Highlight annotation
 */
export interface Highlight {
  id: number;
  insight_id: number;
  text: string;
  start_offset: number;
  end_offset: number;
  color: string;
  note?: string;
  created_at: string;
}

/**
 * Create highlight request
 */
export interface CreateHighlightRequest {
  text: string;
  start_offset: number;
  end_offset: number;
  color: string;
  note?: string;
}

/**
 * Highlight color options
 */
export type HighlightColor = "yellow" | "green" | "blue" | "purple" | "red";

/**
 * Highlights list response
 */
export interface HighlightsListResponse {
  highlights: Highlight[];
}

/**
 * Insight detail response from GET /api/v1/insights/:id
 */
export interface InsightDetailResponse {
  id: number;
  source_type: InsightSourceType;
  source_url: string;
  source_id: string;
  title: string;
  author: string;
  thumbnail_url: string;
  duration: number; // seconds
  published_at?: string;
  summary: string;
  key_points: string[];
  raw_content?: string;
  trans_content?: string;
  transcripts?: TranscriptItem[];
  status: InsightStatus;
  error_message?: string;
  highlights?: Highlight[];
  created_at: string;
}

/**
 * Chat 相关类型定义
 */

export type ChatMessageRole = "user" | "assistant";

/**
 * Chat message
 */
export interface ChatMessage {
  id: number;
  role: ChatMessageRole;
  content: string;
  created_at: string;
}

/**
 * Send chat message request
 */
export interface SendChatMessageRequest {
  message: string;
  highlight_id?: number;
}

/**
 * Chat history response
 */
export interface ChatHistoryResponse {
  messages: ChatMessage[];
}

/**
 * SSE streaming chunk
 */
export interface SSEChunk {
  role: ChatMessageRole;
  content: string;
  done: boolean;
  message_id?: number;
}

/**
 * Entity types
 */
export type EntityType = "stock" | "crypto";

/**
 * Detected entity
 */
export interface Entity {
  type: EntityType;
  name: string;
  ticker: string;
}

/**
 * Suggestion types
 */
export type SuggestionType = "position" | "prediction";

/**
 * AI suggestion
 */
export interface Suggestion {
  type: SuggestionType;
  entity: string;
  prompt: string;
}

/**
 * Analyze entities response
 */
export interface AnalyzeEntitiesResponse {
  entities: Entity[];
  suggestions: Suggestion[];
}

/**
 * 分享功能相关类型定义
 */

/**
 * 分享请求
 */
export interface ShareInsightRequest {
  include_summary: boolean;
  include_key_points: boolean;
  include_highlights: boolean;
  include_chat: boolean;
  is_public: boolean;
  password?: string;
}

/**
 * 分享响应
 */
export interface ShareInsightResponse {
  share_token: string;
  share_url: string;
  expires_at?: string;
}

/**
 * 公共分享内容
 */
export interface SharedContent {
  summary?: string;
  key_points?: string[];
  highlights?: Highlight[];
  chat?: ChatMessage[];
}

/**
 * 公共分享 Insight 响应
 */
export interface SharedInsightResponse {
  title: string;
  author: string;
  thumbnail_url: string;
  shared_by: string;
  shared_at?: string;
  source_type: InsightSourceType;
  source_url: string;
  content: SharedContent;
}

/**
 * Translation 相关类型定义
 */

/**
 * 翻译请求
 */
export interface TranslateRequest {
  source_text?: string;
  youtube_url?: string;
  source_language?: string;
  target_language: string;
  enable_dual_subtitles?: boolean;
}

/**
 * 双语字幕条目
 */
export interface DualSubtitle {
  original: string;
  translated: string;
  start_time?: string;
  end_time?: string;
}

/**
 * 翻译响应
 */
export interface TranslateResponse {
  status: string;
  message?: string;
  translated_text?: string;
  dual_subtitles?: DualSubtitle[];
  source_language?: string;
}

