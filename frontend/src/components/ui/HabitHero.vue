<script setup>
import { computed, ref } from 'vue'
import WaterBottle from './WaterBottle.vue'

// Ilustración grande y animada por hábito, en el mismo espíritu que
// WaterBottle: escena "física" con formas rellenas, colores propios de cada
// escena + tokens de marca para contornos, un loop ambiental sutil y una
// reacción al tocarla. El agua conserva su WaterBottle original.
const props = defineProps({
  iconKey: { type: String, default: '' },
  habitName: { type: String, default: '' },
  size: { type: Number, default: 150 },
})

const labels = {
  exercise: 'Ejercicio',
  steps: 'Pasos',
  fruit: 'Fruta o verdura',
  no_sugar: 'Sin comida chatarra',
  read: 'Leer',
  focus: 'Foco sin distracciones',
  journal: 'Journaling',
  no_social: 'Sin redes',
  no_phone: 'Sin teléfono de noche',
  bed: 'Tender la cama',
  outdoors: 'Aire libre',
}

// Mismo hook robusto que usaba CheckinView: icon_key manda, el nombre rescata.
const isWater = computed(
  () => props.iconKey === 'water' || /agua|water/i.test(props.habitName || '')
)

const label = computed(() => labels[props.iconKey] || props.habitName || 'Hábito')

const popping = ref(false)
function pop() {
  popping.value = false
  requestAnimationFrame(() => {
    popping.value = true
  })
}
</script>

