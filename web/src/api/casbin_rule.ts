import http from './http';

const basePath = '/casbin_rules'

export function listcasbin_rules(params: Record<string, unknown> = {}): Promise<{ items?: any[]; total?: number }> {
  return http.get(basePath, { params });
}

export function getCasbinRule(id: string | number): Promise<any> {
  return http.get(basePath + '/' + id);
}

export function createCasbinRule(data: Record<string, unknown>): Promise<any> {
  return http.post(basePath, data);
}

export function updateCasbinRule(id: string | number, data: Record<string, unknown>): Promise<any> {
  return http.put(basePath + '/' + id, data);
}

export function deleteCasbinRule(id: string | number): Promise<any> {
  return http.delete(basePath + '/' + id);
}
