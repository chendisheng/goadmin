const route = {
    path: '/system/casbin/models',
    name: 'CasbinModels',
    component: () => import('@/views/casbin_model/index.vue'),
    meta: {
        title: 'models',
        icon: 'menu',
    },
};
export default route;
