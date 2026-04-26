import axios, { AxiosHeaders, } from 'axios';
import { getStoredAccessToken } from '@/store/session';
import { ApiError } from './types';
let unauthorizedHandler = null;
export function setUnauthorizedHandler(handler) {
    unauthorizedHandler = handler;
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
                return Promise.reject(new ApiError(payload.msg || 'Request failed', payload.code, payload.data, payload.request_id));
            }
            return payload.data;
        }
        return payload;
    }), ((error) => {
        if (error instanceof ApiError) {
            if (error.code === 401) {
                unauthorizedHandler?.(error);
            }
            return Promise.reject(error);
        }
        if (axios.isAxiosError(error)) {
            const axiosError = error;
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
    }));
    return client;
}
const rawClient = createHttpClient();
const http = rawClient;
export default http;
