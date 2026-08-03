<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const showWakeup = ref(false)
const wokeUp = ref(false)
const auth = useAuthStore()

// Render free duerme un servicio a los 15 min y su cold start es de ~13 s, y
// mientras arranca el edge de Render contesta 502 al instante. Esta pantalla
// tiene que sostenerse hasta que el backend conteste de verdad; si se quita
// antes, el dashboard monta directo contra una pared de 502.
const WAKE_DEADLINE = 60_000 // ms — coincide con el "hasta un minuto" del copy
const POLL_INTERVAL = 2000 // ms — 502 vuelve rápido, así que esto es el ritmo real

// Cualquier respuesta que no sea 5xx prueba que el servicio ya está procesando
// (un 401 también lo prueba, y no queremos reintentar por un problema de auth).
async function isAwake(url, headers) {
  try {
    const res = await fetch(url, headers ? { headers } : undefined)
    return res.ok || res.status < 500
  } catch {
    return false // red caída o todavía arrancando
  }
}

onMounted(async () => {
  // Damos 1.5 segundos de gracia; si responde rápido, no mostramos el loader
  const timeoutId = setTimeout(() => {
    if (!wokeUp.value) showWakeup.value = true
  }, 1500)

  try {
    const BASE = import.meta.env.VITE_API_URL
    const deadline = Date.now() + WAKE_DEADLINE

    for (;;) {
      const token = auth.session?.access_token
      // Sin sesión solo podemos despertar al gateway: su middleware de Auth
      // rechaza /api/* antes de proxear, así que una sonda anónima nunca llega
      // a habits. Con sesión despertamos groups y habits a la vez — el
      // dashboard necesita los dos, y así arrancan en paralelo en vez de en
      // serie cuando el usuario ya está viendo la pantalla.
      const probes = token
        ? [
            isAwake(`${BASE}/api/groups`, { Authorization: `Bearer ${token}` }),
            isAwake(`${BASE}/api/habits`, { Authorization: `Bearer ${token}` }),
          ]
        : [isAwake(`${BASE}/health`)]

      if ((await Promise.all(probes)).every(Boolean)) break
      if (Date.now() >= deadline) break // nunca dejar al usuario atrapado aquí
      await new Promise((r) => setTimeout(r, POLL_INTERVAL))
    }

    wokeUp.value = true
    setTimeout(() => {
      showWakeup.value = false
    }, 500) // ligero delay visual
  } finally {
    clearTimeout(timeoutId)
  }
})
</script>

<template>
  <Transition name="fade">
    <div
      v-if="showWakeup"
      class="fixed inset-0 z-[100] bg-cream flex flex-col items-center justify-center p-6 text-center"
    >
      <div class="mb-8 relative">
        <div
          class="w-16 h-16 rounded-full border-4 border-hairline border-t-sage-deep animate-spin"
        />
        <div class="absolute inset-0 flex items-center justify-center">
          <span class="w-3 h-3 bg-terracotta rounded-full animate-pulse" />
        </div>
      </div>

      <h2 class="font-serif text-2xl font-semibold text-ink mb-3">Despertando a tu squad…</h2>
      <p class="text-sm text-ink-soft max-w-xs mx-auto">
        El servidor estaba dormido y está arrancando. Esto puede tardar hasta un minuto.
      </p>
    </div>
  </Transition>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.5s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
