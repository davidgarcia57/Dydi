import { useAuthStore } from '@/stores/auth'

const BASE = import.meta.env.VITE_API_URL

export async function api(path, options = {}) {
  const auth = useAuthStore()
  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${auth.token}`,
      ...options.headers,
    },
  })
  if (!res.ok) throw await res.json()
  return res.json()
}
