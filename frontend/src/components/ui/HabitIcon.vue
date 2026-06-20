<script setup>
// Iconos de línea animados, uno por hábito del catálogo (icon_key). Mismo
// espíritu que WaterBottle: trazo con currentColor (el padre elige el color),
// movimiento sutil y continuo (solo transform/opacity → GPU), y respeto por
// prefers-reduced-motion. El agua conserva su WaterBottle aparte.
defineProps({
  iconKey: { type: String, default: '' },
  size: { type: Number, default: 40 },
})

const labels = {
  water: 'Agua',
  exercise: 'Ejercicio',
  steps: 'Pasos',
  fruit: 'Fruta o verdura',
  no_sugar: 'Sin comida chatarra',
  read: 'Leer',
  focus: 'Foco sin teléfono',
  journal: 'Journaling',
  no_social: 'Sin redes',
  no_phone: 'Sin teléfono de noche',
  bed: 'Tender la cama',
  outdoors: 'Aire libre',
}
</script>

<template>
  <svg
    class="hi"
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.7"
    stroke-linecap="round"
    stroke-linejoin="round"
    role="img"
    :aria-label="labels[iconKey] || 'Hábito'"
  >
    <title>{{ labels[iconKey] || 'Hábito' }}</title>

    <!-- Ejercicio: mancuerna que hace una repetición lenta -->
    <g v-if="iconKey === 'exercise'" class="ic lift">
      <line x1="8" y1="12" x2="16" y2="12" />
      <rect x="4.5" y="7.5" width="3" height="9" rx="1.2" />
      <rect x="16.5" y="7.5" width="3" height="9" rx="1.2" />
      <line x1="3" y1="9.5" x2="3" y2="14.5" />
      <line x1="21" y1="9.5" x2="21" y2="14.5" />
    </g>

    <!-- Agua: gota que sube/baja con una onda interior -->
    <g v-else-if="iconKey === 'water'">
      <g class="drop">
        <path d="M12 3.5 C12 3.5 6 10.5 6 14.8 A6 6 0 0 0 18 14.8 C18 10.5 12 3.5 12 3.5 Z" />
        <path class="ripple" d="M9.4 15.2 A2.8 2.8 0 0 0 11.6 17.4" />
      </g>
    </g>

    <!-- Pasos: dos huellas que alternan al "caminar" -->
    <g v-else-if="iconKey === 'steps'">
      <g class="foot fl">
        <ellipse cx="8.5" cy="9" rx="2.2" ry="3.2" />
        <circle cx="8.5" cy="13.6" r="1.1" />
      </g>
      <g class="foot fr">
        <ellipse cx="15.5" cy="13" rx="2.2" ry="3.2" />
        <circle cx="15.5" cy="17.6" r="1.1" />
      </g>
    </g>

    <!-- Fruta: manzana con hoja que se mece -->
    <g v-else-if="iconKey === 'fruit'">
      <circle cx="12" cy="13.5" r="5.5" />
      <path class="leaf" d="M12 8 C12.6 5.8 14.6 5.2 16.2 5.4 C15.9 7.4 14.2 8.4 12 8 Z" />
    </g>

    <!-- Sin comida chatarra: vaso de refresco tachado -->
    <g v-else-if="iconKey === 'no_sugar'">
      <path d="M7.5 8 L16.5 8 L15.6 19 Q15.5 20 14.5 20 L9.5 20 Q8.5 20 8.4 19 Z" />
      <line x1="6" y1="8" x2="18" y2="8" />
      <line x1="13" y1="4.2" x2="11.6" y2="8" />
      <line class="slash" x1="5" y1="19" x2="19" y2="5" />
    </g>

    <!-- Leer: libro abierto con una página que aletea -->
    <g v-else-if="iconKey === 'read'">
      <path d="M12 7 C10 5.8 7 5.6 4.5 6.1 L4.5 16.8 C7 16.3 10 16.6 12 17.9" />
      <path class="page" d="M12 7 C14 5.8 17 5.6 19.5 6.1 L19.5 16.8 C17 16.3 14 16.6 12 17.9" />
      <line x1="12" y1="7" x2="12" y2="17.9" />
    </g>

    <!-- Foco: diana con anillo que pulsa -->
    <g v-else-if="iconKey === 'focus'">
      <circle class="pulse" cx="12" cy="12" r="8" />
      <circle cx="12" cy="12" r="4.4" />
      <circle cx="12" cy="12" r="1.4" fill="currentColor" stroke="none" />
    </g>

    <!-- Journaling: libreta con una línea que se "escribe" -->
    <g v-else-if="iconKey === 'journal'">
      <rect x="6" y="4" width="12" height="16" rx="1.6" />
      <line x1="9" y1="8.5" x2="15" y2="8.5" />
      <line x1="9" y1="12" x2="15" y2="12" />
      <line class="write" x1="9" y1="15.5" x2="14" y2="15.5" />
    </g>

    <!-- Sin redes: burbuja de chat tachada -->
    <g v-else-if="iconKey === 'no_social'">
      <path
        d="M5 6 H19 A1.6 1.6 0 0 1 20.5 7.6 V13.4 A1.6 1.6 0 0 1 19 15 H12 L8 18 V15 H5 A1.6 1.6 0 0 1 3.5 13.4 V7.6 A1.6 1.6 0 0 1 5 6 Z"
      />
      <line class="slash" x1="4" y1="19.5" x2="20" y2="4.5" />
    </g>

    <!-- Sin teléfono de noche: teléfono con luna que titila -->
    <g v-else-if="iconKey === 'no_phone'">
      <rect x="7" y="3" width="10" height="18" rx="2.5" />
      <line x1="10.5" y1="18.2" x2="13.5" y2="18.2" />
      <path
        class="moon"
        d="M14.6 9 A3 3 0 1 1 11.4 6 A2.3 2.3 0 0 0 14.6 9 Z"
        fill="currentColor"
        stroke="none"
      />
    </g>

    <!-- Tender la cama: cama con un destello de "ordenado" -->
    <g v-else-if="iconKey === 'bed'">
      <path d="M4 17 V12 A2 2 0 0 1 6 10 H18 A2 2 0 0 1 20 12 V17" />
      <line x1="3" y1="17" x2="21" y2="17" />
      <line x1="3.5" y1="17" x2="3.5" y2="19" />
      <line x1="20.5" y1="17" x2="20.5" y2="19" />
      <path d="M7 10 V8.4 A1.4 1.4 0 0 1 8.4 7 H10.6 A1.4 1.4 0 0 1 12 8.4 V10" />
      <path
        class="sparkle"
        d="M16.6 6 L17.1 7.3 L18.4 7.8 L17.1 8.3 L16.6 9.6 L16.1 8.3 L14.8 7.8 L16.1 7.3 Z"
        fill="currentColor"
        stroke="none"
      />
    </g>

    <!-- Aire libre: sol con rayos que giran lento -->
    <g v-else-if="iconKey === 'outdoors'">
      <circle cx="12" cy="12" r="3.8" />
      <g class="rays">
        <line x1="12" y1="2.5" x2="12" y2="5" />
        <line x1="12" y1="19" x2="12" y2="21.5" />
        <line x1="2.5" y1="12" x2="5" y2="12" />
        <line x1="19" y1="12" x2="21.5" y2="12" />
        <line x1="5.2" y1="5.2" x2="6.9" y2="6.9" />
        <line x1="17.1" y1="17.1" x2="18.8" y2="18.8" />
        <line x1="5.2" y1="18.8" x2="6.9" y2="17.1" />
        <line x1="17.1" y1="6.9" x2="18.8" y2="5.2" />
      </g>
    </g>

    <!-- Fallback: destello que titila -->
    <g v-else class="twinkle">
      <path d="M12 4 L13.3 10.7 L20 12 L13.3 13.3 L12 20 L10.7 13.3 L4 12 L10.7 10.7 Z" />
    </g>
  </svg>
