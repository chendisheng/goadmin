export interface UploadFileItem {
  id: string;
  tenant_id?: string;
  original_name?: string;
  storage_name?: string;
  storage_key?: string;
  storage_driver?: string;
  storage_path?: string;
  public_url?: string;
  mime_type?: string;
  extension?: string;
  size_bytes?: number;
  checksum_sha256?: string;
  visibility?: string;
  biz_module?: string;
  biz_type?: string;
  biz_id?: string;
  biz_field?: string;
  uploaded_by?: string;
  status?: string;
  remark?: string;
  created_at: string;
  updated_at: string;
}

export interface UploadFileQuery {
  keyword?: string;
  visibility?: string;
  status?: string;
  biz_module?: string;
  biz_type?: string;
  biz_id?: string;
  uploaded_by?: string;
  page: number;
  page_size: number;
}

export interface UploadFileFormState {
  visibility: string;
  biz_module: string;
  biz_type: string;
  biz_id: string;
  biz_field: string;
  remark: string;
}

export interface UploadFileBindFormState {
  biz_module: string;
  biz_type: string;
  biz_id: string;
  biz_field: string;
}
