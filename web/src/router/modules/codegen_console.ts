const route = {
  path: '/codegen_consoles',
  name: 'CodegenConsole',
  component: () => import('@/views/codegen_console/index.vue'),
  meta: {
    title: 'CodegenConsole',
    icon: 'menu',
  },
}

export default route
