export interface ServerMenuRouteMeta {
  title: string;
  titleKey?: string;
  titleDefault?: string;
  icon?: string;
  permission?: string;
  hidden?: boolean;
  noCache?: boolean;
  affix?: boolean;
  link?: string;
  subtitle?: string;
  subtitleKey?: string;
  subtitleDefault?: string;
}

export type ServerMenuRouteType = 'directory' | 'menu' | 'button';

export interface ServerMenuRoute {
  name: string;
  path: string;
  component?: string;
  redirect?: string;
  hidden: boolean;
  alwaysShow?: boolean;
  type?: ServerMenuRouteType;
  meta: ServerMenuRouteMeta;
  children?: ServerMenuRoute[];
}

export interface ServerMenuRoutesResponse {
  items: ServerMenuRoute[];
}

export interface SidebarMenuNode {
  name: string;
  path: string;
  title: string;
  titleKey?: string;
  titleDefault?: string;
  icon?: string;
  component?: string;
  redirect?: string;
  permission?: string;
  hidden: boolean;
  subtitle?: string;
  subtitleKey?: string;
  subtitleDefault?: string;
  children: SidebarMenuNode[];
}
