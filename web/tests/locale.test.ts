// @vitest-environment jsdom
import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

import { initializeI18n, resolveRouteLocaleMeta, translate } from '../src/i18n';
import { useLocaleStore } from '../src/store/locale';
import { useSessionStore } from '../src/store/session';

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
  setActivePinia(createPinia());
});

describe('useLocaleStore', () => {
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
