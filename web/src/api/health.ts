import http from './http';

export interface HealthPayload {
  status: string;
  uptime: string;
  timestamp: string;
}

export function fetchHealth(): Promise<HealthPayload> {
  return http.get<HealthPayload>('/health');
}
