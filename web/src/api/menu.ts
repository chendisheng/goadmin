import http from './http';

import type { BackendMenuRoutesResponse } from '@/types/menu';

export function fetchMenuRoutes(): Promise<BackendMenuRoutesResponse> {
  return http.get<BackendMenuRoutesResponse>('/menus/routes');
}
