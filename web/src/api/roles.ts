import http from './http';

import type { ListResponse, RoleFormState, RoleItem, RoleQuery } from '@/types/admin';

export function fetchRoles(params: RoleQuery): Promise<ListResponse<RoleItem>> {
  return http.get<ListResponse<RoleItem>>('/roles', { params });
}

export function fetchRole(id: string): Promise<RoleItem> {
  return http.get<RoleItem>(`/roles/${id}`);
}

export function createRole(payload: RoleFormState): Promise<RoleItem> {
  return http.post<RoleItem>('/roles', payload);
}

export function updateRole(id: string, payload: RoleFormState): Promise<RoleItem> {
  return http.put<RoleItem>(`/roles/${id}`, payload);
}

export function deleteRole(id: string): Promise<{ deleted: boolean }> {
  return http.delete<{ deleted: boolean }>(`/roles/${id}`);
}
