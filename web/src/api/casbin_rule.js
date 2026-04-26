import http from './http';
const basePath = '/casbin_rules';
export function listcasbin_rules(params = {}) {
    return http.get(basePath, { params });
}
export function getCasbinRule(id) {
    return http.get(basePath + '/' + id);
}
export function createCasbinRule(data) {
    return http.post(basePath, data);
}
export function updateCasbinRule(id, data) {
    return http.put(basePath + '/' + id, data);
}
export function deleteCasbinRule(id) {
    return http.delete(basePath + '/' + id);
}
