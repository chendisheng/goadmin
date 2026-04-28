// @vitest-environment jsdom
import { beforeEach, describe, expect, it } from 'vitest';
import { computed } from 'vue';
import { createPinia, setActivePinia } from 'pinia';

import { initializeI18n, preloadRouteNamespaces, resolveRouteLocaleMeta, setI18nLanguage, useAppI18n } from '../src/i18n';
import { useLocaleStore } from '../src/store/locale';

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
  setActivePinia(createPinia());
});

describe('i18n language switching reactivity', () => {
  it('refreshes translated UI labels after the language changes', async () => {
    const localeStore = useLocaleStore();
    localeStore.setLanguage('zh-CN');

    await initializeI18n(localeStore.language);

    const { t } = useAppI18n();
    const sidebarTitle = computed(() => t('common.expand_sidebar', 'Expand sidebar'));

    expect(sidebarTitle.value).toBe('展开侧栏');

    localeStore.setLanguage('en-US');
    await setI18nLanguage(localeStore.language);

    expect(sidebarTitle.value).toBe('Expand sidebar');

    const routeTitle = computed(() =>
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
          subtitle: '前端核心',
          subtitleKey: 'app.subtitle',
          subtitleDefault: 'Frontend Core',
        },
      } as any).title,
    );

    expect(routeTitle.value).toBe('Dashboard');
  });

  it('updates computed labels after a route namespace is preloaded for the active page', async () => {
    const localeStore = useLocaleStore();
    localeStore.setLanguage('en-US');

    await initializeI18n(localeStore.language);

    const { t } = useAppI18n();
    const categoryTitle = computed(() => t('dictionary.category.title', 'Dictionary categories'));

    expect(categoryTitle.value).toBe('Dictionary categories');

    await preloadRouteNamespaces({
      meta: {
        title: 'Dictionary categories',
        titleKey: 'dictionary.category.title',
        titleDefault: 'Dictionary categories',
        componentName: 'view/system/dictionary/category/index',
        i18nNamespaces: ['dictionary'],
      },
    } as any, 'zh-CN');
    await setI18nLanguage('zh-CN');

    expect(categoryTitle.value).toBe('字典分类管理');
  });
});
