export interface ListResponse<T> {
  total: number;
  items: T[];
}

export interface UserItem {
  id: string;
  tenant_id?: string;
  username: string;
  display_name?: string;
  mobile?: string;
  email?: string;
  status: string;
  role_codes?: string[];
  created_at: string;
  updated_at: string;
}

export interface RoleItem {
  id: string;
  tenant_id?: string;
  name: string;
  code: string;
  status: string;
  remark?: string;
  menu_ids?: string[];
  created_at: string;
  updated_at: string;
}

export interface MenuItem {
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
  children?: MenuItem[];
  created_at: string;
  updated_at: string;
}

export interface UserQuery {
  tenant_id?: string;
  keyword?: string;
  status?: string;
  page: number;
  page_size: number;
}

export interface RoleQuery {
  tenant_id?: string;
  keyword?: string;
  status?: string;
  page: number;
  page_size: number;
}

export interface MenuQuery {
  keyword?: string;
  parent_id?: string;
  visible?: boolean;
  enabled?: boolean;
  page: number;
  page_size: number;
}

export interface UserFormState {
  tenant_id: string;
  username: string;
  display_name: string;
  mobile: string;
  email: string;
  status: string;
  role_codes: string[];
  password_hash: string;
}

export interface RoleFormState {
  tenant_id: string;
  name: string;
  code: string;
  status: string;
  remark: string;
  menu_ids: string[];
}

export interface MenuFormState {
  parent_id: string;
  name: string;
  path: string;
  component: string;
  icon: string;
  sort: number;
  permission: string;
  type: string;
  visible: boolean;
  enabled: boolean;
  redirect: string;
  external_url: string;
}
