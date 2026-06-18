<script setup>
import { ref, onMounted } from 'vue'

const showWakeup = ref(false)
const wokeUp = ref(false)

onMounted(async () => {
  // Damos 1.5 segundos de gracia; si responde rápido, no mostramos el loader
  const timeoutId = setTimeout(() => {
    if (!wokeUp.value) showWakeup.value = true
  }, 1500)

  try {
    const BASE = import.meta.env.VITE_API_URL

    // Hacemos pings al health endpoint cada 2s hasta que esté arriba
    for (let i = 0; i < 20; i++) {
      try {
        const res = await fetch(`${BASE}/health`)
        if (res.ok) {
          wokeUp.value = true
          break
        }
      } catch (e) {
        // Error de red, el gateway probablemente sigue arrancando
      }
      await new Promise((r) => setTimeout(r, 2000))
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

      <h2 class="font-serif text-2xl font-semibold text-ink mb-3">Despertando los servidores...</h2>
      <p class="text-sm text-ink-soft max-w-xs mx-auto">
        Dado que usamos servidores gratuitos, pueden tardar hasta 40 segundos en despertar tras un
        periodo de inactividad.
      </p>
      <p class="text-xs font-bold text-terracotta mt-4">¡Gracias por tu paciencia! ☕</p>
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
