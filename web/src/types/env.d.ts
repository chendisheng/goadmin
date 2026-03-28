/// <reference types="vite/client" />

declare interface ImportMetaEnv {
  readonly BASE_URL?: string;
  readonly VITE_APP_TITLE?: string;
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_API_PROXY_TARGET?: string;
  readonly VITE_DEV_PORT?: string;
}

declare interface ImportMeta {
  readonly env: ImportMetaEnv;
  glob: (pattern: string | string[], options?: Record<string, unknown>) => Record<string, unknown>;
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue';

  const component: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>;
  export default component;
}
