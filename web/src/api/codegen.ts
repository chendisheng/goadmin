import { getStoredAccessToken } from '@/store/session';

import { ApiError, type ApiEnvelope } from './types';
import http from './http';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

export interface CodegenDslRequest {
  dsl: string;
  force?: boolean;
}

export interface CodegenDslDownloadRequest extends CodegenDslRequest {
  package_name?: string;
  include_readme?: boolean;
  include_report?: boolean;
  include_dsl?: boolean;
}

export interface CodegenDslPreviewItem {
  index: number;
  kind: string;
  name: string;
  force: boolean;
  actions: string[];
}

export interface CodegenDslExecutionReport {
  dry_run: boolean;
  messages: string[];
  items: CodegenDslPreviewItem[];
}

export interface CodegenArtifactInfo {
  task_id: string;
  download_url: string;
  filename: string;
  size_bytes: number;
  file_count: number;
  expire_at: string;
}

export function previewCodegenDsl(payload: CodegenDslRequest) {
  return http.post<CodegenDslExecutionReport>('/codegen/dsl/preview', payload);
}

export function generateCodegenDsl(payload: CodegenDslRequest) {
  return http.post<CodegenDslExecutionReport>('/codegen/dsl/generate', payload);
}

export function generateDownloadCodegenDsl(payload: CodegenDslDownloadRequest) {
  return http.post<CodegenArtifactInfo>('/codegen/dsl/generate-download', payload);
}

export async function downloadCodegenArtifact(downloadUrl: string, fallbackFilename?: string) {
  const token = getStoredAccessToken();
  const response = await fetch(resolveApiUrl(downloadUrl), {
    method: 'GET',
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      'X-Requested-With': 'XMLHttpRequest',
    },
  });
  if (!response.ok) {
    throw await toDownloadError(response);
  }
  const blob = await response.blob();
  const filename = extractFilename(response.headers.get('content-disposition'), fallbackFilename || 'codegen-package.zip');
  const objectUrl = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = objectUrl;
  link.download = filename;
  link.style.display = 'none';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.setTimeout(() => window.URL.revokeObjectURL(objectUrl), 0);
}

function resolveApiUrl(path: string): string {
  const value = path.trim();
  if (!value) {
    return API_BASE_URL;
  }
  if (/^https?:\/\//i.test(value)) {
    return value;
  }
  if (/^https?:\/\//i.test(API_BASE_URL)) {
    const base = new URL(API_BASE_URL);
    return new URL(value, `${base.protocol}//${base.host}`).toString();
  }
  return value;
}

function isApiEnvelope<T = unknown>(value: unknown): value is ApiEnvelope<T> {
  return typeof value === 'object' && value !== null && 'code' in value && 'msg' in value;
}

async function toDownloadError(response: Response): Promise<ApiError> {
  const contentType = response.headers.get('content-type') || '';
  if (contentType.includes('application/json')) {
    const payload = (await response.json()) as unknown;
    if (isApiEnvelope(payload)) {
      return new ApiError(payload.msg || 'Download failed', payload.code, payload.data, payload.request_id);
    }
  }
  const text = await response.text();
  return new ApiError(text || 'Download failed', response.status);
}

function extractFilename(contentDisposition: string | null, fallback: string): string {
  if (!contentDisposition) {
    return fallback;
  }
  const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8Match?.[1]) {
    return decodeURIComponent(utf8Match[1]);
  }
  const simpleMatch = contentDisposition.match(/filename="?([^";]+)"?/i);
  if (simpleMatch?.[1]) {
    return simpleMatch[1];
  }
  return fallback;
}
