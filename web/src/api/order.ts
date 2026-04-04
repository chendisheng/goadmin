import http from './http';

const basePath = '/api/v1/orders'

export function listorders(params = {}) {
  return http.get(basePath, { params });
}

export function getOrder(id) {
  return http.get(basePath + '/' + id);
}

export function createOrder(data) {
  return http.post(basePath, data);
}

export function updateOrder(id, data) {
  return http.put(basePath + '/' + id, data);
}

export function deleteOrder(id) {
  return http.delete(basePath + '/' + id);
}
