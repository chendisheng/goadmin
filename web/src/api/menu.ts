import http from './http';

import type { ServerMenuRoutesResponse } from '@/types/menu';

export function fetchMenuRoutes(): Promise<ServerMenuRoutesResponse> {
  return http.get<ServerMenuRoutesResponse>('/menus/routes');
}
