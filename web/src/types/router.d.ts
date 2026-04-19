import 'vue-router';

declare module 'vue-router' {
  interface RouteMeta {
    title?: string;
    subtitle?: string;
    icon?: string;
    componentName?: string;
    alwaysShow?: boolean;
    affix?: boolean;
    keepAlive?: boolean;
    restorable?: boolean;
    permission?: string;
    link?: string;
    inMenu?: boolean;
    hideInMenu?: boolean;
    public?: boolean;
    requiresAuth?: boolean;
    order?: number;
  }
}