<template>
  <WaterBottle v-if="isWater" :size="size" />

  <svg
    v-else
    class="hh"
    :class="{ pop: popping }"
    :width="size"
    viewBox="0 0 200 240"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    @click="pop"
    @animationend="popping = false"
  >
    <title>{{ label }}</title>

    <!-- Sombra de piso compartida -->
    <ellipse cx="100" cy="212" rx="56" ry="8" fill="var(--color-ink)" opacity="0.08" />

    <!-- Ejercicio: barra con discos haciendo repeticiones -->
    <g v-if="iconKey === 'exercise'" class="scene">
      <g class="reps">
        <rect x="34" y="128" width="132" height="7" rx="3.5" fill="var(--color-ink)" />
        <rect x="38" y="114" width="12" height="35" rx="5" fill="#A85B39" />
        <rect x="150" y="114" width="12" height="35" rx="5" fill="#A85B39" />
        <rect x="52" y="106" width="14" height="51" rx="5" fill="#C9714A" />
        <rect x="134" y="106" width="14" height="51" rx="5" fill="#C9714A" />
        <rect x="68" y="122" width="6" height="19" rx="2" fill="var(--color-ink)" />
        <rect x="126" y="122" width="6" height="19" rx="2" fill="var(--color-ink)" />
      </g>
      <g class="effort" stroke="#C9714A" stroke-width="4" stroke-linecap="round">
        <line x1="88" y1="78" x2="84" y2="66" />
        <line x1="100" y1="74" x2="100" y2="62" />
        <line x1="112" y1="78" x2="116" y2="66" />
      </g>
    </g>

    <!-- Pasos: tenis que trota con líneas de velocidad -->
    <g v-else-if="iconKey === 'steps'" class="scene">
      <g class="dashes" stroke="#C9714A" stroke-width="5" stroke-linecap="round" opacity="0.5">
        <line x1="18" y1="150" x2="40" y2="150" />
        <line x1="10" y1="168" x2="36" y2="168" />
        <line x1="20" y1="186" x2="42" y2="186" />
      </g>
      <g class="walk">
        <path
          d="M56 190 L56 156 Q56 146 66 143 Q92 136 108 122 Q116 115 124 119 Q146 132 160 136 Q170 139 170 148 L170 190 Z"
          fill="#C9714A"
        />
        <path
          d="M56 168 Q80 158 104 166 Q136 176 170 170 L170 190 L56 190 Z"
          fill="#FCF9F3"
          opacity="0.35"
        />
        <rect
          x="50"
          y="186"
          width="126"
          height="16"
          rx="8"
          fill="#FCF9F3"
          stroke="var(--color-ink)"
          stroke-width="4"
        />
        <g stroke="#FCF9F3" stroke-width="4" stroke-linecap="round">
          <line x1="104" y1="136" x2="118" y2="146" />
          <line x1="96" y1="146" x2="112" y2="156" />
          <line x1="88" y1="156" x2="104" y2="166" />
        </g>
      </g>
    </g>

    <!-- Fruta: manzana con hoja que se mece y brillo -->
    <g v-else-if="iconKey === 'fruit'" class="scene">
      <circle cx="100" cy="150" r="46" fill="#D96C57" />
      <ellipse
        cx="82"
        cy="132"
        rx="14"
        ry="20"
        fill="#FCF9F3"
        opacity="0.28"
        transform="rotate(-24 82 132)"
      />
      <path
        d="M100 106 q2 -16 12 -21"
        stroke="var(--color-ink)"
        stroke-width="5"
        fill="none"
        stroke-linecap="round"
      />
      <path
        class="leaf"
        d="M104 96 C110 78 128 72 142 76 C138 94 120 100 104 96 Z"
        fill="#A8C39A"
      />
    </g>

    <!-- Sin chatarra: vaso de refresco con señal de prohibido -->
    <g v-else-if="iconKey === 'no_sugar'" class="scene">
      <path d="M112 84 L126 42 L134 45 L120 84 Z" fill="#E9C281" />
      <path
        d="M66 94 L134 94 L126 194 Q125 202 116 202 L84 202 Q75 202 74 194 Z"
        fill="#FCF9F3"
        stroke="var(--color-ink)"
        stroke-width="4"
        stroke-linejoin="round"
      />
      <rect x="60" y="82" width="80" height="13" rx="5" fill="var(--color-terracotta)" />
      <rect x="72" y="124" width="56" height="10" rx="5" fill="#EDA48F" opacity="0.6" />
      <rect x="74" y="152" width="52" height="10" rx="5" fill="#EDA48F" opacity="0.6" />
      <g fill="#7FC4D1" opacity="0.8">
        <circle class="fizz f1" cx="90" cy="182" r="3" />
        <circle class="fizz f2" cx="104" cy="188" r="2.4" />
        <circle class="fizz f3" cx="98" cy="176" r="2.8" />
      </g>
      <g class="ban">
        <circle cx="100" cy="140" r="66" fill="none" stroke="#BC5C42" stroke-width="8" />
        <line
          x1="54"
          y1="187"
          x2="146"
          y2="93"
          stroke="#BC5C42"
          stroke-width="8"
          stroke-linecap="round"
        />
      </g>
    </g>

    <!-- Leer: libro abierto con página que respira -->
    <g v-else-if="iconKey === 'read'" class="scene">
      <path
        d="M28 152 Q100 130 172 152 L172 180 Q100 158 28 180 Z"
        fill="var(--color-terracotta)"
      />
      <path
        d="M100 88 C80 74 52 72 32 78 L32 162 C52 156 80 158 100 172 Z"
        fill="#FCF9F3"
        stroke="var(--color-ink)"
        stroke-width="4"
        stroke-linejoin="round"
      />
      <path
        class="page"
        d="M100 88 C120 74 148 72 168 78 L168 162 C148 156 120 158 100 172 Z"
        fill="#FCF9F3"
        stroke="var(--color-ink)"
        stroke-width="4"
        stroke-linejoin="round"
      />
      <line x1="100" y1="88" x2="100" y2="172" stroke="var(--color-ink)" stroke-width="4" />
      <g stroke="#A89C89" stroke-width="3.5" stroke-linecap="round">
        <line x1="44" y1="98" x2="84" y2="92" />
        <line x1="44" y1="112" x2="84" y2="106" />
        <line x1="44" y1="126" x2="76" y2="121" />
        <line x1="116" y1="92" x2="156" y2="98" />
        <line x1="116" y1="106" x2="156" y2="112" />
      </g>
      <path
        class="mote"
        d="M148 52 L150 57 L155 59 L150 61 L148 66 L146 61 L141 59 L146 57 Z"
        fill="#D4A847"
      />
    </g>

    <!-- Foco: diana con flecha clavada -->
    <g v-else-if="iconKey === 'focus'" class="scene">
      <g stroke="var(--color-ink)" stroke-width="5" stroke-linecap="round">
        <line x1="82" y1="182" x2="70" y2="208" />
        <line x1="118" y1="182" x2="130" y2="208" />
      </g>
      <circle class="ring" cx="100" cy="126" r="58" fill="none" stroke="#EDA48F" stroke-width="4" />
      <circle cx="100" cy="126" r="56" fill="#FCF9F3" stroke="var(--color-ink)" stroke-width="4" />
      <circle cx="100" cy="126" r="40" fill="#EDA48F" />
      <circle cx="100" cy="126" r="24" fill="#FCF9F3" />
      <circle cx="100" cy="126" r="10" fill="#BC5C42" />
      <g class="arrow">
        <line
          x1="54"
          y1="64"
          x2="97"
          y2="120"
          stroke="var(--color-ink)"
          stroke-width="5"
          stroke-linecap="round"
        />
        <path d="M54 64 L42 58 L50 52 Z" fill="#A8C39A" />
        <path d="M60 72 L48 66 L56 60 Z" fill="#A8C39A" />
      </g>
    </g>

    <!-- Journaling: libreta con línea que se escribe sola -->
    <g v-else-if="iconKey === 'journal'" class="scene">
      <rect
        x="52"
        y="60"
        width="96"
        height="136"
        rx="10"
        fill="#FCF9F3"
        stroke="var(--color-ink)"
        stroke-width="4"
      />
      <g stroke="var(--color-ink)" stroke-width="4" stroke-linecap="round">
        <line x1="70" y1="54" x2="70" y2="68" />
        <line x1="90" y1="54" x2="90" y2="68" />
        <line x1="110" y1="54" x2="110" y2="68" />
        <line x1="130" y1="54" x2="130" y2="68" />
      </g>
      <g stroke="#A89C89" stroke-width="3.5" stroke-linecap="round">
        <line x1="68" y1="92" x2="132" y2="92" />
        <line x1="68" y1="112" x2="132" y2="112" />
        <line x1="68" y1="132" x2="120" y2="132" />
      </g>
      <line
        class="write"
        x1="68"
        y1="156"
        x2="126"
        y2="156"
        stroke="var(--color-terracotta)"
        stroke-width="4"
        stroke-linecap="round"
      />
      <g class="pen">
        <rect
          x="120"
          y="130"
          width="10"
          height="30"
          rx="4"
          fill="var(--color-terracotta)"
          transform="rotate(38 125 145)"
        />
        <path d="M112 158 L118 166 L108 164 Z" fill="var(--color-ink)" />
      </g>
    </g>

    <!-- Sin redes: burbuja de chat en pausa -->
    <g v-else-if="iconKey === 'no_social'" class="scene">
      <rect x="60" y="120" width="80" height="50" rx="14" fill="#E4EDDC" />
      <rect
        x="44"
        y="70"
        width="112"
        height="72"
        rx="16"
        fill="#FCF9F3"
        stroke="var(--color-ink)"
        stroke-width="4"
      />
      <path
        d="M66 140 L58 162 L86 142 Z"
        fill="#FCF9F3"
        stroke="var(--color-ink)"
        stroke-width="4"
        stroke-linejoin="round"
      />
      <g fill="#7B8FA1">
        <circle class="dot d1" cx="76" cy="106" r="6" />
        <circle class="dot d2" cx="100" cy="106" r="6" />
        <circle class="dot d3" cx="124" cy="106" r="6" />
      </g>
      <g class="ban">
        <circle cx="100" cy="122" r="64" fill="none" stroke="#BC5C42" stroke-width="8" />
        <line
          x1="55"
          y1="167"
          x2="145"
          y2="77"
          stroke="#BC5C42"
          stroke-width="8"
          stroke-linecap="round"
        />
      </g>
    </g>

    <!-- Sin teléfono de noche: teléfono bocabajo bajo luna y estrellas -->
    <g v-else-if="iconKey === 'no_phone'" class="scene">
      <path d="M148 58 A26 26 0 1 1 118 34 A20 20 0 0 0 148 58 Z" fill="#F5E8CD" class="moon" />
      <path
        class="star s1"
        d="M56 52 L58 58 L64 60 L58 62 L56 68 L54 62 L48 60 L54 58 Z"
        fill="#D4A847"
      />
      <path
        class="star s2"
        d="M84 34 L85.5 38.5 L90 40 L85.5 41.5 L84 46 L82.5 41.5 L78 40 L82.5 38.5 Z"
        fill="#D4A847"
      />
      <g transform="rotate(-6 100 166)">
        <rect x="54" y="142" width="92" height="50" rx="10" fill="var(--color-ink)" />
        <rect x="60" y="148" width="80" height="38" rx="6" fill="#3A342C" />
        <circle cx="100" cy="188" r="2.5" fill="#FCF9F3" opacity="0.5" />
      </g>
      <g
        class="zzz"
        fill="none"
        stroke="#7B8FA1"
        stroke-width="4"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path class="z z1" d="M124 122 h12 l-12 12 h12" />
        <path class="z z2" d="M146 100 h9 l-9 9 h9" />
      </g>
    </g>

    <!-- Tender la cama: cama lista con almohada que respira -->
    <g v-else-if="iconKey === 'bed'" class="scene">
      <rect x="36" y="92" width="16" height="90" rx="7" fill="var(--color-terracotta)" />
      <rect
        x="44"
        y="142"
        width="126"
        height="38"
        rx="11"
        fill="#FCF9F3"
        stroke="var(--color-ink)"
        stroke-width="4"
      />
      <path
        d="M84 142 L160 142 Q170 142 170 153 L170 168 Q170 180 158 180 L84 180 Q76 160 84 142 Z"
        fill="#A8C39A"
      />
      <path
        d="M92 148 Q88 161 92 174"
        stroke="#5C7650"
        stroke-width="3.5"
        fill="none"
        stroke-linecap="round"
      />
      <rect
        class="pillow"
        x="52"
        y="126"
        width="36"
        height="22"
        rx="10"
        fill="#EFE7D8"
        stroke="var(--color-ink)"
        stroke-width="3.5"
      />
      <g fill="var(--color-ink)">
        <rect x="48" y="180" width="8" height="14" rx="3" />
        <rect x="158" y="180" width="8" height="14" rx="3" />
      </g>
      <path
        class="spark sp1"
        d="M150 104 L153 112 L161 115 L153 118 L150 126 L147 118 L139 115 L147 112 Z"
        fill="#D4A847"
      />
      <path
        class="spark sp2"
        d="M120 84 L122 89 L127 91 L122 93 L120 98 L118 93 L113 91 L118 89 Z"
        fill="#D4A847"
      />
    </g>

    <!-- Aire libre: sol, lomas, árbol y nube que pasea -->
    <g v-else-if="iconKey === 'outdoors'" class="scene">
      <g class="sun">
        <g class="rays" stroke="#D4A847" stroke-width="5" stroke-linecap="round">
          <line x1="142" y1="34" x2="142" y2="44" />
          <line x1="142" y1="100" x2="142" y2="110" />
          <line x1="104" y1="72" x2="114" y2="72" />
          <line x1="170" y1="72" x2="180" y2="72" />
          <line x1="115" y1="45" x2="122" y2="52" />
          <line x1="162" y1="92" x2="169" y2="99" />
          <line x1="115" y1="99" x2="122" y2="92" />
          <line x1="162" y1="52" x2="169" y2="45" />
        </g>
        <circle cx="142" cy="72" r="20" fill="#D4A847" />
      </g>
      <g class="cloud" fill="#FCF9F3" stroke="var(--color-ink)" stroke-width="3.5">
        <path
          d="M42 80 Q42 68 54 68 Q58 56 72 58 Q84 58 86 70 Q96 72 94 82 Q92 90 82 90 L52 90 Q42 90 42 80 Z"
        />
      </g>
      <path d="M4 210 Q52 148 116 174 Q166 192 198 180 L198 210 Z" fill="#A8C39A" opacity="0.55" />
      <path d="M4 210 Q70 168 130 188 Q170 200 198 192 L198 210 Z" fill="#A8C39A" />
      <rect x="58" y="152" width="9" height="32" rx="4" fill="#8A6B4F" />
      <g fill="#5C7650">
        <circle cx="63" cy="138" r="17" />
        <circle cx="49" cy="149" r="11" />
        <circle cx="77" cy="149" r="11" />
      </g>
    </g>

    <!-- Fallback: destello grande que titila -->
    <g v-else class="scene">
      <path
        class="spark sp1"
        d="M100 76 L108 118 L150 126 L108 134 L100 176 L92 134 L50 126 L92 118 Z"
        fill="#D4A847"
      />
      <path
        class="spark sp2"
        d="M146 84 L149 92 L157 95 L149 98 L146 106 L143 98 L135 95 L143 92 Z"
        fill="#E9C281"
      />
    </g>
  </svg>
