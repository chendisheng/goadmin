const route = {
    path: '/system/codegen',
    name: 'codegen-console',
    component: () => import('@/views/system/codegen/index.vue'),
    meta: {
        title: 'Codegen console',
        permission: 'codegen:list',
    },
};
export default route;
