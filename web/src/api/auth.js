import http from './http';
export function login(payload) {
    return http.post('/auth/login', payload);
}
export function fetchCurrentUser() {
    return http.get('/auth/me');
}
export function logout() {
    return http.post('/auth/logout');
}
