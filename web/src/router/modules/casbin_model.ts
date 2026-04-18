const route = {
  path: '/casbin_models',
  name: 'CasbinModel',
  component: () => import('@/views/casbin_model/index.vue'),
  meta: {
    title: 'CasbinModel',
    icon: 'menu',
  },
}

export default route
