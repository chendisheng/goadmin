import http from './http';

import type {
  DictionaryCategoryFormState,
  DictionaryCategoryItem,
  DictionaryCategoryQuery,
  DictionaryItem,
  DictionaryItemFormState,
  DictionaryItemQuery,
  DictionaryLookupResponse,
} from '@/types/dictionary';
import type { ListResponse } from '@/types/admin';

const categoryBasePath = '/dictionaries/categories';
const itemBasePath = '/dictionaries/items';
const lookupBasePath = '/dictionaries/lookup';

export function fetchDictionaryCategories(params: DictionaryCategoryQuery): Promise<ListResponse<DictionaryCategoryItem>> {
  return http.get<ListResponse<DictionaryCategoryItem>>(categoryBasePath, { params });
}

export function fetchDictionaryCategory(id: string): Promise<DictionaryCategoryItem> {
  return http.get<DictionaryCategoryItem>(`${categoryBasePath}/${id}`);
}

export function createDictionaryCategory(payload: DictionaryCategoryFormState): Promise<DictionaryCategoryItem> {
  return http.post<DictionaryCategoryItem>(categoryBasePath, payload);
}

export function updateDictionaryCategory(id: string, payload: DictionaryCategoryFormState): Promise<DictionaryCategoryItem> {
  return http.put<DictionaryCategoryItem>(`${categoryBasePath}/${id}`, payload);
}

export function deleteDictionaryCategory(id: string): Promise<{ deleted: boolean }> {
  return http.delete<{ deleted: boolean }>(`${categoryBasePath}/${id}`);
}

export function fetchDictionaryItems(params: DictionaryItemQuery): Promise<ListResponse<DictionaryItem>> {
  return http.get<ListResponse<DictionaryItem>>(itemBasePath, { params });
}

export function fetchDictionaryItem(id: string): Promise<DictionaryItem> {
  return http.get<DictionaryItem>(`${itemBasePath}/${id}`);
}

export function createDictionaryItem(payload: DictionaryItemFormState): Promise<DictionaryItem> {
  return http.post<DictionaryItem>(itemBasePath, payload);
}

export function updateDictionaryItem(id: string, payload: DictionaryItemFormState): Promise<DictionaryItem> {
  return http.put<DictionaryItem>(`${itemBasePath}/${id}`, payload);
}

export function deleteDictionaryItem(id: string): Promise<{ deleted: boolean }> {
  return http.delete<{ deleted: boolean }>(`${itemBasePath}/${id}`);
}

export function fetchDictionaryLookupItems(categoryCode: string): Promise<DictionaryLookupResponse> {
  return http.get<DictionaryLookupResponse>(`${lookupBasePath}/${categoryCode}`);
}

export function fetchDictionaryLookupItem(categoryCode: string, value: string): Promise<DictionaryItem> {
  return http.get<DictionaryItem>(`${lookupBasePath}/${categoryCode}/${value}`);
}