</template>

<style scoped>
.hh {
  cursor: pointer;
  user-select: none;
}
.scene {
  transform-box: fill-box;
  transform-origin: 50% 100%;
}
.hh.pop .scene {
  animation: hh-pop 0.55s ease;
}
@keyframes hh-pop {
  0% {
    transform: scale(1, 1);
  }
  30% {
    transform: scale(1.05, 0.93);
  }
  60% {
    transform: scale(0.97, 1.04);
  }
  100% {
    transform: scale(1, 1);
  }
}

/* Ejercicio */
.reps {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hh-lift 2.2s ease-in-out infinite;
}
@keyframes hh-lift {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-26px);
  }
}
.effort {
  animation: hh-effort 2.2s ease-in-out infinite;
  opacity: 0;
}
@keyframes hh-effort {
  0%,
  30%,
  75%,
  100% {
    opacity: 0;
  }
  45%,
  60% {
    opacity: 0.9;
  }
}

/* Pasos */
.walk {
  transform-box: fill-box;
  transform-origin: 50% 100%;
  animation: hh-walk 1.4s ease-in-out infinite;
}
@keyframes hh-walk {
  0%,
  100% {
    transform: translateX(0) rotate(0);
  }
  30% {
    transform: translateX(5px) rotate(-2.5deg);
  }
  65% {
    transform: translateX(-3px) rotate(1.5deg);
  }
}
.dashes line {
  animation: hh-dash 1.4s ease-in-out infinite;
}
.dashes line:nth-child(2) {
  animation-delay: 0.25s;
}
.dashes line:nth-child(3) {
  animation-delay: 0.5s;
}
@keyframes hh-dash {
  0%,
  100% {
    opacity: 0.15;
    transform: translateX(0);
  }
  50% {
    opacity: 0.6;
    transform: translateX(-7px);
  }
}

