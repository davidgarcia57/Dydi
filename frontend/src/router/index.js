import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { readPendingInvite } from '@/pendingInvite'

const routes = [
  {
    path: '/login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true },
  },
  // Pública a propósito: el caso que importa es que te inviten sin tener cuenta.
  // JoinView guarda la invitación y rebota al login; el guard de abajo la
  // recupera en cuanto hay sesión.
  {
    path: '/join',
    component: () => import('@/views/JoinView.vue'),
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
  // Ya logueado → rebotar de /login. Si quedó una invitación pendiente (llegó por
  // enlace sin sesión y tuvo que registrarse), se retoma en vez de caer en /today
  // y dejarlo fuera del squad al que lo invitaron.
  if (to.path === '/login' && auth.isLoggedIn) {
    const pending = readPendingInvite()
    return pending ? { path: '/join', query: { g: pending.g, c: pending.c } } : '/today'
  }
})

export default router
