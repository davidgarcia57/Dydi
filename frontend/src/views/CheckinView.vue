<script setup>
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<template>
  <div class="min-h-screen bg-cream flex flex-col items-center justify-center px-6">
    <button
      class="absolute top-4 left-4 text-ink-soft"
      @click="router.back()"
    >
      <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18 18 6M6 6l12 12"/>
      </svg>
    </button>
    <p class="text-eyebrow mb-2">PRÓXIMAMENTE</p>
    <h1 class="font-serif text-3xl font-semibold text-ink text-center">Flujo de check-in</h1>
    <p class="text-sm text-ink-soft mt-2 text-center">Esta pantalla se construye en la siguiente iteración.</p>
  </div>
</template>
