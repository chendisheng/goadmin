import http from './http';
export function fetchRoles(params) {
    return http.get('/roles', { params });
}
export function fetchRole(id) {
    return http.get(`/roles/${id}`);
}
export function createRole(payload) {
    return http.post('/roles', payload);
}
export function updateRole(id, payload) {
    return http.put(`/roles/${id}`, payload);
}
export function deleteRole(id) {
    return http.delete(`/roles/${id}`);
}