/* Fruta */
.leaf {
  transform-box: fill-box;
  transform-origin: 0% 100%;
  animation: hh-sway 2.8s ease-in-out infinite;
}
@keyframes hh-sway {
  0%,
  100% {
    transform: rotate(-8deg);
  }
  50% {
    transform: rotate(9deg);
  }
}

/* Burbujas (refresco) */
.fizz {
  transform-box: fill-box;
  animation: hh-rise 3.2s ease-in infinite;
  opacity: 0;
}
.f2 {
  animation-delay: 1.1s;
}
.f3 {
  animation-delay: 2.1s;
}
@keyframes hh-rise {
  0% {
    transform: translateY(0);
    opacity: 0;
  }
  15% {
    opacity: 0.8;
  }
  100% {
    transform: translateY(-52px);
    opacity: 0;
  }
}

/* Señal de prohibido compartida */
.ban {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hh-ban 2.4s ease-in-out infinite;
}
@keyframes hh-ban {
  0%,
  100% {
    opacity: 0.55;
    transform: scale(0.985);
  }
  50% {
    opacity: 1;
    transform: scale(1);
  }
}

/* Leer */
.page {
  transform-box: fill-box;
  transform-origin: 0% 50%;
  animation: hh-flip 3.2s ease-in-out infinite;
}
@keyframes hh-flip {
  0%,
  100% {
    transform: scaleX(1);
  }
  50% {
    transform: scaleX(0.85);
  }
}
.mote,
.spark,
.star {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hh-twinkle 2.6s ease-in-out infinite;
}
.sp2,
.s2 {
  animation-delay: 1.3s;
}
@keyframes hh-twinkle {
  0%,
  100% {
    transform: scale(0.6);
    opacity: 0.35;
  }
  50% {
    transform: scale(1);
    opacity: 1;
  }
}

