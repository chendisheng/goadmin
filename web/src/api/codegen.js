import { getStoredAccessToken } from '@/store/session';
import { ApiError } from './types';
import http from './http';
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';
export function previewCodegenDsl(payload) {
    return http.post('/codegen/dsl/preview', payload);
}
export function generateCodegenDsl(payload) {
    return http.post('/codegen/dsl/generate', payload);
}
export function generateDownloadCodegenDsl(payload) {
    return http.post('/codegen/dsl/generate-download', payload);
}
export function previewCodegenDatabase(payload) {
    return http.post('/codegen/db/preview', payload);
}
export function generateCodegenDatabase(payload) {
    return http.post('/codegen/db/generate', payload);
}
export function generateDownloadCodegenDatabase(payload) {
    return http.post('/codegen/db/generate-download', payload);
}
export function previewCodegenDelete(payload) {
    return http.post('/codegen/delete/preview', payload);
}
export function executeCodegenDelete(payload) {
    return http.post('/codegen/delete/execute', payload);
}
export function installCodegenManifest(payload) {
    return http.post('/codegen/install/manifest', payload);
}
export async function downloadCodegenArtifact(downloadUrl, fallbackFilename) {
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
function resolveApiUrl(path) {
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
function isApiEnvelope(value) {
    return typeof value === 'object' && value !== null && 'code' in value && 'msg' in value;
}
async function toDownloadError(response) {
    const contentType = response.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
        const payload = (await response.json());
        if (isApiEnvelope(payload)) {
            return new ApiError(payload.msg || 'Download failed', payload.code, payload.data, payload.request_id);
        }
    }
    const text = await response.text();
    return new ApiError(text || 'Download failed', response.status);
}
function extractFilename(contentDisposition, fallback) {
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
