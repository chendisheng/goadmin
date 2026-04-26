import http from './http';
const categoryBasePath = '/dictionaries/categories';
const itemBasePath = '/dictionaries/items';
const lookupBasePath = '/dictionaries/lookup';
export function fetchDictionaryCategories(params) {
    return http.get(categoryBasePath, { params });
}
export function fetchDictionaryCategory(id) {
    return http.get(`${categoryBasePath}/${id}`);
}
export function createDictionaryCategory(payload) {
    return http.post(categoryBasePath, payload);
}
export function updateDictionaryCategory(id, payload) {
    return http.put(`${categoryBasePath}/${id}`, payload);
}
export function deleteDictionaryCategory(id) {
    return http.delete(`${categoryBasePath}/${id}`);
}
export function fetchDictionaryItems(params) {
    return http.get(itemBasePath, { params });
}
export function fetchDictionaryItem(id) {
    return http.get(`${itemBasePath}/${id}`);
}
export function createDictionaryItem(payload) {
    return http.post(itemBasePath, payload);
}
export function updateDictionaryItem(id, payload) {
    return http.put(`${itemBasePath}/${id}`, payload);
}
export function deleteDictionaryItem(id) {
    return http.delete(`${itemBasePath}/${id}`);
}
export function fetchDictionaryLookupItems(categoryCode) {
    return http.get(`${lookupBasePath}/${categoryCode}`);
}
export function fetchDictionaryLookupItem(categoryCode, value) {
    return http.get(`${lookupBasePath}/${categoryCode}/${value}`);
}