/* Foco */
.ring {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hh-ring 2.4s ease-out infinite;
}
@keyframes hh-ring {
  0% {
    transform: scale(0.96);
    opacity: 0.8;
  }
  70%,
  100% {
    transform: scale(1.12);
    opacity: 0;
  }
}
.arrow {
  transform-box: fill-box;
  transform-origin: 100% 100%;
  animation: hh-wobble 2.4s ease-in-out infinite;
}
@keyframes hh-wobble {
  0%,
  100% {
    transform: rotate(0);
  }
  50% {
    transform: rotate(-2.5deg);
  }
}

/* Journaling */
.write {
  stroke-dasharray: 58;
  animation: hh-write 2.8s ease-in-out infinite;
}
@keyframes hh-write {
  0% {
    stroke-dashoffset: 58;
  }
  55%,
  100% {
    stroke-dashoffset: 0;
  }
}
.pen {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hh-pen 2.8s ease-in-out infinite;
}
@keyframes hh-pen {
  0% {
    transform: translateX(-54px);
  }
  55%,
  100% {
    transform: translateX(0);
  }
}

/* Sin redes */
.dot {
  animation: hh-blink 1.6s ease-in-out infinite;
}
.d2 {
  animation-delay: 0.25s;
}
.d3 {
  animation-delay: 0.5s;
}
@keyframes hh-blink {
  0%,
  100% {
    opacity: 0.25;
  }
  40% {
    opacity: 1;
  }
}

