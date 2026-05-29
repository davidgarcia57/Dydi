import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

export function useAuth() {
  const store = useAuthStore()
  const router = useRouter()

  async function logout() {
    await store.logout()
    router.push('/login')
  }

  return {
    user: store.user,
    isLoggedIn: store.isLoggedIn,
    login: store.login,
    register: store.register,
    logout,
  }
}
