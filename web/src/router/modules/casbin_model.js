const route = {
    path: '/system/casbin/models',
    name: 'CasbinModels',
    component: () => import('@/views/casbin_model/index.vue'),
    meta: {
        title: 'models',
        titleKey: 'route.casbin_models',
        titleDefault: 'Model management',
        icon: 'menu',
    },
};
export default route;
