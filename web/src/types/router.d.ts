import 'vue-router';

declare module 'vue-router' {
  interface RouteMeta {
    title?: string;
    titleKey?: string;
    titleDefault?: string;
    subtitle?: string;
    subtitleKey?: string;
    subtitleDefault?: string;
    i18nNamespaces?: string[];
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
