export const I18N_SUPPORTED_LANGUAGES = ['zh-CN', 'en-US'] as const;

export type I18nLanguage = (typeof I18N_SUPPORTED_LANGUAGES)[number];

export const I18N_DEFAULT_LANGUAGE: I18nLanguage = 'zh-CN';

export const I18N_DEFAULT_NS = 'app';

export const I18N_NAMESPACES = [I18N_DEFAULT_NS] as const;

export function tryNormalizeI18nLanguage(value: string | null | undefined): I18nLanguage | null {
  const raw = typeof value === 'string' ? value.trim().toLowerCase() : '';
  if (raw.startsWith('en')) {
    return 'en-US';
  }
  if (raw.startsWith('zh')) {
    return 'zh-CN';
  }
  return null;
}

export function normalizeI18nLanguage(value: string | null | undefined): I18nLanguage {
  return tryNormalizeI18nLanguage(value) ?? 'zh-CN';
}

function normalizeBrowserLanguageCandidate(value: unknown): I18nLanguage | null {
  if (typeof value !== 'string') {
    return null;
  }
  const raw = value.trim();
  if (raw === '') {
    return null;
  }
  return tryNormalizeI18nLanguage(raw);
}

function detectBrowserLanguage(): I18nLanguage | null {
  if (typeof window === 'undefined' || typeof window.navigator === 'undefined') {
    return null;
  }

  const navigatorLike = window.navigator as Navigator & { userLanguage?: string };
  const candidates = [navigatorLike.language, navigatorLike.languages?.[0], navigatorLike.userLanguage];

  for (const candidate of candidates) {
    const resolved = normalizeBrowserLanguageCandidate(candidate);
    if (resolved !== null) {
      return resolved;
    }
  }

  return null;
}

export function resolveInitialI18nLanguage(value?: string | null): I18nLanguage {
  const storedLanguage = normalizeBrowserLanguageCandidate(value);
  if (storedLanguage !== null) {
    return storedLanguage;
  }

  return detectBrowserLanguage() ?? I18N_DEFAULT_LANGUAGE;
}

export function resolvePreferredI18nLanguage(
  explicitLanguage?: string | null,
  profileLanguage?: string | null,
): I18nLanguage {
  return tryNormalizeI18nLanguage(explicitLanguage)
    ?? tryNormalizeI18nLanguage(profileLanguage)
    ?? resolveInitialI18nLanguage();
}
