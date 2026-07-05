<script setup>
// Diana pequeña — el glifo de "en riesgo de ruleta". currentColor para que el
// padre decida el tono; el anillo exterior pulsa suave. Reemplaza al emoji 🎯.
defineProps({
  size: { type: Number, default: 14 },
})
</script>

<template>
  <svg
    class="tg"
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2.4"
    aria-hidden="true"
  >
    <circle class="ring" cx="12" cy="12" r="9.5" />
    <circle cx="12" cy="12" r="5" />
    <circle cx="12" cy="12" r="1.6" fill="currentColor" stroke="none" />
  </svg>
</template>

<style scoped>
.tg {
  display: inline-block;
  flex-shrink: 0;
}
.ring {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: tg-pulse 2.2s ease-in-out infinite;
}
@keyframes tg-pulse {
  0%,
  100% {
    transform: scale(0.9);
    opacity: 0.6;
  }
  50% {
    transform: scale(1);
    opacity: 1;
  }
}
@media (prefers-reduced-motion: reduce) {
  .ring {
    animation: none !important;
  }
}
</style>
