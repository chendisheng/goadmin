import http from './http';

import type {
  ExamplePluginPingResponse,
  PluginFormState,
  PluginMenuListResponse,
  PluginListResponse,
  PluginItem,
  PluginPermissionListResponse,
} from '@/types/plugin';

export function fetchPlugins(): Promise<PluginListResponse> {
  return http.get<PluginListResponse>('/plugins');
}

export function fetchPlugin(name: string): Promise<PluginItem> {
  return http.get<PluginItem>(`/plugins/${name}`);
}

export function createPlugin(payload: PluginFormState): Promise<PluginItem> {
  return http.post<PluginItem>('/plugins', payload);
}

export function updatePlugin(name: string, payload: PluginFormState): Promise<PluginItem> {
  return http.put<PluginItem>(`/plugins/${name}`, payload);
}

export function deletePlugin(name: string): Promise<{ deleted: boolean }> {
  return http.delete<{ deleted: boolean }>(`/plugins/${name}`);
}

export function fetchPluginMenus(): Promise<PluginMenuListResponse> {
  return http.get<PluginMenuListResponse>('/plugins/menus');
}

export function fetchPluginPermissions(): Promise<PluginPermissionListResponse> {
  return http.get<PluginPermissionListResponse>('/plugins/permissions');
}

export function pingExamplePlugin(): Promise<ExamplePluginPingResponse> {
  return http.get<ExamplePluginPingResponse>('/plugins/example/ping');
}
