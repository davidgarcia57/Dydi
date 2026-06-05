import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/onboarding',
    component: () => import('@/views/OnboardingView.vue'),
    meta: { onboarding: true },
  },
  {
    path: '/',
    redirect: '/today',
  },
  {
    path: '/today',
    component: () => import('@/views/TodayView.vue'),
  },
  {
    path: '/squad',
    component: () => import('@/views/SquadView.vue'),
  },
  {
    path: '/propuestas',
    component: () => import('@/views/ProposalsView.vue'),
  },
  {
    path: '/ruleta',
    component: () => import('@/views/TrialView.vue'),
  },
  {
    path: '/tu',
    component: () => import('@/views/ShameView.vue'),
  },
  // checkin flow — modal screen, no tab
  {
    path: '/checkin',
    component: () => import('@/views/CheckinView.vue'),
    meta: { checkinFlow: true },
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLoggedIn) {
    return '/login'
  }
  // Ya logueado → rebotar de /login a /today
  if (to.path === '/login' && auth.isLoggedIn) {
    return '/today'
  }
})

export default router
