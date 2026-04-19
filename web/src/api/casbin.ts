import http from './http';

export interface AuthorizationModuleStatus {
  enabled?: boolean;
  source?: string;
  model_path?: string;
  policy_path?: string;
  summary?: string;
  legacy_modules?: string[];
  routes?: string[];
}

export function fetchAuthorizationStatus(): Promise<AuthorizationModuleStatus> {
  return http.get<AuthorizationModuleStatus>('/casbin/status');
}

export function reloadAuthorizationPolicies(): Promise<{ reloaded?: boolean }> {
  return http.post<{ reloaded?: boolean }>('/casbin/reload');
}

export function seedAuthorizationPolicies(): Promise<{ seeded?: boolean }> {
  return http.post<{ seeded?: boolean }>('/casbin/seed');
}

export type CasbinModuleStatus = AuthorizationModuleStatus;
export const fetchCasbinStatus = fetchAuthorizationStatus;
export const reloadCasbinPolicies = reloadAuthorizationPolicies;
export const seedCasbinPolicies = seedAuthorizationPolicies;
