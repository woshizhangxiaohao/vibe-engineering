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
  text: string; // transcript text
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

