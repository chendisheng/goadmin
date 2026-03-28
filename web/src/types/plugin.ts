export interface PluginMenu {
  plugin?: string;
  id: string;
  parent_id?: string;
  name: string;
  path: string;
  component?: string;
  icon?: string;
  sort: number;
  permission?: string;
  type: string;
  visible: boolean;
  enabled: boolean;
  redirect?: string;
  external_url?: string;
  children?: PluginMenu[];
}

export interface PluginMenuListResponse {
  items: PluginMenu[];
}

export interface PluginItem {
  name: string;
  description?: string;
  enabled: boolean;
  menus: PluginMenu[];
  permissions: PluginPermission[];
  created_at: string;
  updated_at: string;
}

export interface PluginListResponse {
  total: number;
  items: PluginItem[];
}

export interface PluginPermission {
  plugin?: string;
  object: string;
  action: string;
  description: string;
}

export interface PluginPermissionListResponse {
  items: PluginPermission[];
}

export interface PluginFormState {
  name: string;
  description: string;
  enabled: boolean;
  menus: PluginMenu[];
  permissions: PluginPermission[];
}

export interface ExamplePluginPingResponse {
  message: string;
  plugin: string;
}
