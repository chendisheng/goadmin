import axios, { AxiosHeaders, } from 'axios';
import { useAppI18n } from '@/i18n';
import { getStoredAccessToken } from '@/store/session';
import { ApiError } from './types';
let unauthorizedHandler = null;
export function setUnauthorizedHandler(handler) {
    unauthorizedHandler = handler;
}
function resolveHttpErrorMessage(code, message = '') {
    const { t } = useAppI18n();
    const normalizedMessage = message.trim();
    if (code === 401) {
        return normalizedMessage || t('common.authentication_required', 'Authentication required');
    }
    if (code === 403) {
        return normalizedMessage || t('common.permission_denied', 'Access denied');
    }
    if (normalizedMessage !== '') {
        return normalizedMessage;
    }
    return t('common.request_failed', 'Request failed');
}
function resolveNetworkErrorMessage(message = '') {
    const { t } = useAppI18n();
    const normalizedMessage = message.trim();
    if (normalizedMessage !== '') {
        return normalizedMessage;
    }
    return t('common.network_error', 'Network error');
}
function isApiEnvelope(value) {
    return typeof value === 'object' && value !== null && 'code' in value && 'msg' in value;
}
function createHttpClient() {
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
    client.interceptors.response.use(((response) => {
        const payload = response.data;
        if (isApiEnvelope(payload)) {
            if (payload.code !== 200) {
                return Promise.reject(new ApiError(resolveHttpErrorMessage(payload.code, payload.msg), payload.code, payload.data, payload.request_id));
            }
            return payload.data;
        }
        return payload;
    }), ((error) => {
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
            const axiosError = error;
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
    }));
    return client;
}
const rawClient = createHttpClient();
const http = rawClient;
export default http;
