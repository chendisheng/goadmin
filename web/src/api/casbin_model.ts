import http from './http';

const basePath = '/casbin_models'

export function listcasbin_models(params: Record<string, unknown> = {}): Promise<{ items?: any[]; total?: number }> {
  return http.get(basePath, { params });
}

export function getCasbinModel(id: string | number): Promise<any> {
  return http.get(basePath + '/' + id);
}

export function createCasbinModel(data: Record<string, unknown>): Promise<any> {
  return http.post(basePath, data);
}

export function updateCasbinModel(id: string | number, data: Record<string, unknown>): Promise<any> {
  return http.put(basePath + '/' + id, data);
}

export function deleteCasbinModel(id: string | number): Promise<any> {
  return http.delete(basePath + '/' + id);
}
