export const I18N_SUPPORTED_LANGUAGES = ['zh-CN', 'en-US'];
export const I18N_DEFAULT_LANGUAGE = 'zh-CN';
export const I18N_DEFAULT_NS = 'app';
export const I18N_NAMESPACES = [I18N_DEFAULT_NS];
export function normalizeI18nLanguage(value) {
    const raw = typeof value === 'string' ? value.trim().toLowerCase() : '';
    if (raw.startsWith('en')) {
        return 'en-US';
    }
    return 'zh-CN';
}
