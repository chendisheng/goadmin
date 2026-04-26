import http from './http';
export function fetchUsers(params) {
    return http.get('/users', { params });
}
export function fetchUser(id) {
    return http.get(`/users/${id}`);
}
export function createUser(payload) {
    return http.post('/users', payload);
}
export function updateUser(id, payload) {
    return http.put(`/users/${id}`, payload);
}
export function deleteUser(id) {
    return http.delete(`/users/${id}`);
}
