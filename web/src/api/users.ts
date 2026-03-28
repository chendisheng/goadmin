import http from './http';

import type { ListResponse, UserFormState, UserItem, UserQuery } from '@/types/admin';

export function fetchUsers(params: UserQuery): Promise<ListResponse<UserItem>> {
  return http.get<ListResponse<UserItem>>('/users', { params });
}

export function fetchUser(id: string): Promise<UserItem> {
  return http.get<UserItem>(`/users/${id}`);
}

export function createUser(payload: UserFormState): Promise<UserItem> {
  return http.post<UserItem>('/users', payload);
}

export function updateUser(id: string, payload: UserFormState): Promise<UserItem> {
  return http.put<UserItem>(`/users/${id}`, payload);
}

export function deleteUser(id: string): Promise<{ deleted: boolean }> {
  return http.delete<{ deleted: boolean }>(`/users/${id}`);
}
