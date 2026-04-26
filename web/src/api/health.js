import http from './http';
export function fetchHealth() {
    return http.get('/health');
}
export function fetchPublicConfig() {
    return http.get('/meta/config');
}
