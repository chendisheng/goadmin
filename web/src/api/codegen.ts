import http from './http';

export interface CodegenDslRequest {
  dsl: string;
  force?: boolean;
}

export interface CodegenDslPreviewItem {
  index: number;
  kind: string;
  name: string;
  force: boolean;
  actions: string[];
}

export interface CodegenDslExecutionReport {
  dry_run: boolean;
  messages: string[];
  items: CodegenDslPreviewItem[];
}

export function previewCodegenDsl(payload: CodegenDslRequest) {
  return http.post<CodegenDslExecutionReport>('/codegen/dsl/preview', payload);
}

export function generateCodegenDsl(payload: CodegenDslRequest) {
  return http.post<CodegenDslExecutionReport>('/codegen/dsl/generate', payload);
}
