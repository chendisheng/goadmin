const route = {
  path: '/casbin_rules',
  name: 'CasbinRule',
  component: () => import('@/views/casbin_rule/index.vue'),
  meta: {
    title: 'CasbinRule',
    icon: 'menu',
  },
}

export default route
