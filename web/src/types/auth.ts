export interface LoginRequest {
  username: string;
  password: string;
}

export interface AuthUser {
  user_id: string;
  tenant_id?: string;
  username: string;
  display_name?: string;
  language?: string;
  roles?: string[];
  permissions?: string[];
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  refresh_expires_in: number;
  user: AuthUser;
}
