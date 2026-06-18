<script setup>
import { useToast } from '@/composables/useToast'

const { toast } = useToast()
</script>

<template>
  <div class="fixed inset-x-0 bottom-24 z-[60] flex justify-center px-4 pointer-events-none">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="toast.visible"
        class="bg-ink text-paper rounded-pill px-4 py-2.5 text-sm font-semibold shadow-card"
        role="status"
        aria-live="polite"
      >
        {{ toast.message }}
      </div>
    </Transition>
  </div>
</template>
