// @vitest-environment jsdom
import { describe, expect, it } from 'vitest';

import { ensureNamespaces, getLoadedNamespaces, hasLoadedNamespace, loadNamespace } from '../src/i18n/loader';
import { I18N_DEFAULT_NS } from '../src/i18n/language';
import { collectNamespacesFromRouteMeta, namespaceFromKey } from '../src/i18n/namespaces';

describe('i18n task-3 namespace utilities', () => {
  it('resolves namespace from translation keys', () => {
    expect(namespaceFromKey('menu.title')).toBe('menu');
    expect(namespaceFromKey('route.dashboard')).toBe('route');
    expect(namespaceFromKey('')).toBe(I18N_DEFAULT_NS);
    expect(namespaceFromKey('invalid-key-without-dot')).toBe(I18N_DEFAULT_NS);
  });

  it('collects route namespaces from route meta keys', () => {
    const namespaces = collectNamespacesFromRouteMeta({
      titleKey: 'menu.title',
      subtitleKey: 'route.placeholder.subtitle',
      title: 'route.dashboard',
    } as any);

    expect(namespaces).toEqual(expect.arrayContaining([I18N_DEFAULT_NS, 'common', 'route', 'menu']));
  });
});

describe('i18n task-3 namespace loader', () => {
  it('loads migrated app/common/route keys from locale files', async () => {
    const app = await loadNamespace('zh-CN', 'app');
    const common = await loadNamespace('zh-CN', 'common');
    const route = await loadNamespace('en-US', 'route');

    expect(app['app.title']).toBe('GoAdmin');
    expect(common['common.refresh']).toBe('刷新');
    expect(route['route.casbin_overview']).toBe('Overview');
  });

  it('loads migrated menu/role/user/login/tabs keys from locale files', async () => {
    const menu = await loadNamespace('en-US', 'menu');
    const role = await loadNamespace('en-US', 'role');
    const user = await loadNamespace('zh-CN', 'user');
    const login = await loadNamespace('zh-CN', 'login');
    const tabs = await loadNamespace('zh-CN', 'tabs');

    expect(menu['menu.description']).toBeTruthy();
    expect(role['role.description']).toBeTruthy();
    expect(user['user.username']).toBe('用户名');
    expect(login['login.welcome']).toBe('欢迎使用 GoAdmin');
    expect(tabs['tabs.title']).toBe('页面');
  });

  it('loads migrated business namespaces from locale files', async () => {
    const plugin = await loadNamespace('zh-CN', 'plugin');
    const dictionary = await loadNamespace('zh-CN', 'dictionary');
    const upload = await loadNamespace('en-US', 'upload');
    const casbinRule = await loadNamespace('en-US', 'casbin_rule');
    const book = await loadNamespace('zh-CN', 'book');
    const dashboard = await loadNamespace('en-US', 'dashboard');

    expect(plugin['plugin.title']).toBe('插件中心');
    expect(dictionary['dictionary.category.title']).toBe('字典分类管理');
    expect(upload['upload.preview.public_direct']).toBe('Public direct link');
    expect(casbinRule['casbin_rule.title']).toBe('Policy management');
    expect(book['book.title']).toBe('图书管理');
    expect(dashboard['dashboard.hero_title']).toBe('System overview');
  });

  it('loads and caches namespace resources', async () => {
    const menuResource = await loadNamespace('zh-CN', 'menu');

    expect(menuResource['menu.title']).toBeTruthy();
    expect(hasLoadedNamespace('zh-CN', 'menu')).toBe(true);
    expect(getLoadedNamespaces('zh-CN')).toEqual(expect.arrayContaining(['menu']));
  });

  it('ensures multiple namespaces in one call', async () => {
    const bundle = await ensureNamespaces('en-US', ['menu', 'route']);

    expect(Object.keys(bundle)).toEqual(expect.arrayContaining(['menu', 'route']));
    expect(bundle.menu['menu.title']).toBeTruthy();
    expect(bundle.route['route.dashboard']).toBeTruthy();
  });

  it('returns empty resource for blank namespace', async () => {
    const resource = await loadNamespace('zh-CN', '   ');
    expect(resource).toEqual({});
  });
});
