const route = {
  path: '/system/casbin/rules',
  name: 'CasbinRules',
  component: () => import('@/views/casbin_rule/index.vue'),
  meta: {
    title: 'rules',
    titleKey: 'route.casbin_rules',
    titleDefault: 'Policy management',
    i18nNamespaces: ['casbin_rule'],
    icon: 'menu',
  },
}

export default route
