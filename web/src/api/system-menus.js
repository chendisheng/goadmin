import http from './http';
export function fetchMenuTree(params) {
    return http.get('/menus/tree', { params });
}
export function fetchMenus(params) {
    return http.get('/menus', { params });
}
export function fetchMenu(id) {
    return http.get(`/menus/${id}`);
}
export function createMenu(payload) {
    return http.post('/menus', payload);
}
export function updateMenu(id, payload) {
    return http.put(`/menus/${id}`, payload);
}
export function deleteMenu(id) {
    return http.delete(`/menus/${id}`);
}