</template>

<style scoped>
.hi {
  display: inline-block;
}
.lift {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hi-lift 2.1s ease-in-out infinite;
}
@keyframes hi-lift {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-2.4px);
  }
}
.foot {
  transform-box: fill-box;
  transform-origin: 50% 50%;
}
.fl {
  animation: hi-step 1.2s ease-in-out infinite;
}
.fr {
  animation: hi-step 1.2s ease-in-out infinite;
  animation-delay: 0.6s;
}
@keyframes hi-step {
  0%,
  100% {
    opacity: 0.3;
    transform: translateY(1px);
  }
  50% {
    opacity: 1;
    transform: translateY(-1px);
  }
}
.drop {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hi-bob 2.6s ease-in-out infinite;
}
@keyframes hi-bob {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-2px);
  }
}
.ripple {
  animation: hi-rip 2.6s ease-in-out infinite;
}
@keyframes hi-rip {
  0%,
  100% {
    opacity: 0.3;
  }
  50% {
    opacity: 0.9;
  }
}
.leaf {
  transform-box: fill-box;
  transform-origin: 0% 100%;
  animation: hi-sway 2.6s ease-in-out infinite;
}
@keyframes hi-sway {
  0%,
  100% {
    transform: rotate(-12deg);
  }
  50% {
    transform: rotate(12deg);
  }
}
.slash {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hi-slash 1.8s ease-in-out infinite;
}
@keyframes hi-slash {
  0%,
  100% {
    opacity: 0.35;
  }
  50% {
    opacity: 1;
  }
}
.page {
  transform-box: fill-box;
  transform-origin: 0% 50%;
  animation: hi-page 3s ease-in-out infinite;
}
@keyframes hi-page {
  0%,
  100% {
    transform: scaleX(1);
  }
  50% {
    transform: scaleX(0.86);
  }
}
.pulse {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hi-pulse 2.2s ease-out infinite;
}
@keyframes hi-pulse {
  0% {
    transform: scale(0.65);
    opacity: 0.75;
  }
  70% {
    opacity: 0;
  }
  100% {
    transform: scale(1.2);
    opacity: 0;
  }
}
.write {
  transform-box: fill-box;
  transform-origin: 0% 50%;
  animation: hi-write 2.4s ease-in-out infinite;
}
@keyframes hi-write {
  0% {
    transform: scaleX(0);
  }
  60%,
  100% {
    transform: scaleX(1);
  }
}
.moon,
.sparkle,
.twinkle {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hi-twinkle 2.4s ease-in-out infinite;
}
@keyframes hi-twinkle {
  0%,
  100% {
    transform: scale(0.7);
    opacity: 0.4;
  }
  50% {
    transform: scale(1);
    opacity: 1;
  }
}
.rays {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hi-spin 12s linear infinite;
}
@keyframes hi-spin {
  to {
    transform: rotate(360deg);
  }
}
@media (prefers-reduced-motion: reduce) {
  .hi * {
    animation: none !important;
  }
}
</style>
