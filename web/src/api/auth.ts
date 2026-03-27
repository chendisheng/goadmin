import http from './http';

import type { AuthUser, LoginRequest, LoginResponse } from '@/types/auth';

export function login(payload: LoginRequest): Promise<LoginResponse> {
  return http.post<LoginResponse>('/auth/login', payload);
}

export function fetchCurrentUser(): Promise<AuthUser> {
  return http.get<AuthUser>('/auth/me');
}

export function logout(): Promise<{ logged_out: boolean }> {
  return http.post<{ logged_out: boolean }>('/auth/logout');
}
