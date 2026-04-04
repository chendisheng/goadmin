const route = {
  path: '/orders',
  name: 'Order',
  component: () => import('@/views/order/index.vue'),
  meta: {
    title: 'Order',
    icon: 'menu',
  },
}

export default route
