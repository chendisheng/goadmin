import http from './http';
const basePath = '/books';
export function listbooks(params = {}) {
    return http.get(basePath, { params });
}
export function getBook(id) {
    return http.get(basePath + '/' + id);
}
export function createBook(data) {
    return http.post(basePath, data);
}
export function updateBook(id, data) {
    return http.put(basePath + '/' + id, data);
}
export function deleteBook(id) {
    return http.delete(basePath + '/' + id);
}
