import http from './http';

import type { ListResponse } from '@/types/admin';

export interface BookItem {
  id: string;
  tenant_id?: string;
  title?: string;
  author?: string;
  isbn?: string;
  publisher?: string;
  publish_date?: string;
  category?: string;
  description?: string;
  status?: string;
  price?: number;
  stock_quantity?: number;
  cover_image_url?: string;
  tags?: string;
  created_at?: string;
  updated_at?: string;
}

export interface BookListQuery {
  keyword?: string;
  page?: number;
  page_size?: number;
}

const basePath = '/books';

export function listbooks(params: BookListQuery = {}): Promise<ListResponse<BookItem>> {
  return http.get<ListResponse<BookItem>>(basePath, { params });
}

export function getBook(id: string | number): Promise<BookItem> {
  return http.get<BookItem>(`${basePath}/${id}`);
}

export function createBook(data: Record<string, unknown>): Promise<BookItem> {
  return http.post<BookItem>(basePath, data);
}

export function updateBook(id: string | number, data: Record<string, unknown>): Promise<BookItem> {
  return http.put<BookItem>(`${basePath}/${id}`, data);
}

export function deleteBook(id: string | number): Promise<{ deleted: boolean }> {
  return http.delete<{ deleted: boolean }>(`${basePath}/${id}`);
}
