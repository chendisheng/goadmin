import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
const ACCESS_TOKEN_KEY = 'goadmin.access_token';
const REFRESH_TOKEN_KEY = 'goadmin.refresh_token';
const TOKEN_TYPE_KEY = 'goadmin.token_type';
const ACCESS_EXPIRES_AT_KEY = 'goadmin.access_expires_at';
const REFRESH_EXPIRES_AT_KEY = 'goadmin.refresh_expires_at';
const LANGUAGE_KEY = 'goadmin.language';
function canUseStorage() {
    return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
}
function readLanguageValue() {
    if (!canUseStorage()) {
        return 'zh-CN';
    }
    return window.localStorage.getItem(LANGUAGE_KEY) || 'zh-CN';
}
function readAccessToken() {
    if (!canUseStorage()) {
        return '';
    }
    return window.localStorage.getItem(ACCESS_TOKEN_KEY) ?? '';
}
function readStringValue(key) {
    if (!canUseStorage()) {
        return '';
    }
    return window.localStorage.getItem(key) ?? '';
}
function readNumberValue(key) {
    const raw = readStringValue(key);
    if (raw === '') {
        return 0;
    }
    const parsed = Number(raw);
    return Number.isFinite(parsed) ? parsed : 0;
}
function persistAccessToken(token) {
    if (!canUseStorage()) {
        return;
    }
    if (token.trim() === '') {
        window.localStorage.removeItem(ACCESS_TOKEN_KEY);
        return;
    }
    window.localStorage.setItem(ACCESS_TOKEN_KEY, token);
}
function persistStringValue(key, value) {
    if (!canUseStorage()) {
        return;
    }
    if (value.trim() === '') {
        window.localStorage.removeItem(key);
        return;
    }
    window.localStorage.setItem(key, value);
}
function persistNumberValue(key, value) {
    if (!canUseStorage()) {
        return;
    }
    if (!Number.isFinite(value) || value <= 0) {
        window.localStorage.removeItem(key);
        return;
    }
    window.localStorage.setItem(key, String(value));
}
function persistLanguageValue(value) {
    if (!canUseStorage()) {
        return;
    }
    const language = value.trim();
    if (language === '') {
        window.localStorage.removeItem(LANGUAGE_KEY);
        return;
    }
    window.localStorage.setItem(LANGUAGE_KEY, language);
}
export function getStoredAccessToken() {
    return readAccessToken();
}
export const useSessionStore = defineStore('session', () => {
    const accessToken = ref(readAccessToken());
    const refreshToken = ref(readStringValue(REFRESH_TOKEN_KEY));
    const tokenType = ref(readStringValue(TOKEN_TYPE_KEY) || 'Bearer');
    const accessExpiresAt = ref(readNumberValue(ACCESS_EXPIRES_AT_KEY));
    const refreshExpiresAt = ref(readNumberValue(REFRESH_EXPIRES_AT_KEY));
    const language = ref(readLanguageValue());
    const currentUser = ref(null);
    const isAuthenticated = computed(() => accessToken.value.trim().length > 0);
    const displayName = computed(() => currentUser.value?.display_name || currentUser.value?.username || '访客');
    const roleLabels = computed(() => currentUser.value?.roles ?? []);
    const permissions = computed(() => currentUser.value?.permissions ?? []);
    function normalizePermission(value) {
        return value.trim();
    }
    function hasPermission(permission) {
        const list = permissions.value.map(normalizePermission).filter((item) => item !== '');
        if (list.includes('*')) {
            return true;
        }
        const candidates = Array.isArray(permission) ? permission : [permission];
        return candidates.some((candidate) => list.includes(normalizePermission(candidate)));
    }
    function hydrate() {
        accessToken.value = readAccessToken();
        refreshToken.value = readStringValue(REFRESH_TOKEN_KEY);
        tokenType.value = readStringValue(TOKEN_TYPE_KEY) || 'Bearer';
        accessExpiresAt.value = readNumberValue(ACCESS_EXPIRES_AT_KEY);
        refreshExpiresAt.value = readNumberValue(REFRESH_EXPIRES_AT_KEY);
        language.value = readLanguageValue();
        currentUser.value = null;
    }
    function applyLoginResponse(response) {
        accessToken.value = response.access_token.trim();
        refreshToken.value = response.refresh_token.trim();
        tokenType.value = response.token_type.trim() || 'Bearer';
        accessExpiresAt.value = Math.max(0, Math.trunc(Date.now() / 1000) + Math.max(0, response.expires_in));
        refreshExpiresAt.value = Math.max(0, Math.trunc(Date.now() / 1000) + Math.max(0, response.refresh_expires_in));
        currentUser.value = response.user;
        language.value = response.user.language?.trim() || language.value || 'zh-CN';
        persistAccessToken(accessToken.value);
        persistStringValue(REFRESH_TOKEN_KEY, refreshToken.value);
        persistStringValue(TOKEN_TYPE_KEY, tokenType.value);
        persistNumberValue(ACCESS_EXPIRES_AT_KEY, accessExpiresAt.value);
        persistNumberValue(REFRESH_EXPIRES_AT_KEY, refreshExpiresAt.value);
        persistLanguageValue(language.value);
    }
    function setCurrentUser(user) {
        currentUser.value = user;
    }
    function setAccessToken(token) {
        accessToken.value = token.trim();
        persistAccessToken(accessToken.value);
    }
    function clearSession() {
        accessToken.value = '';
        refreshToken.value = '';
        tokenType.value = 'Bearer';
        accessExpiresAt.value = 0;
        refreshExpiresAt.value = 0;
        currentUser.value = null;
        persistAccessToken('');
        persistStringValue(REFRESH_TOKEN_KEY, '');
        persistStringValue(TOKEN_TYPE_KEY, '');
        persistNumberValue(ACCESS_EXPIRES_AT_KEY, 0);
        persistNumberValue(REFRESH_EXPIRES_AT_KEY, 0);
    }
    function setLanguage(value) {
        language.value = value.trim() || 'zh-CN';
        persistLanguageValue(language.value);
    }
    return {
        accessToken,
        refreshToken,
        tokenType,
        accessExpiresAt,
        refreshExpiresAt,
        language,
        currentUser,
        isAuthenticated,
        displayName,
        roleLabels,
        permissions,
        hasPermission,
        hydrate,
        applyLoginResponse,
        setCurrentUser,
        setAccessToken,
        setLanguage,
        clearSession,
    };
});
