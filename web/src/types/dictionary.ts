export interface DictionaryCategoryItem {
  id: string;
  code: string;
  name: string;
  description?: string;
  status: string;
  sort: number;
  remark?: string;
  created_at: string;
  updated_at: string;
}

export interface DictionaryCategoryQuery {
  keyword?: string;
  status?: string;
  page: number;
  page_size: number;
}

export interface DictionaryCategoryFormState {
  id?: string;
  code: string;
  name: string;
  description: string;
  status: string;
  sort: number;
  remark: string;
}

export interface DictionaryItem {
  id: string;
  category_id: string;
  value: string;
  label: string;
  tag_type?: string;
  tag_color?: string;
  extra?: string;
  is_default: boolean;
  status: string;
  sort: number;
  remark?: string;
  created_at: string;
  updated_at: string;
}

export interface DictionaryItemQuery {
  category_id?: string;
  category_code?: string;
  keyword?: string;
  status?: string;
  page: number;
  page_size: number;
}

export interface DictionaryItemFormState {
  id?: string;
  category_id: string;
  value: string;
  label: string;
  tag_type: string;
  tag_color: string;
  extra: string;
  is_default: boolean;
  status: string;
  sort: number;
  remark: string;
}

export interface DictionaryLookupResponse {
  items: DictionaryItem[];
}
