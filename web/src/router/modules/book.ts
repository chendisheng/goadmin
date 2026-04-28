const route = {
  path: '/books',
  name: 'Book',
  component: () => import('@/views/book/index.vue'),
  meta: {
    title: 'Book Management',
    titleKey: 'route.book',
    titleDefault: 'Book Management',
    icon: 'menu',
  },
}

export default route
