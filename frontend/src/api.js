import { useAuthStore } from '@/stores/auth'

const BASE = import.meta.env.VITE_API_URL

const delay = (ms) => new Promise((res) => setTimeout(res, ms))

const MAX_RETRIES = 3
const PER_ATTEMPT_TIMEOUT = 30_000 // ms — a single attempt never hangs forever

export async function api(path, options = {}, retries = MAX_RETRIES) {
  const auth = useAuthStore()
  let lastErr = null

  for (let i = 0; i < retries; i++) {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), PER_ATTEMPT_TIMEOUT)
    try {
      const accessToken = await auth.getAccessToken()
      if (!accessToken) {
        throw { status: 401, error: 'missing session' }
      }

      const res = await fetch(`${BASE}${path}`, {
        ...options,
        signal: controller.signal,
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
          ...options.headers,
        },
      })

      const text = await res.text()
      let body = null
      try {
        body = text ? JSON.parse(text) : null
      } catch {
        body = { message: text } // tolerate non-JSON bodies instead of throwing
      }

      if (!res.ok) {
        const err = { status: res.status, ...(body || {}) }
        // Only Render free-tier cold-start 5xx are worth retrying; 4xx won't change.
        if (res.status >= 502 && res.status <= 504 && i < retries - 1) {
          lastErr = err
          await delay(Math.min(1000 * 2 ** i, 8000))
          continue
        }
        throw err
      }
      return body
    } catch (err) {
      // Retry only transient transport failures (network down / our timeout abort).
      const transient = err instanceof TypeError || err?.name === 'AbortError'
      if (!transient || i === retries - 1) throw err
      lastErr = err
      await delay(Math.min(1000 * 2 ** i, 8000))
    } finally {
      clearTimeout(timer)
    }
  }
  throw lastErr
}
