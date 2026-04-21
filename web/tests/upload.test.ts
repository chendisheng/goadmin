// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createUploadFilePreviewUrl, downloadUploadFile, uploadUploadFile } from '../src/api/upload';
import {
  canSubmitUploadForm,
  formatUploadFileSize,
  isPreviewableImage,
  resolveUploadStatusLabel,
  resolveUploadStatusTagType,
  resolveUploadVisibilityLabel,
  resolveUploadVisibilityTagType,
} from '../src/utils/upload';

vi.mock('../src/store/session', () => ({
  getStoredAccessToken: vi.fn(() => 'test-token'),
}));

afterEach(() => {
  vi.restoreAllMocks();
});

describe('upload helpers', () => {
  it('formats upload file sizes and labels consistently', () => {
    expect(formatUploadFileSize(undefined)).toBe('-');
    expect(formatUploadFileSize(512)).toBe('512 B');
    expect(formatUploadFileSize(2048)).toBe('2.0 KB');
    expect(formatUploadFileSize(1024 * 1024)).toBe('1.0 MB');

    expect(resolveUploadVisibilityLabel('private')).toBe('私有');
    expect(resolveUploadVisibilityLabel('public')).toBe('公开');
    expect(resolveUploadStatusLabel('active')).toBe('有效');
    expect(resolveUploadStatusLabel('archived')).toBe('已归档');
    expect(resolveUploadStatusLabel('deleted')).toBe('已删除');

    expect(resolveUploadVisibilityTagType('public')).toBe('success');
    expect(resolveUploadVisibilityTagType('private')).toBe('info');
    expect(resolveUploadStatusTagType('active')).toBe('success');
    expect(resolveUploadStatusTagType('archived')).toBe('warning');
    expect(resolveUploadStatusTagType('deleted')).toBe('danger');

    expect(isPreviewableImage('image/png')).toBe(true);
    expect(isPreviewableImage('application/pdf')).toBe(false);
  });

  it('refuses to submit upload form without a file', () => {
    expect(canSubmitUploadForm(null, { visibility: 'private' })).toBe(false);
    expect(canSubmitUploadForm(new File(['x'], 'demo.txt'), { visibility: '' })).toBe(false);
    expect(canSubmitUploadForm(new File(['x'], 'demo.txt'), { visibility: 'public' })).toBe(true);
  });
});

describe('upload api smoke', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('posts multipart form data with bearer auth when uploading a file', async () => {
    const responsePayload = {
      id: 'file-1',
      original_name: 'avatar.png',
      storage_name: 'upload-file.png',
      storage_key: 'uploads/2026/04/20/upload-file.png',
      storage_driver: 'local',
      storage_path: 'uploads/2026/04/20/upload-file.png',
      public_url: 'https://cdn.example.test/avatar.png',
      mime_type: 'image/png',
      extension: '.png',
      size_bytes: 128,
      checksum_sha256: 'abc123',
      visibility: 'public',
      biz_module: 'system',
      biz_type: 'avatar',
      biz_id: '42',
      biz_field: 'cover',
      uploaded_by: 'alice',
      status: 'active',
      remark: 'profile image',
      created_at: '2026-04-20T00:00:00Z',
      updated_at: '2026-04-20T00:00:00Z',
    };

    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ code: 200, msg: 'ok', data: responsePayload }), {
      status: 200,
      headers: { 'content-type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);

    const file = new File(['hello'], 'avatar.png', { type: 'image/png' });
    const result = await uploadUploadFile(file, {
      visibility: 'public',
      biz_module: 'system',
      biz_type: 'avatar',
      biz_id: '42',
      biz_field: 'cover',
      remark: 'profile image',
    });

    expect(result.id).toBe('file-1');
    expect(fetchMock).toHaveBeenCalledTimes(1);

    const [requestUrl, requestInit] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(requestUrl).toBe('/api/v1/uploads/files');
    expect(requestInit.method).toBe('POST');
    expect(requestInit.headers).toMatchObject({
      Authorization: 'Bearer test-token',
      'X-Requested-With': 'XMLHttpRequest',
    });

    const body = requestInit.body as FormData;
    expect(body).toBeInstanceOf(FormData);
    expect(Array.from(body.entries())).toEqual(expect.arrayContaining([
      ['file', file],
      ['visibility', 'public'],
      ['biz_module', 'system'],
      ['biz_type', 'avatar'],
      ['biz_id', '42'],
      ['biz_field', 'cover'],
      ['remark', 'profile image'],
    ]));
  });

  it('downloads a file using the backend filename and a temporary object URL', async () => {
    const fetchMock = vi.fn(async () => new Response(new Blob(['demo file']), {
      status: 200,
      headers: {
        'content-type': 'application/octet-stream',
        'content-disposition': 'attachment; filename="report.pdf"',
      },
    }));
    vi.stubGlobal('fetch', fetchMock);

    const urlApi = window.URL as typeof window.URL & {
      createObjectURL?: (blob: Blob | MediaSource) => string;
      revokeObjectURL?: (url: string) => void;
    };
    Object.defineProperty(urlApi, 'createObjectURL', {
      value: vi.fn(() => 'blob:upload-test'),
      writable: true,
    });
    Object.defineProperty(urlApi, 'revokeObjectURL', {
      value: vi.fn(() => undefined),
      writable: true,
    });
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined);

    await downloadUploadFile('file-1', 'fallback.pdf');

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(urlApi.createObjectURL).toHaveBeenCalledTimes(1);
    expect(urlApi.revokeObjectURL).toHaveBeenCalledWith('blob:upload-test');
    expect(clickSpy).toHaveBeenCalledTimes(1);
  });

  it('creates a preview blob URL using authenticated download response', async () => {
    const fetchMock = vi.fn(async () => new Response(new Blob(['preview file']), {
      status: 200,
      headers: {
        'content-type': 'application/octet-stream',
        'content-disposition': 'attachment; filename="preview.png"',
      },
    }));
    vi.stubGlobal('fetch', fetchMock);

    const urlApi = window.URL as typeof window.URL & {
      createObjectURL?: (blob: Blob | MediaSource) => string;
    };
    Object.defineProperty(urlApi, 'createObjectURL', {
      value: vi.fn(() => 'blob:preview-test'),
      writable: true,
    });

    const previewUrl = await createUploadFilePreviewUrl('file-1');

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(previewUrl).toBe('blob:preview-test');
    expect(urlApi.createObjectURL).toHaveBeenCalledTimes(1);
  });
});
