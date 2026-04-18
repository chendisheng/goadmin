import http from './http';

const basePath = '/casbin_models'

export function listcasbin_models(params = {}) {
  return http.get(basePath, { params });
}

export function getCasbinModel(id) {
  return http.get(basePath + '/' + id);
}

export function createCasbinModel(data) {
  return http.post(basePath, data);
}

export function updateCasbinModel(id, data) {
  return http.put(basePath + '/' + id, data);
}

export function deleteCasbinModel(id) {
  return http.delete(basePath + '/' + id);
}
