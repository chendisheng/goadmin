import { computed } from 'vue';
import { useLocaleStore } from '@/store/locale';
import { appMessages } from './messages';
function isMessageKey(value) {
    return Object.prototype.hasOwnProperty.call(appMessages['zh-CN'], value);
}
function formatMessage(template, values) {
    if (!values) {
        return template;
    }
    return template.replace(/\{([a-zA-Z0-9_]+)\}/g, (_, token) => {
        const value = values[token];
        return value === undefined || value === null ? `{${token}}` : String(value);
    });
}
export function getAppLanguage() {
    const localeStore = useLocaleStore();
    return localeStore.language;
}
export function translate(key, fallback = '', values) {
    const localeStore = useLocaleStore();
    const locale = localeStore.language;
    const normalizedKey = key.trim();
    if (normalizedKey !== '' && isMessageKey(normalizedKey)) {
        const bucket = (appMessages[locale] ?? appMessages['zh-CN']);
        const fallbackBucket = appMessages['zh-CN'];
        const translated = bucket[normalizedKey] ?? fallbackBucket[normalizedKey];
        if (translated) {
            return formatMessage(translated, values);
        }
    }
    if (fallback.trim() !== '') {
        return formatMessage(fallback, values);
    }
    return normalizedKey;
}
export function useAppI18n() {
    const localeStore = useLocaleStore();
    const language = computed(() => localeStore.language);
    function t(key, fallback = '', values) {
        const currentLanguage = language.value;
        const normalizedKey = key.trim();
        if (normalizedKey !== '' && isMessageKey(normalizedKey)) {
            const bucket = (appMessages[currentLanguage] ?? appMessages['zh-CN']);
            const fallbackBucket = appMessages['zh-CN'];
            const translated = bucket[normalizedKey] ?? fallbackBucket[normalizedKey];
            if (translated) {
                return formatMessage(translated, values);
            }
        }
        if (fallback.trim() !== '') {
            return formatMessage(fallback, values);
        }
        return normalizedKey;
    }
    return {
        language,
        t,
    };
}
export function getRouteTitle(meta, fallback = '') {
    const localeStore = useLocaleStore();
    const locale = localeStore.language;
    const key = (meta.titleKey || meta.componentName || meta.title || '').trim();
    const defaultTitle = (meta.titleDefault || meta.title || fallback).trim();
    if (key !== '' && isMessageKey(key)) {
        const bucket = (appMessages[locale] ?? appMessages['zh-CN']);
        const fallbackBucket = appMessages['zh-CN'];
        const translated = bucket[key] ?? fallbackBucket[key];
        if (translated) {
            return translated;
        }
    }
    return defaultTitle || fallback || meta.title || '';
}
export function resolveRouteLocaleMeta(route) {
    const meta = route.meta;
    const title = getRouteTitle(meta, '');
    const subtitleKey = typeof meta.subtitleKey === 'string' ? meta.subtitleKey : '';
    const subtitleDefault = typeof meta.subtitleDefault === 'string' && meta.subtitleDefault.trim() !== ''
        ? meta.subtitleDefault
        : typeof meta.subtitle === 'string'
            ? meta.subtitle
            : '';
    const subtitle = subtitleKey.trim() !== '' && isMessageKey(subtitleKey.trim())
        ? translate(subtitleKey, subtitleDefault)
        : subtitleDefault;
    return {
        title,
        subtitle,
    };
}
