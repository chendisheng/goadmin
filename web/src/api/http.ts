import axios, {
  AxiosError,
  AxiosHeaders,
  type AxiosInstance,
  type AxiosRequestConfig,
  type AxiosResponse,
} from 'axios';

import { translate } from '@/i18n';
import { getStoredAccessToken } from '@/store/session';

import { ApiError, type ApiEnvelope } from './types';

type UnauthorizedHandler = (error: ApiError | AxiosError<unknown>) => void;

let unauthorizedHandler: UnauthorizedHandler | null = null;

export function setUnauthorizedHandler(handler: UnauthorizedHandler | null) {
  unauthorizedHandler = handler;
}

function resolveHttpErrorMessage(code: number, message = ''): string {
  const normalizedMessage = message.trim();
  if (code === 401) {
    return normalizedMessage || translate('common.authentication_required', '需要登录');
  }
  if (code === 403) {
    return normalizedMessage || translate('common.permission_denied', '无权访问');
  }
  if (normalizedMessage !== '') {
    return normalizedMessage;
  }
  return translate('common.request_failed', '请求失败');
}

function resolveNetworkErrorMessage(message = ''): string {
  const normalizedMessage = message.trim();
  if (normalizedMessage !== '') {
    return normalizedMessage;
  }
  return translate('common.network_error', '网络错误');
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
      const headers = AxiosHeaders.from(config.headers);
      headers.set('Authorization', `Bearer ${token}`);
      headers.set('X-Requested-With', 'XMLHttpRequest');
      config.headers = headers;
    }
    return config;
  });

  client.interceptors.response.use(
    ((response: AxiosResponse<unknown>) => {
      const payload = response.data;
      if (isApiEnvelope(payload)) {
        if (payload.code !== 200) {
          return Promise.reject(new ApiError(resolveHttpErrorMessage(payload.code, payload.msg), payload.code, payload.data, payload.request_id));
        }
        return payload.data;
      }
      return payload;
    }) as any,
    ((error: unknown) => {
      if (error instanceof ApiError) {
        if (error.code === 401) {
          unauthorizedHandler?.(error);
        }
        const normalizedMessage = resolveHttpErrorMessage(error.code, error.message);
        if (normalizedMessage !== error.message) {
          return Promise.reject(new ApiError(normalizedMessage, error.code, error.payload, error.requestId));
        }
        return Promise.reject(error);
      }
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError<unknown>;
        const data = axiosError.response?.data;
        if (axiosError.response?.status === 401) {
          unauthorizedHandler?.(axiosError);
        }
        if (axiosError.response?.status === 403) {
          return Promise.reject(new ApiError(resolveHttpErrorMessage(403, isApiEnvelope(data) ? data.msg : axiosError.message), 403, data, isApiEnvelope(data) ? data.request_id : undefined));
        }
        if (isApiEnvelope(data)) {
          return Promise.reject(new ApiError(resolveHttpErrorMessage(data.code, data.msg), data.code, data.data, data.request_id));
        }
        return Promise.reject(new ApiError(resolveNetworkErrorMessage(axiosError.message), axiosError.response?.status ?? 500, data));
      }
      return Promise.reject(error);
    }) as any,
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
