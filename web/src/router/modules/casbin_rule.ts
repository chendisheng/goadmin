const route = {
  path: '/system/casbin/rules',
  name: 'CasbinRules',
  component: () => import('@/views/casbin_rule/index.vue'),
  meta: {
    title: 'rules',
    icon: 'menu',
  },
}

export default route
