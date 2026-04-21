import type { UploadFileFormState } from '@/types/upload';

export type UploadPreviewKind = 'image' | 'pdf' | 'text' | 'download-only';

export function formatUploadFileSize(bytes?: number): string {
  if (bytes == null || Number.isNaN(Number(bytes))) {
    return '-';
  }
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  if (bytes < 1024 * 1024 * 1024) {
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}

export function resolveUploadVisibilityLabel(value?: string): string {
  if (!value) {
    return '-';
  }
  switch (value) {
    case 'public':
      return '公开';
    case 'private':
      return '私有';
    default:
      return value;
  }
}

export function resolveUploadVisibilityTagType(value?: string): 'success' | 'info' | 'warning' {
  return value === 'public' ? 'success' : 'info';
}

export function resolveUploadStatusLabel(value?: string): string {
  if (!value) {
    return '-';
  }
  switch (value) {
    case 'active':
      return '有效';
    case 'archived':
      return '已归档';
    case 'deleted':
      return '已删除';
    default:
      return value;
  }
}

export function resolveUploadStatusTagType(value?: string): 'success' | 'warning' | 'danger' | 'info' {
  switch (value) {
    case 'active':
      return 'success';
    case 'archived':
      return 'warning';
    case 'deleted':
      return 'danger';
    default:
      return 'info';
  }
}

export function isPreviewableImage(mimeType?: string): boolean {
  return Boolean(mimeType?.toLowerCase().startsWith('image/'));
}

export function isPreviewablePdf(mimeType?: string): boolean {
  return normalizeMimeType(mimeType) === 'application/pdf';
}

export function isPreviewableText(mimeType?: string): boolean {
  const normalized = normalizeMimeType(mimeType);
  if (!normalized) {
    return false;
  }
  return normalized.startsWith('text/') || ['application/json', 'application/xml', 'application/xhtml+xml'].includes(normalized);
}

export function isPreviewableDocument(mimeType?: string): boolean {
  return isPreviewablePdf(mimeType) || isPreviewableText(mimeType);
}

export function resolveUploadPreviewKind(mimeType?: string): UploadPreviewKind {
  if (isPreviewableImage(mimeType)) {
    return 'image';
  }
  if (isPreviewablePdf(mimeType)) {
    return 'pdf';
  }
  if (isPreviewableText(mimeType)) {
    return 'text';
  }
  return 'download-only';
}

export function isBrowserDirectPublicUrl(url?: string): boolean {
  const normalized = url?.trim() ?? '';
  if (!normalized) {
    return false;
  }
  if (/^https?:\/\//i.test(normalized)) {
    return true;
  }
  return normalized.startsWith('/') && !normalized.startsWith('/api/');
}

export function canSubmitUploadForm(file: File | null, form: Pick<UploadFileFormState, 'visibility'>): boolean {
  return file instanceof File && Boolean(form.visibility.trim());
}

function normalizeMimeType(value?: string): string {
  return value?.trim().toLowerCase() ?? '';
}
