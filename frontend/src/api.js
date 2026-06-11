import { useAuthStore } from '@/stores/auth'

const BASE = import.meta.env.VITE_API_URL

const delay = ms => new Promise(res => setTimeout(res, ms))

export async function api(path, options = {}, retries = 5) {
  const auth = useAuthStore()
  let lastErr = null

  for (let i = 0; i < retries; i++) {
    try {
      const res = await fetch(`${BASE}${path}`, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${auth.token}`,
          ...options.headers,
        },
      })
      const text = await res.text()
      const body = text ? JSON.parse(text) : null
      
      // Render free tier puede devolver 502/503/504 durante el cold start
      if (!res.ok) {
        if (res.status >= 502 && res.status <= 504) {
          throw { status: res.status, ...body }
        }
        throw body ?? { message: `HTTP ${res.status}` }
      }
      return body
    } catch (err) {
      lastErr = err
      const isNetworkOr50x = err instanceof TypeError || (err.status >= 502 && err.status <= 504)
      if (!isNetworkOr50x || i === retries - 1) {
        throw err
      }
      // Exponential backoff: 1s, 2s, 4s, 8s, max 10s
      const wait = Math.min(1000 * Math.pow(2, i), 10000)
      await delay(wait)
    }
  }
  throw lastErr
}
