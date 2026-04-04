import http from './http';

const basePath = '/api/v1/codegen_consoles';

export function listcodegen_consoles(params: Record<string, any> = {}) {
  return http.get(basePath, { params });
}

export function getCodegenConsole(id: string) {
  return http.get(basePath + '/' + id);
}

export function createCodegenConsole(data: Record<string, any>) {
  return http.post(basePath, data);
}

export function updateCodegenConsole(id: string, data: Record<string, any>) {
  return http.put(basePath + '/' + id, data);
}

export function deleteCodegenConsole(id: string) {
  return http.delete(basePath + '/' + id);
}
