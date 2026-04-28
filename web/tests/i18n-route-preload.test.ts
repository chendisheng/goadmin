// @vitest-environment jsdom
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

const preloadRouteNamespaces = vi.fn().mockResolvedValue(undefined);

vi.mock('@/i18n', () => ({
  preloadRouteNamespaces,
  resolveRouteLocaleMeta: () => ({ title: 'Login', subtitle: '' }),
  useAppI18n: () => ({
    t: (key: string, fallback = '') => fallback || key,
  }),
}));

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
  window.scrollTo = vi.fn();
  setActivePinia(createPinia());
});

describe('i18n route namespace preloading', () => {
  it('calls preloadRouteNamespaces in router.beforeEach', async () => {
    const { default: router } = await import('../src/router');

    await router.push('/login');

    expect(preloadRouteNamespaces).toHaveBeenCalled();
  });

  it('keeps route namespace preloading after backend routes are registered', async () => {
    const { default: router } = await import('../src/router');

    await router.push('/system/dictionary/categories');

    expect(preloadRouteNamespaces).toHaveBeenCalled();
  });

  it('preloads explicit and component-derived namespaces for backend routes', async () => {
    const { preloadRouteNamespaces, initializeI18n, setI18nLanguage, useAppI18n } = await import('../src/i18n');
    const localeStore = (await import('../src/store/locale')).useLocaleStore();
    localeStore.setLanguage('zh-CN');
    await initializeI18n(localeStore.language);

    await preloadRouteNamespaces({
      meta: {
        title: 'Dictionary categories',
        titleKey: 'dictionary.category.title',
        titleDefault: 'Dictionary categories',
        componentName: 'view/system/dictionary/category/index',
        i18nNamespaces: ['dictionary'],
      },
    } as any);

    const { t } = useAppI18n();
    expect(t('dictionary.category.title', 'Dictionary categories')).toBe('字典分类管理');

    await preloadRouteNamespaces({
      meta: {
        title: 'Dictionary categories',
        titleKey: 'dictionary.category.title',
        titleDefault: 'Dictionary categories',
        componentName: 'view/system/dictionary/category/index',
        i18nNamespaces: ['dictionary'],
      },
    } as any, 'en-US');
    await setI18nLanguage('en-US');
    expect(t('dictionary.category.title', 'Dictionary categories')).toBe('Dictionary categories');
  });

  it('falls back to route name and path when componentName is missing', async () => {
    const { preloadRouteNamespaces, initializeI18n, useAppI18n } = await import('../src/i18n');
    const localeStore = (await import('../src/store/locale')).useLocaleStore();
    localeStore.setLanguage('zh-CN');
    await initializeI18n(localeStore.language);

    await preloadRouteNamespaces({
      path: '/system/dictionary/categories',
      fullPath: '/system/dictionary/categories',
      hash: '',
      query: {},
      params: {},
      name: 'systemDictionaryCategories',
      matched: [],
      redirectedFrom: undefined,
      meta: {
        title: 'Dictionary categories',
        titleKey: 'dictionary.category.title',
        titleDefault: 'Dictionary categories',
      },
    } as any);

    const { t } = useAppI18n();
    expect(t('dictionary.category.title', 'Dictionary categories')).toBe('字典分类管理');
  });
});
