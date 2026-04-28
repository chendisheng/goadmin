// @vitest-environment jsdom
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

import { initializeI18n, resolveRouteLocaleMeta, translate } from '../src/i18n';
import { useLocaleStore } from '../src/store/locale';
import { useSessionStore } from '../src/store/session';

const fetchCurrentUser = vi.fn();

vi.mock('@/api/auth', () => ({
  fetchCurrentUser,
}));

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
  fetchCurrentUser.mockReset();
  Object.defineProperty(window.navigator, 'language', {
    configurable: true,
    value: 'zh-CN',
  });
  setActivePinia(createPinia());
});

describe('useLocaleStore', () => {
  it('defaults to the browser language when nothing is persisted', async () => {
    Object.defineProperty(window.navigator, 'language', {
      configurable: true,
      value: 'en-GB',
    });

    const localeStore = useLocaleStore();

    expect(localeStore.language).toBe('en-US');

    await initializeI18n(localeStore.language);
    expect(translate('common.language', 'Language')).toBe('Language');
  });

  it('normalizes and persists the language preference', async () => {
    const localeStore = useLocaleStore();

    localeStore.setLanguage('en');
    await initializeI18n(localeStore.language);

    expect(localeStore.language).toBe('en-US');
    expect(localStorage.getItem('goadmin.language')).toBe('en-US');
  });

  it('translates a shared base namespace fallback in zh-CN', async () => {
    const localeStore = useLocaleStore();

    localeStore.setLanguage('zh-CN');
    await initializeI18n(localeStore.language);

    expect(translate('common.expand_sidebar', 'Expand sidebar')).toBe('展开侧栏');
  });

  it('syncs from the authenticated user language during session restore', () => {
    const localeStore = useLocaleStore();
    const sessionStore = useSessionStore();

    sessionStore.setLanguage('en-US');
    localeStore.syncFromUser({ language: 'en-US' });

    expect(sessionStore.language).toBe('en-US');
    expect(localeStore.language).toBe('en-US');
    expect(localStorage.getItem('goadmin.language')).toBe('en-US');
  });

  it('lets the server profile language win when the login page did not make an explicit choice', () => {
    const localeStore = useLocaleStore();
    const sessionStore = useSessionStore();

    expect(localeStore.hasLanguagePreference).toBe(false);

    sessionStore.applyLoginResponse({
      access_token: 'access-token',
      refresh_token: 'refresh-token',
      token_type: 'Bearer',
      expires_in: 3600,
      refresh_expires_in: 7200,
      user: {
        user_id: '1',
        username: 'admin',
        language: 'en-US',
      },
    }, null);

    localeStore.applyLanguagePreference(undefined, sessionStore.currentUser?.language, false);

    expect(sessionStore.language).toBe('en-US');
    expect(localeStore.language).toBe('en-US');
    expect(localeStore.hasLanguagePreference).toBe(false);
  });

  it('prefers the explicit language over the profile language', () => {
    const localeStore = useLocaleStore();

    localeStore.applyLanguagePreference('en-US', 'zh-CN');

    expect(localeStore.language).toBe('en-US');
    expect(localStorage.getItem('goadmin.language')).toBe('en-US');
  });

  it('does not mark a profile-driven fallback as an explicit preference', () => {
    const localeStore = useLocaleStore();

    localeStore.applyLanguagePreference(undefined, 'en-US', false);

    expect(localeStore.language).toBe('en-US');
    expect(localeStore.hasLanguagePreference).toBe(false);
    expect(localStorage.getItem('goadmin.language_preference')).toBeNull();
  });

  it('keeps the login page language when applying the login response', () => {
    const localeStore = useLocaleStore();
    const sessionStore = useSessionStore();

    localeStore.setLanguage('en-US');
    sessionStore.setLanguage('en-US');

    sessionStore.applyLoginResponse({
      access_token: 'access-token',
      refresh_token: 'refresh-token',
      token_type: 'Bearer',
      expires_in: 3600,
      refresh_expires_in: 7200,
      user: {
        user_id: '1',
        username: 'admin',
        language: 'zh-CN',
      },
    });

    expect(sessionStore.language).toBe('en-US');
    expect(localeStore.language).toBe('en-US');
    expect(localStorage.getItem('goadmin.language')).toBe('en-US');
  });

  it('keeps the login page language after authenticated session restore', async () => {
    fetchCurrentUser.mockResolvedValue({
      user_id: '1',
      username: 'admin',
      language: 'zh-CN',
    });

    const { restoreAuthenticatedSession } = await import('../src/auth/bootstrap');
    const localeStore = useLocaleStore();
    const sessionStore = useSessionStore();

    sessionStore.setAccessToken('access-token');
    sessionStore.setLanguage('en-US');
    localeStore.setLanguage('en-US');

    await restoreAuthenticatedSession();

    expect(fetchCurrentUser).toHaveBeenCalledTimes(1);
    expect(sessionStore.language).toBe('en-US');
    expect(localeStore.language).toBe('en-US');
    expect(localStorage.getItem('goadmin.language')).toBe('en-US');
  });

  it('translates keys and resolves route titles with locale fallbacks', async () => {
    const localeStore = useLocaleStore();
    localeStore.setLanguage('en-US');
    await initializeI18n(localeStore.language);

    expect(translate('menu.title', 'Menu management')).toBe('Menu management');
    expect(translate('route.book', 'Book')).toBe('Book management');
    expect(translate('route.codegen_console', 'CodeGen console')).toBe('CodeGen console');
    expect(translate('route.casbin_models', 'Model management')).toBe('Model management');
    expect(translate('route.casbin_rules', 'Policy management')).toBe('Policy management');
    expect(translate('role.title', 'Role management')).toBe('Role management');
    expect(translate('codegen.mode.delete', 'Delete')).toBe('Delete');
    expect(
      resolveRouteLocaleMeta({
        path: '/dashboard',
        fullPath: '/dashboard',
        hash: '',
        query: {},
        params: {},
        name: 'dashboard',
        matched: [],
        redirectedFrom: undefined,
        meta: {
          title: '工作台',
          titleKey: 'route.dashboard',
          titleDefault: 'Dashboard',
          subtitle: 'Frontend Core',
          subtitleKey: 'app.subtitle',
          subtitleDefault: 'Frontend Core',
        },
      } as any).title,
    ).toBe('Dashboard');
  });
});
