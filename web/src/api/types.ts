export interface ApiEnvelope<T = unknown> {
  code: number;
  msg: string;
  data: T;
  request_id?: string;
  timestamp?: number;
}

export class ApiError extends Error {
  readonly code: number;
  readonly payload: unknown;
  readonly requestId?: string;

  constructor(message: string, code: number, payload?: unknown, requestId?: string) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.payload = payload;
    this.requestId = requestId;
  }
}
