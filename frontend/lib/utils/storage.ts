/**
 * 本地存储工具函数
 */

import { STORAGE_KEYS } from "@/lib/constants";

/**
 * 存储类型
 */
type StorageType = "localStorage" | "sessionStorage";

/**
 * 获取存储对象
 */
function getStorage(type: StorageType = "localStorage"): Storage {
  if (typeof window === "undefined") {
    throw new Error("Storage is only available in browser");
  }
  return type === "localStorage" ? localStorage : sessionStorage;
}

/**
 * 设置存储项
 */
export function setStorageItem<T>(
  key: string,
  value: T,
  type: StorageType = "localStorage"
): void {
  try {
    const storage = getStorage(type);
    const serialized = JSON.stringify(value);
    storage.setItem(key, serialized);
  } catch (error) {
    console.error(`Failed to set storage item: ${key}`, error);
  }
}

/**
 * 获取存储项
 */
export function getStorageItem<T>(
  key: string,
  defaultValue: T | null = null,
  type: StorageType = "localStorage"
): T | null {
  try {
    const storage = getStorage(type);
    const item = storage.getItem(key);
    if (item === null) return defaultValue;
    return JSON.parse(item) as T;
  } catch (error) {
    console.error(`Failed to get storage item: ${key}`, error);
    return defaultValue;
  }
}

/**
 * 删除存储项
 */
export function removeStorageItem(
  key: string,
  type: StorageType = "localStorage"
): void {
  try {
    const storage = getStorage(type);
    storage.removeItem(key);
  } catch (error) {
    console.error(`Failed to remove storage item: ${key}`, error);
  }
}

/**
 * 清空存储
 */
export function clearStorage(type: StorageType = "localStorage"): void {
  try {
    const storage = getStorage(type);
    storage.clear();
  } catch (error) {
    console.error(`Failed to clear storage`, error);
  }
}

/**
 * 设置认证 token
 */
export function setAuthToken(token: string): void {
  setStorageItem(STORAGE_KEYS.AUTH_TOKEN, token);
}

/**
 * 获取认证 token
 */
export function getAuthToken(): string | null {
  return getStorageItem<string>(STORAGE_KEYS.AUTH_TOKEN);
}

/**
 * 删除认证 token
 */
export function removeAuthToken(): void {
  removeStorageItem(STORAGE_KEYS.AUTH_TOKEN);
}

/**
 * 设置用户信息
 */
export function setUserInfo<T>(userInfo: T): void {
  setStorageItem(STORAGE_KEYS.USER_INFO, userInfo);
}

/**
 * 获取用户信息
 */
export function getUserInfo<T>(): T | null {
  return getStorageItem<T>(STORAGE_KEYS.USER_INFO);
}

/**
 * 删除用户信息
 */
export function removeUserInfo(): void {
  removeStorageItem(STORAGE_KEYS.USER_INFO);
}

/**
 * Google OAuth Token 类型
 */
export interface GoogleOAuthToken {
  accessToken: string;
  refreshToken: string;
  tokenType: string;
  expiry: string;
}

/**
 * 设置 Google OAuth token
 */
export function setGoogleOAuthToken(token: GoogleOAuthToken): void {
  setStorageItem('google_oauth_token', token);
}

/**
 * 获取 Google OAuth token
 */
export function getGoogleOAuthToken(): GoogleOAuthToken | null {
  return getStorageItem<GoogleOAuthToken>('google_oauth_token');
}

/**
 * 删除 Google OAuth token
 */
export function removeGoogleOAuthToken(): void {
  removeStorageItem('google_oauth_token');
  removeStorageItem('google_access_token');
  removeStorageItem('google_token_expiry');
}

/**
 * 检查 Google OAuth token 是否已过期
 */
export function isGoogleTokenExpired(): boolean {
  if (typeof window === "undefined") return true;
  const expiry = localStorage.getItem('google_token_expiry');
  if (!expiry) return true;
  return new Date(expiry) < new Date();
}

/**
 * 检查用户是否已授权 Google
 */
export function isGoogleAuthorized(): boolean {
  const token = getGoogleOAuthToken();
  return token !== null && !isGoogleTokenExpired();
}

