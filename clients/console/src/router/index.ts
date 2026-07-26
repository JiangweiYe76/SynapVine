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
    {
      path: '/communities',
      name: 'Communities',
      component: () => import('../views/communities/List.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/papers',
      name: 'Papers',
      component: () => import('../views/papers/List.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/review',
      name: 'Review Queue',
      component: () => import('../views/review/List.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/llm',
      name: 'LLM Providers',
      component: () => import('../views/llm/Providers.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/embedding',
      name: 'Embedding Providers',
      component: () => import('../views/embedding/Providers.vue'),
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
      // On a fresh page load the in-memory access token is empty even
      // when a valid refresh cookie exists. Attempt one silent refresh
      // (the browser attaches the httpOnly cookie automatically); the
      // response carries a new access token + user, repopulating the
      // in-memory session. If the cookie is absent/expired the refresh
      // fails and we redirect to /login.
      const refreshed = await authStore.refresh()
      if (refreshed) {
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
