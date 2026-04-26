import http from './http';
export function fetchPlugins() {
    return http.get('/plugins');
}
export function fetchPlugin(name) {
    return http.get(`/plugins/${name}`);
}
export function createPlugin(payload) {
    return http.post('/plugins', payload);
}
export function updatePlugin(name, payload) {
    return http.put(`/plugins/${name}`, payload);
}
export function deletePlugin(name) {
    return http.delete(`/plugins/${name}`);
}
export function fetchPluginMenus() {
    return http.get('/plugins/menus');
}
export function fetchPluginPermissions() {
    return http.get('/plugins/permissions');
}
export function pingExamplePlugin() {
    return http.get('/plugins/example/ping');
}
