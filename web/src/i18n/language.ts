export const I18N_SUPPORTED_LANGUAGES = ['zh-CN', 'en-US'] as const;

export type I18nLanguage = (typeof I18N_SUPPORTED_LANGUAGES)[number];

export const I18N_DEFAULT_LANGUAGE: I18nLanguage = 'zh-CN';

export const I18N_DEFAULT_NS = 'app';

export const I18N_NAMESPACES = [I18N_DEFAULT_NS] as const;

export function normalizeI18nLanguage(value: string | null | undefined): I18nLanguage {
  const raw = typeof value === 'string' ? value.trim().toLowerCase() : '';
  if (raw.startsWith('en')) {
    return 'en-US';
  }
  return 'zh-CN';
}
