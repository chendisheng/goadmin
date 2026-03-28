export interface BackendMenuRouteMeta {
  title: string;
  icon?: string;
  permission?: string;
  hidden?: boolean;
  noCache?: boolean;
  affix?: boolean;
  link?: string;
}

export interface BackendMenuRoute {
  name: string;
  path: string;
  component?: string;
  redirect?: string;
  hidden: boolean;
  alwaysShow?: boolean;
  meta: BackendMenuRouteMeta;
  children?: BackendMenuRoute[];
}

export interface BackendMenuRoutesResponse {
  items: BackendMenuRoute[];
}

export interface SidebarMenuNode {
  name: string;
  path: string;
  title: string;
  icon?: string;
  component?: string;
  redirect?: string;
  permission?: string;
  hidden: boolean;
  children: SidebarMenuNode[];
}
