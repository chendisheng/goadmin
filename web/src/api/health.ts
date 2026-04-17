import http from './http';

export interface HealthPayload {
  status: string;
  uptime: string;
  timestamp: string;
}

export interface PublicDatabaseConfig {
  driver: string;
  name: string;
}

export interface PublicConfigPayload {
  app?: {
    name?: string;
    env?: string;
    version?: string;
  };
  server?: Record<string, unknown>;
  logger?: Record<string, unknown>;
  database?: PublicDatabaseConfig;
  codegen?: Record<string, unknown>;
  auth?: Record<string, unknown>;
  loaded_at?: string;
  loaded_from?: string;
}

export function fetchHealth(): Promise<HealthPayload> {
  return http.get<HealthPayload>('/health');
}

export function fetchPublicConfig(): Promise<PublicConfigPayload> {
  return http.get<PublicConfigPayload>('/meta/config');
}
