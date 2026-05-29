import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/nodes',
      name: 'Nodes',
      component: () => import('../views/nodes/List.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/edges',
      name: 'Edges',
      component: () => import('../views/edges/List.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()

  if (to.meta.public) {
    if (authStore.isAuthenticated) {
      next('/')
    } else {
      next()
    }
    return
  }

  if (to.meta.requiresAuth) {
    if (!authStore.isAuthenticated) {
      const hasUser = await authStore.fetchUser()
      if (hasUser) {
        next()
      } else {
        next('/login')
      }
    } else {
      next()
    }
    return
  }

  next()
})

export default router
