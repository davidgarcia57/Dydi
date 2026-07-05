<script setup>
import { computed } from 'vue'

// La rueda de penitencias, dibujada: aro exterior de tinta, rim crema, pines
// entre segmentos, hub con tapa terracotta y puntero integrado. El padre
// controla la rotación (deg + spinning), igual que con el SVG anterior; aquí
// solo vive el aspecto. `sleeping` es la variante del empty state: la rueda
// respira desaturada mientras sueltan zetas.
const props = defineProps({
  count: { type: Number, required: true },
  colors: { type: Array, required: true },
  size: { type: Number, default: 220 },
  deg: { type: Number, default: 0 },
  spinning: { type: Boolean, default: false },
  sleeping: { type: Boolean, default: false },
})

function polar(cx, cy, r, degrees) {
  const rad = ((degrees - 90) * Math.PI) / 180
  return [cx + r * Math.cos(rad), cy + r * Math.sin(rad)]
}

function segmentPath(cx, cy, r, start, end) {
  const [sx, sy] = polar(cx, cy, r, start)
  const [ex, ey] = polar(cx, cy, r, end)
  const large = end - start > 180 ? 1 : 0
  return `M ${cx} ${cy} L ${sx} ${sy} A ${r} ${r} 0 ${large} 1 ${ex} ${ey} Z`
}

const CX = 110
const CY = 134

const segments = computed(() => {
  const n = Math.max(2, props.count)
  const angle = 360 / n
  return Array.from({ length: n }, (_, i) => ({
    path: segmentPath(CX, CY, 88, i * angle, (i + 1) * angle),
    color: props.colors[i % props.colors.length],
  }))
})

const pegs = computed(() => {
  const n = Math.max(2, props.count)
  const angle = 360 / n
  return Array.from({ length: n }, (_, i) => {
    const [x, y] = polar(CX, CY, 93, i * angle)
    return { x, y }
  })
})

// La rotación real (spin) va como estilo inline sobre el grupo exterior; las
// animaciones idle/sleeping viven en el grupo interior para no pelearse.
const wheelStyle = computed(() => ({
  transform: `rotate(${props.deg}deg)`,
  transition: props.spinning ? 'transform 4.2s cubic-bezier(0.17, 0.67, 0.12, 0.99)' : 'none',
  transformOrigin: `${CX}px ${CY}px`,
}))
</script>

<template>
  <svg
    class="rw"
    :class="{ sleeping, spinning }"
    :width="size"
    viewBox="0 0 220 250"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
  >
    <title>Ruleta de penitencias</title>

    <ellipse cx="110" cy="240" rx="62" ry="7" fill="var(--color-ink)" opacity="0.08" />

    <g :style="wheelStyle">
      <g class="disc">
        <circle :cx="CX" :cy="CY" r="102" fill="var(--color-ink)" />
        <circle :cx="CX" :cy="CY" r="97" fill="#FCF9F3" />
        <path v-for="(seg, i) in segments" :key="i" :d="seg.path" :fill="seg.color" />
        <circle
          v-for="(peg, i) in pegs"
          :key="`p${i}`"
          :cx="peg.x"
          :cy="peg.y"
          r="2.6"
          fill="#FCF9F3"
        />
        <circle :cx="CX" :cy="CY" r="26" fill="#FCF9F3" />
        <circle :cx="CX" :cy="CY" r="17" fill="var(--color-terracotta)" />
        <circle :cx="CX" :cy="CY" r="6" fill="var(--color-ink)" />
        <circle :cx="CX - 6" :cy="CY - 7" r="3.5" fill="#FCF9F3" opacity="0.45" />
      </g>
    </g>

    <!-- Puntero (fijo, fuera del grupo que rota) -->
    <g class="pointer">
      <path
        d="M110 46 L96 16 Q110 6 124 16 Z"
        fill="var(--color-ink)"
        stroke="#FCF9F3"
        stroke-width="3"
        stroke-linejoin="round"
      />
    </g>

    <!-- Zetas de la variante dormida -->
    <g
      v-if="sleeping"
      fill="none"
      stroke="#A89C89"
      stroke-width="4"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <path class="z z1" d="M158 70 h13 l-13 13 h13" />
      <path class="z z2" d="M180 44 h9 l-9 9 h9" />
    </g>
  </svg>
</template>

<style scoped>
.rw {
  user-select: none;
}
.disc {
  transform-box: fill-box;
  transform-origin: 50% 50%;
}
/* Invitación sutil mientras nadie gira */
.rw:not(.spinning):not(.sleeping) .disc {
  animation: rw-rock 5s ease-in-out infinite;
}
@keyframes rw-rock {
  0%,
  100% {
    transform: rotate(-1.6deg);
  }
  50% {
    transform: rotate(1.6deg);
  }
}
.rw.spinning .pointer {
  transform-box: fill-box;
  transform-origin: 50% 15%;
  animation: rw-tick 0.14s ease-in-out infinite;
}
@keyframes rw-tick {
  0%,
  100% {
    transform: rotate(0);
  }
  50% {
    transform: rotate(-7deg);
  }
}
.rw.sleeping {
  filter: grayscale(0.55) opacity(0.55);
}
.rw.sleeping .disc {
  animation: rw-breathe 4.6s ease-in-out infinite;
}
@keyframes rw-breathe {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.025);
  }
}
.z {
  animation: rw-zrise 4.2s ease-in-out infinite;
  opacity: 0;
}
.z2 {
  animation-delay: 2.1s;
}
@keyframes rw-zrise {
  0% {
    transform: translateY(8px);
    opacity: 0;
  }
  25% {
    opacity: 0.9;
  }
  70% {
    opacity: 0;
  }
  100% {
    transform: translateY(-18px);
    opacity: 0;
  }
}
@media (prefers-reduced-motion: reduce) {
  .rw .disc,
  .rw .pointer,
  .rw .z {
    animation: none !important;
  }
}
</style>
