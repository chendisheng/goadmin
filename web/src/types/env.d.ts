/// <reference types="vite/client" />

declare interface ImportMetaEnv {
  readonly VITE_APP_TITLE?: string;
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_API_PROXY_TARGET?: string;
  readonly VITE_DEV_PORT?: string;
}

declare interface ImportMeta {
  readonly env: ImportMetaEnv;
}
