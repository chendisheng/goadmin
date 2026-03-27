import 'vue-router';

declare module 'vue-router' {
  interface RouteMeta {
    title?: string;
    subtitle?: string;
    icon?: string;
    inMenu?: boolean;
    hideInMenu?: boolean;
    public?: boolean;
    requiresAuth?: boolean;
    order?: number;
  }
}
