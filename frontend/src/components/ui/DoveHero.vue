<script setup>
// La paloma del perdón: cuerpo de papel con contorno de tinta, ala que aletea
// suave y una ramita de olivo. Mismo espíritu que WaterBottle — escena física,
// animación ambiental, sin emojis.
defineProps({
  size: { type: Number, default: 72 },
})
</script>

<template>
  <svg
    class="dove"
    :width="size"
    viewBox="0 0 200 160"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
  >
    <title>Paloma del perdón</title>

    <g class="fly">
      <!-- Cola -->
      <path
        d="M42 96 L14 112 L34 88 L18 78 Z"
        fill="#FCF9F3"
        stroke="var(--color-ink)"
        stroke-width="4"
        stroke-linejoin="round"
      />
      <!-- Cuerpo -->
      <path
        d="M36 92 Q54 62 92 60 Q130 58 148 74 Q160 84 156 94 Q150 108 122 110 Q76 114 52 104 Q40 100 36 92 Z"
        fill="#FCF9F3"
        stroke="var(--color-ink)"
        stroke-width="4"
        stroke-linejoin="round"
      />
      <!-- Cabeza + pico -->
      <circle cx="150" cy="76" r="15" fill="#FCF9F3" stroke="var(--color-ink)" stroke-width="4" />
      <circle cx="155" cy="73" r="2.4" fill="var(--color-ink)" />
      <path d="M164 78 L178 82 L164 87 Z" fill="var(--color-terracotta)" />
      <!-- Ramita de olivo -->
      <g class="branch">
        <path
          d="M176 84 Q186 96 182 112"
          fill="none"
          stroke="#5C7650"
          stroke-width="3.5"
          stroke-linecap="round"
        />
        <path d="M181 94 Q190 92 194 96 Q188 102 181 99 Z" fill="#A8C39A" />
        <path d="M182 106 Q191 106 194 111 Q187 116 181 112 Z" fill="#A8C39A" />
      </g>
      <!-- Ala trasera (quieta, da profundidad) -->
      <path
        d="M84 88 Q104 92 122 88 Q112 102 92 102 Q84 96 84 88 Z"
        fill="#EFE7D8"
        stroke="var(--color-ink)"
        stroke-width="3"
        stroke-linejoin="round"
      />
      <!-- Ala que aletea -->
      <path
        class="wing"
        d="M88 84 Q74 46 96 26 Q100 52 118 66 Q106 82 88 84 Z"
        fill="#FCF9F3"
        stroke="var(--color-ink)"
        stroke-width="4"
        stroke-linejoin="round"
      />
    </g>
  </svg>
</template>

<style scoped>
.dove {
  user-select: none;
}
.fly {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: dv-bob 3s ease-in-out infinite;
}
@keyframes dv-bob {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-5px);
  }
}
.wing {
  transform-box: fill-box;
  transform-origin: 50% 100%;
  animation: dv-flap 1.5s ease-in-out infinite;
}
@keyframes dv-flap {
  0%,
  100% {
    transform: rotate(0);
  }
  50% {
    transform: rotate(-14deg) translateY(3px);
  }
}
.branch {
  transform-box: fill-box;
  transform-origin: 0% 0%;
  animation: dv-sway 3s ease-in-out infinite;
}
@keyframes dv-sway {
  0%,
  100% {
    transform: rotate(0);
  }
  50% {
    transform: rotate(4deg);
  }
}
@media (prefers-reduced-motion: reduce) {
  .dove * {
    animation: none !important;
  }
}
</style>
