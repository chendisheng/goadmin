import axios, { AxiosError, type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios';

import { getStoredAccessToken } from '@/store/session';

import { ApiError, type ApiEnvelope } from './types';

type UnauthorizedHandler = (error: ApiError | AxiosError<unknown>) => void;

let unauthorizedHandler: UnauthorizedHandler | null = null;

export function setUnauthorizedHandler(handler: UnauthorizedHandler | null) {
  unauthorizedHandler = handler;
}

function isApiEnvelope<T = unknown>(value: unknown): value is ApiEnvelope<T> {
  return typeof value === 'object' && value !== null && 'code' in value && 'msg' in value;
}

function createHttpClient(): AxiosInstance {
  const client = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
    timeout: 15000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  client.interceptors.request.use((config) => {
    const token = getStoredAccessToken();
    if (token) {
      const headers = (config.headers ?? {}) as Record<string, string>;
      config.headers = {
        ...headers,
        Authorization: `Bearer ${token}`,
        'X-Requested-With': 'XMLHttpRequest',
      } as typeof config.headers;
    }
    return config;
  });

  client.interceptors.response.use(
    (response: AxiosResponse<unknown>) => {
      const payload = response.data;
      if (isApiEnvelope(payload)) {
        if (payload.code !== 200) {
          return Promise.reject(new ApiError(payload.msg || 'Request failed', payload.code, payload.data, payload.request_id));
        }
        return payload.data;
      }
      return payload;
    },
    (error: unknown) => {
      if (error instanceof ApiError) {
        if (error.code === 401) {
          unauthorizedHandler?.(error);
        }
        return Promise.reject(error);
      }
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError<unknown>;
        const data = axiosError.response?.data;
        if (axiosError.response?.status === 401) {
          unauthorizedHandler?.(axiosError);
        }
        if (isApiEnvelope(data)) {
          return Promise.reject(new ApiError(data.msg || axiosError.message, data.code, data.data, data.request_id));
        }
        return Promise.reject(new ApiError(axiosError.message || 'Network error', axiosError.response?.status ?? 500, data));
      }
      return Promise.reject(error);
    },
  );

  return client;
}

export interface HttpClient {
  request<T = unknown>(config: AxiosRequestConfig): Promise<T>;
  get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T>;
  post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>;
  put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>;
  delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T>;
}

const rawClient = createHttpClient();
const http = rawClient as unknown as HttpClient;

export default http;
