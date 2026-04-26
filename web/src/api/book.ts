import http from './http';
import type { ListResponse } from '@/types/admin';

const basePath = '/books'

export function listbooks(params: Record<string, unknown> = {}): Promise<ListResponse<any>> {
  return http.get<ListResponse<any>>(basePath, { params });
}

export function getBook(id: string | number): Promise<any> {
  return http.get<any>(basePath + '/' + id);
}

export function createBook(data: Record<string, unknown>): Promise<any> {
  return http.post<any>(basePath, data);
}

export function updateBook(id: string | number, data: Record<string, unknown>): Promise<any> {
  return http.put<any>(basePath + '/' + id, data);
}

export function deleteBook(id: string | number): Promise<any> {
  return http.delete<any>(basePath + '/' + id);
}