/* Noche */
.moon {
  animation: hh-glow 3.4s ease-in-out infinite;
}
@keyframes hh-glow {
  0%,
  100% {
    opacity: 0.75;
  }
  50% {
    opacity: 1;
  }
}
.z {
  animation: hh-zrise 3.6s ease-in-out infinite;
  opacity: 0;
}
.z2 {
  animation-delay: 1.8s;
}
@keyframes hh-zrise {
  0% {
    transform: translateY(6px);
    opacity: 0;
  }
  25% {
    opacity: 0.9;
  }
  70% {
    opacity: 0;
  }
  100% {
    transform: translateY(-16px);
    opacity: 0;
  }
}

/* Cama */
.pillow {
  transform-box: fill-box;
  transform-origin: 50% 100%;
  animation: hh-breathe 3s ease-in-out infinite;
}
@keyframes hh-breathe {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.06, 1.1);
  }
}

/* Aire libre */
.rays {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hh-spin 16s linear infinite;
}
@keyframes hh-spin {
  to {
    transform: rotate(360deg);
  }
}
.cloud {
  transform-box: fill-box;
  transform-origin: 50% 50%;
  animation: hh-drift 7s ease-in-out infinite;
}
@keyframes hh-drift {
  0%,
  100% {
    transform: translateX(0);
  }
  50% {
    transform: translateX(14px);
  }
}

@media (prefers-reduced-motion: reduce) {
  .hh * {
    animation: none !important;
  }
}
</style>
