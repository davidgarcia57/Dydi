<script setup>
import { ref } from 'vue'

// Ilustración del hábito "tomar agua": el agua ondula sola y chapotea al
// interactuar. Colores de agua hardcodeados (escena física); tapa y contorno
// usan tokens de marca. Es el único hábito con gráfico propio por ahora.
defineProps({
  size: { type: Number, default: 160 },
})

// id de clip único por instancia (los ids SVG son globales en el documento).
const clipId = `wb-clip-${Math.random().toString(36).slice(2, 9)}`

const sloshing = ref(false)
function slosh() {
  sloshing.value = false
  requestAnimationFrame(() => {
    sloshing.value = true
  })
}
</script>

<template>
  <svg
    class="bottle"
    :width="size"
    viewBox="0 0 200 360"
    role="img"
    xmlns="http://www.w3.org/2000/svg"
    @click="slosh"
  >
    <title>Botella de agua</title>
    <desc>El agua ondula y chapotea al tocarla.</desc>
    <defs>
      <clipPath :id="clipId">
        <path
          d="M84 54 L84 73 Q84 80 79 84 Q59 94 59 118 L59 290 Q59 315 84 315 L116 315 Q141 315 141 290 L141 118 Q141 94 121 84 Q116 80 116 73 L116 54 Z"
        />
      </clipPath>
    </defs>

    <g :clip-path="`url(#${clipId})`">
      <g class="water" :class="{ sloshing }" @animationend="sloshing = false">
        <rect x="-60" y="184" width="320" height="200" fill="#4F9FB0" />
        <path
          class="wave w2"
          d="M-100 184 q25 8 50 0 t50 0 t50 0 t50 0 t50 0 t50 0 t50 0 t50 0 L300 360 L-100 360 Z"
          fill="#4F9FB0"
        />
        <path
          class="wave w1"
          d="M-100 178 q25 -9 50 0 t50 0 t50 0 t50 0 t50 0 t50 0 t50 0 t50 0 L300 360 L-100 360 Z"
          fill="#7FC4D1"
        />
        <circle class="b b1" cx="86" cy="300" r="3.5" fill="#ffffff" opacity="0.5" />
        <circle class="b b2" cx="108" cy="305" r="2.5" fill="#ffffff" opacity="0.5" />
        <circle class="b b3" cx="98" cy="298" r="3" fill="#ffffff" opacity="0.5" />
      </g>
    </g>

    <rect x="66" y="120" width="7" height="150" rx="3.5" fill="#ffffff" opacity="0.22" />
    <path
      class="outline"
      d="M80 50 L80 72 Q80 79 74 83 Q54 93 54 117 L54 292 Q54 320 82 320 L118 320 Q146 320 146 292 L146 117 Q146 93 126 83 Q120 79 120 72 L120 50 Z"
      fill="none"
      stroke-width="4"
      stroke-linejoin="round"
    />
    <rect class="cap" x="77" y="20" width="46" height="30" rx="7" />
    <rect class="cap" x="82" y="44" width="36" height="9" rx="3" />
  </svg>
</template>

<style scoped>
.bottle {
  cursor: pointer;
  user-select: none;
}
.outline {
  stroke: var(--color-ink);
}
.cap {
  fill: var(--color-terracotta);
}
.wave {
  transform-box: fill-box;
  transform-origin: 50% 50%;
}
.w1 {
  animation: wb-wmove 3s linear infinite;
}
.w2 {
  animation: wb-wmove 4.6s linear infinite;
  opacity: 0.5;
}
.bottle:hover .w1 {
  animation-duration: 1.3s;
}
.bottle:hover .w2 {
  animation-duration: 2s;
}
.water {
  transform-box: fill-box;
  transform-origin: 50% 50%;
}
.water.sloshing {
  animation: wb-slosh 1.1s ease-out;
}
.b {
  transform-box: fill-box;
  animation: wb-rise 4.2s ease-in infinite;
  opacity: 0;
}
.b2 {
  animation-delay: 1.4s;
}
.b3 {
  animation-delay: 2.8s;
}
@keyframes wb-wmove {
  to {
    transform: translateX(-100px);
  }
}
@keyframes wb-slosh {
  0% {
    transform: rotate(0);
  }
  22% {
    transform: rotate(5deg);
  }
  48% {
    transform: rotate(-4deg);
  }
  72% {
    transform: rotate(2.5deg);
  }
  100% {
    transform: rotate(0);
  }
}
@keyframes wb-rise {
  0% {
    transform: translateY(0);
    opacity: 0;
  }
  15% {
    opacity: 0.7;
  }
  100% {
    transform: translateY(-95px);
    opacity: 0;
  }
}
</style>
