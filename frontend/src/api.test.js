import { describe, it, expect, vi, beforeEach } from 'vitest'
import { api } from './api'
import { setActivePinia, createPinia } from 'pinia'

// Mockear el store de Pinia para no depender del navegador
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ token: 'fake-token' })
}))

// Mockear variables de entorno de Vite
vi.stubEnv('VITE_API_URL', 'http://localhost:8080')

// Mockear el fetch global
global.fetch = vi.fn()

describe('api.js (Cold Start Resilience)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('debe reintentar la petición si Render devuelve 502/503 y eventualmente tener éxito', async () => {
    // Simulamos que las 2 primeras veces Render devuelve error de Gateway
    global.fetch
      .mockResolvedValueOnce({
        ok: false,
        status: 502,
        text: async () => JSON.stringify({ error: 'Bad Gateway' })
      })
      .mockResolvedValueOnce({
        ok: false,
        status: 503,
        text: async () => JSON.stringify({ error: 'Service Unavailable' })
      })
      // A la 3ra vez, los servidores ya despertaron y responde bien
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        text: async () => JSON.stringify({ data: 'ok' })
      })

    // Ejecutamos la petición reduciendo los retries y sin esperar los tiempos reales de delay 
    // (vitest maneja los promesas de delay rápidamente si mockeamos timers, o aquí como usamos timeouts reales tomará ~3 segundos)
    const result = await api('/test-endpoint')
    
    expect(result).toEqual({ data: 'ok' })
    expect(global.fetch).toHaveBeenCalledTimes(3)
  }, 10000) // Timeout de 10s porque nuestro delay real esperará 1s + 2s = 3s

  it('NO debe reintentar si es un error del usuario (ej. 400 Bad Request)', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 400,
      text: async () => JSON.stringify({ error: 'Bad request' })
    })

    await expect(api('/fail-endpoint')).rejects.toEqual({ error: 'Bad request' })
    expect(global.fetch).toHaveBeenCalledTimes(1) // Solo 1 intento
  })
})
