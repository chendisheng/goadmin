const route = {
    path: '/books',
    name: 'Book',
    component: () => import('@/views/book/index.vue'),
    meta: {
        title: 'Book',
        titleKey: 'route.book',
        titleDefault: 'Book',
        icon: 'menu',
    },
};
export default route;
