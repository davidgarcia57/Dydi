<script setup>
defineProps({
  size: {
    type: String,
    default: 'md', // sm | md | lg | xl
    validator: (v) => ['sm', 'md', 'lg', 'xl'].includes(v),
  },
  showText: {
    type: Boolean,
    default: true,
  },
})

const TEXT_SIZES = {
  sm: 'text-xl',
  md: 'text-2xl',
  lg: 'text-4xl',
  xl: 'text-5xl',
}

const LOGO_SIZES = {
  sm: 'h-6 w-6',
  md: 'h-8 w-8',
  lg: 'h-12 w-12',
  xl: 'h-16 w-16',
}
</script>

<template>
  <div class="group flex cursor-pointer items-center gap-2">
    <!-- SVG Logo -->
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 240 240"
      :class="LOGO_SIZES[size]"
      class="shrink-0 transition-transform duration-1000 ease-out group-hover:rotate-[360deg]"
    >
      <g transform="translate(120, 120)">
        <g>
          <!-- Animación nativa de rotación (SMIL) libre de bugs de transform-origin -->
          <animateTransform
            attributeName="transform"
            type="rotate"
            from="0"
            to="360"
            dur="30s"
            repeatCount="indefinite"
          />
          <!-- Segmento 3: Hairline -->
          <circle
            cx="0"
            cy="0"
            r="70"
            class="fill-none stroke-hairline"
            stroke-width="20"
            stroke-linecap="round"
            stroke-dasharray="115 325"
            transform="rotate(240)"
          />
          <!-- Segmento 1: Terracotta -->
          <circle
            cx="0"
            cy="0"
            r="70"
            class="fill-none stroke-terracotta"
            stroke-width="20"
            stroke-linecap="round"
            stroke-dasharray="115 325"
            transform="rotate(0)"
          />
          <!-- Segmento 2: Sage-deep (Desplazado - Consecuencia) -->
          <g transform="rotate(120) translate(0, 12)">
            <circle
              cx="0"
              cy="0"
              r="70"
              class="fill-none stroke-sage-deep"
              stroke-width="20"
              stroke-linecap="round"
              stroke-dasharray="115 325"
            />
          </g>
          <!-- Bolas de ruleta asimétricas (Movimiento) -->
          <g class="animate-ball-float">
            <circle cx="82" cy="-40" r="7" class="fill-ink" />
            <circle cx="94" cy="-22" r="4" class="fill-ink-soft" />
          </g>
        </g>
      </g>
    </svg>

    <span
      v-if="showText"
      class="font-serif font-semibold lowercase tracking-tight text-terracotta transition-colors duration-300 group-hover:text-accent-deep"
      :class="TEXT_SIZES[size]"
    >
      dydi
    </span>
  </div>
</template>

<style scoped>
@keyframes float-ball {
  0%,
  100% {
    transform: translate(0, 0);
  }
  50% {
    transform: translate(3px, -3px);
  }
}

.animate-ball-float {
  animation: float-ball 2s ease-in-out infinite;
  transform-origin: center;
}
</style>
