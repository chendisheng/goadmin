const route = {
    path: '/system/codegen',
    name: 'codegen-console',
    component: () => import('@/views/system/codegen/index.vue'),
    meta: {
        title: 'Codegen console',
        titleKey: 'route.codegen_console',
        titleDefault: 'CodeGen console',
        permission: 'codegen:list',
    },
};
export default route;
