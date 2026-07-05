<script setup>
// Destello de cuatro puntas que titila — el glifo de "semana perfecta".
// currentColor para heredar el tono del padre. Reemplaza al emoji ✨.
defineProps({
  size: { type: Number, default: 12 },
})
</script>

<template>
  <svg
    class="sg"
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="currentColor"
    aria-hidden="true"
  >
    <path class="big" d="M12 2 L14.2 9.8 L22 12 L14.2 14.2 L12 22 L9.8 14.2 L2 12 L9.8 9.8 Z" />
    <path
      class="small"
      d="M19 3 L19.9 5.6 L22.5 6.5 L19.9 7.4 L19 10 L18.1 7.4 L15.5 6.5 L18.1 5.6 Z"
    />
  </svg>
</template>

<style scoped>
.sg {
  display: inline-block;
  flex-shrink: 0;
  vertical-align: -0.1em;
}
.big,
.small {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: sg-twinkle 2.4s ease-in-out infinite;
}
.small {
  animation-delay: 1.2s;
}
@keyframes sg-twinkle {
  0%,
  100% {
    transform: scale(0.7);
    opacity: 0.45;
  }
  50% {
    transform: scale(1);
    opacity: 1;
  }
}
@media (prefers-reduced-motion: reduce) {
  .big,
  .small {
    animation: none !important;
  }
}
</style>
