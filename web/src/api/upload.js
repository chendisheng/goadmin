import { getStoredAccessToken } from '@/store/session';
import http from './http';
import { ApiError } from './types';
const basePath = '/uploads/files';
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';
const storageSettingPath = `${basePath}/storage/default`;
export function fetchUploadFiles(params) {
    return http.get(basePath, { params });
}
export function fetchUploadFile(id) {
    return http.get(`${basePath}/${id}`);
}
export function previewUploadFile(id) {
    return http.get(`${basePath}/${id}/preview`);
}
export function deleteUploadFile(id) {
    return http.delete(`${basePath}/${id}`);
}
export function bindUploadFile(id, payload) {
    return http.post(`${basePath}/${id}/bind`, payload);
}
export function unbindUploadFile(id) {
    return http.delete(`${basePath}/${id}/bind`);
}
export function fetchUploadStorageSetting() {
    return http.get(storageSettingPath);
}
export function updateUploadStorageSetting(payload) {
    return http.put(storageSettingPath, payload);
}
export async function uploadUploadFile(file, payload) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('visibility', payload.visibility);
    formData.append('biz_module', payload.biz_module);
    formData.append('biz_type', payload.biz_type);
    formData.append('biz_id', payload.biz_id);
    formData.append('biz_field', payload.biz_field);
    formData.append('remark', payload.remark);
    const response = await fetch(resolveApiUrl(basePath), {
        method: 'POST',
        headers: authHeaders(),
        body: formData,
    });
    const data = await parseJsonResponse(response, 'Upload failed');
    return data;
}
export async function downloadUploadFile(id, fallbackFilename) {
    const response = await fetch(resolveApiUrl(`${basePath}/${id}/download`), {
        method: 'GET',
        headers: authHeaders(),
    });
    if (!response.ok) {
        throw await toDownloadError(response);
    }
    const blob = await response.blob();
    const filename = extractFilename(response.headers.get('content-disposition'), fallbackFilename || 'upload-file');
    const objectUrl = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = objectUrl;
    link.download = filename;
    link.style.display = 'none';
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(objectUrl);
}
export async function createUploadFilePreviewUrl(id) {
    const response = await fetch(resolveApiUrl(`${basePath}/${id}/download`), {
        method: 'GET',
        headers: authHeaders(),
    });
    if (!response.ok) {
        throw await toDownloadError(response);
    }
    const blob = await response.blob();
    return window.URL.createObjectURL(blob);
}
function authHeaders() {
    const token = getStoredAccessToken();
    const headers = {
        'X-Requested-With': 'XMLHttpRequest',
    };
    if (token) {
        headers.Authorization = `Bearer ${token}`;
    }
    return headers;
}
function resolveApiUrl(path) {
    if (/^https?:\/\//i.test(path)) {
        return path;
    }
    const prefix = API_BASE_URL.endsWith('/') ? API_BASE_URL.slice(0, -1) : API_BASE_URL;
    const suffix = path.startsWith('/') ? path : `/${path}`;
    return `${prefix}${suffix}`;
}
async function parseJsonResponse(response, fallbackMessage) {
    const payload = await response.json().catch(() => null);
    if (!response.ok) {
        if (isApiEnvelope(payload)) {
            throw new ApiError(payload.msg || fallbackMessage, payload.code, payload.data, payload.request_id);
        }
        throw new ApiError(fallbackMessage, response.status, payload);
    }
    if (isApiEnvelope(payload)) {
        if (payload.code !== 200) {
            throw new ApiError(payload.msg || fallbackMessage, payload.code, payload.data, payload.request_id);
        }
        return payload.data;
    }
    return payload;
}
function isApiEnvelope(value) {
    return typeof value === 'object' && value !== null && 'code' in value && 'msg' in value;
}
async function toDownloadError(response) {
    const contentType = response.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
        const payload = (await response.json().catch(() => null));
        if (isApiEnvelope(payload)) {
            return new ApiError(payload.msg || 'Download failed', payload.code, payload.data, payload.request_id);
        }
    }
    const text = await response.text().catch(() => '');
    return new ApiError(text || 'Download failed', response.status);
}
function extractFilename(contentDisposition, fallback) {
    if (!contentDisposition) {
        return fallback;
    }
    const utf8Match = /filename\*=UTF-8''([^;]+)/i.exec(contentDisposition);
    if (utf8Match?.[1]) {
        try {
            return decodeURIComponent(utf8Match[1]);
        }
        catch {
            return utf8Match[1];
        }
    }
    const filenameMatch = /filename=\"?([^\";]+)\"?/i.exec(contentDisposition);
    if (filenameMatch?.[1]) {
        return filenameMatch[1];
    }
    return fallback;
}
