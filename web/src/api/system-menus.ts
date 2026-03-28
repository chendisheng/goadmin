import http from './http';

import type { ListResponse, MenuFormState, MenuItem, MenuQuery } from '@/types/admin';

export function fetchMenuTree(params?: Partial<MenuQuery>): Promise<{ items: MenuItem[] }> {
  return http.get<{ items: MenuItem[] }>('/menus/tree', { params });
}

export function fetchMenus(params: MenuQuery): Promise<ListResponse<MenuItem>> {
  return http.get<ListResponse<MenuItem>>('/menus', { params });
}

export function fetchMenu(id: string): Promise<MenuItem> {
  return http.get<MenuItem>(`/menus/${id}`);
}

export function createMenu(payload: MenuFormState): Promise<MenuItem> {
  return http.post<MenuItem>('/menus', payload);
}

export function updateMenu(id: string, payload: MenuFormState): Promise<MenuItem> {
  return http.put<MenuItem>(`/menus/${id}`, payload);
}

export function deleteMenu(id: string): Promise<{ deleted: boolean }> {
  return http.delete<{ deleted: boolean }>(`/menus/${id}`);
}
