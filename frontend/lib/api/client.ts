/**
 * HTTP 客户端封装
 * 提供统一的请求方法、拦截器和错误处理
 */

import { API_BASE_PATH, API_TIMEOUT, DEFAULT_HEADERS } from "./config";
import { ApiError, RequestOptions } from "./types";
import { getAuthToken } from "@/lib/utils/storage";

/**
 * 构建查询字符串
 */
function buildQueryString(
  params: Record<string, string | number | boolean | null | undefined>
): string {
  const searchParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== null && value !== undefined) {
      searchParams.append(key, String(value));
    }
  });
  const queryString = searchParams.toString();
  return queryString ? `?${queryString}` : "";
}

/**
 * 检查 Google OAuth token 是否过期
 */
function isGoogleTokenExpired(): boolean {
  const expiry = localStorage.getItem("google_token_expiry");
  if (!expiry) return false;

  try {
    const expiryDate = new Date(expiry);
    const now = new Date();
    // 提前 5 分钟刷新 token，避免在请求过程中过期
    const bufferTime = 5 * 60 * 1000; // 5 minutes in milliseconds
    return expiryDate.getTime() - now.getTime() < bufferTime;
  } catch {
    return false;
  }
}

/**
 * 刷新 Google OAuth token
 * @returns 是否成功刷新 token，如果没有 refresh token 则返回 false
 */
async function refreshGoogleToken(): Promise<boolean> {
  const refreshToken = localStorage.getItem("google_refresh_token");
  if (!refreshToken) {
    // 没有 refresh token，静默返回 false，不抛出错误
    return false;
  }

  try {
    const response = await fetch(`${API_BASE_PATH}/v1/auth/google/refresh`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ refreshToken }),
    });

    if (!response.ok) {
      throw new Error("Failed to refresh token");
    }

    const data = await response.json();

    // 更新存储的 token
    localStorage.setItem("google_oauth_token", data.tokenJSON);
    localStorage.setItem("google_access_token", data.accessToken);
    localStorage.setItem("google_refresh_token", data.refreshToken);
    localStorage.setItem("google_token_expiry", data.expiry);
    return true;
  } catch (error) {
    console.error("Token refresh failed:", error);
    // 清除过期的 token
    localStorage.removeItem("google_oauth_token");
    localStorage.removeItem("google_access_token");
    localStorage.removeItem("google_refresh_token");
    localStorage.removeItem("google_token_expiry");
    return false;
  }
}

/**
 * 构建请求头
 */
function buildHeaders(customHeaders?: Record<string, string>): HeadersInit {
  const headers = new Headers({
    ...DEFAULT_HEADERS,
    ...customHeaders,
  });

  // 优先使用 Google OAuth token，如果不存在则使用通用 auth token
  const googleAccessToken = localStorage.getItem("google_access_token");
  const token = googleAccessToken || getAuthToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  return headers;
}

/**
 * 处理响应
 */
async function handleResponse<T>(response: Response): Promise<T> {
  const contentType = response.headers.get("content-type");
  const isJson = contentType?.includes("application/json");

  let data: unknown;
  if (isJson) {
    data = await response.json();
  } else {
    data = await response.text();
  }

  if (!response.ok) {
    throw new ApiError(
      response.status,
      response.statusText,
      data,
      isJson && typeof data === "object" && data !== null && "error" in data
        ? String(data.error)
        : response.statusText
    );
  }

  return data as T;
}

/**
 * 创建带超时的 fetch
 */
function fetchWithTimeout(
  url: string,
  options: RequestInit,
  timeout: number
): Promise<Response> {
  return new Promise((resolve, reject) => {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => {
      controller.abort();
      reject(new Error("Request timeout"));
    }, timeout);

    fetch(url, {
      ...options,
      signal: controller.signal,
    })
      .then((response) => {
        clearTimeout(timeoutId);
        resolve(response);
      })
      .catch((error) => {
        clearTimeout(timeoutId);
        reject(error);
      });
  });
}

/**
 * HTTP 客户端类
 */
class ApiClient {
  /**
   * 基础请求方法
   */
  async request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    const {
      method = "GET",
      body,
      params,
      headers: customHeaders,
      timeout = API_TIMEOUT,
      signal,
    } = options;

    // 检查并刷新 Google OAuth token（如果需要且不是刷新 token 请求本身）
    if (
      endpoint !== "/v1/auth/google/refresh" &&
      localStorage.getItem("google_access_token") &&
      isGoogleTokenExpired()
    ) {
      // 尝试刷新 token，如果失败则清除过期的 token
      const refreshed = await refreshGoogleToken();
      if (!refreshed) {
        // 刷新失败，清除 Google token，后续请求将使用其他认证方式（如果有）
        console.warn("Token refresh failed, cleared Google OAuth tokens");
      }
    }

    // 构建 URL
    const url = `${API_BASE_PATH}${endpoint}${
      params ? buildQueryString(params) : ""
    }`;

    // Debug logging in development
    if (process.env.NODE_ENV === "development") {
      console.log(`[API Client] ${method} ${url}`, body ? { body } : "");
    }

    // 构建请求配置
    const requestInit: RequestInit = {
      method,
      headers: buildHeaders(customHeaders),
      signal,
    };

    // 添加请求体
    if (body !== undefined) {
      requestInit.body = JSON.stringify(body);
    }

    try {
      // 发送请求
      const response = await fetchWithTimeout(url, requestInit, timeout);
      return await handleResponse<T>(response);
    } catch (error) {
      // 处理错误
      if (error instanceof ApiError) {
        throw error;
      }
      if (error instanceof Error && error.name === "AbortError") {
        throw new ApiError(0, "Request aborted", undefined, "请求已取消");
      }
      throw new ApiError(
        0,
        "Network Error",
        undefined,
        error instanceof Error ? error.message : "网络请求失败"
      );
    }
  }

  /**
   * GET 请求
   */
  get<T>(endpoint: string, options?: Omit<RequestOptions, "method" | "body">) {
    return this.request<T>(endpoint, { ...options, method: "GET" });
  }

  /**
   * POST 请求
   */
  post<T>(
    endpoint: string,
    body?: unknown,
    options?: Omit<RequestOptions, "method">
  ) {
    return this.request<T>(endpoint, { ...options, method: "POST", body });
  }

  /**
   * PUT 请求
   */
  put<T>(
    endpoint: string,
    body?: unknown,
    options?: Omit<RequestOptions, "method">
  ) {
    return this.request<T>(endpoint, { ...options, method: "PUT", body });
  }

  /**
   * PATCH 请求
   */
  patch<T>(
    endpoint: string,
    body?: unknown,
    options?: Omit<RequestOptions, "method">
  ) {
    return this.request<T>(endpoint, { ...options, method: "PATCH", body });
  }

  /**
   * DELETE 请求
   */
  delete<T>(
    endpoint: string,
    options?: Omit<RequestOptions, "method" | "body">
  ) {
    return this.request<T>(endpoint, { ...options, method: "DELETE" });
  }
}

// 导出单例实例
export const apiClient = new ApiClient();

// 导出类以便需要时可以创建新实例
export default ApiClient;
