<script setup>
import { computed } from 'vue'

const props = defineProps({
  // 'symbol' (only the icon), 'wordmark' (only the text), or 'full' (both combined)
  variant: {
    type: String,
    default: 'full',
    validator: (v) => ['symbol', 'wordmark', 'full'].includes(v),
  },
  // 'brand' (default theme colors), 'white' (monochrome white), 'dark' (monochrome dark/ink), 'terracotta' (monochrome terracotta)
  theme: {
    type: String,
    default: 'brand',
    validator: (v) => ['brand', 'white', 'dark', 'terracotta'].includes(v),
  },
  // Size preset: 'sm', 'md', 'lg', 'xl'
  size: {
    type: String,
    default: 'md',
    validator: (v) => ['sm', 'md', 'lg', 'xl'].includes(v),
  },
})

// Resolve size dimensions
const sizeClasses = computed(() => {
  if (props.variant === 'symbol') {
    return {
      sm: 'w-6 h-6',      // 24px
      md: 'w-12 h-12',    // 48px
      lg: 'w-16 h-16',    // 64px
      xl: 'w-32 h-32',    // 128px
    }[props.size]
  } else if (props.variant === 'wordmark') {
    return {
      sm: 'w-24 h-8',     // 96x32
      md: 'w-36 h-12',    // 144x48
      lg: 'w-48 h-16',    // 192x64
      xl: 'w-72 h-24',    // 288x96
    }[props.size]
  } else {
    // 'full'
    return {
      sm: 'w-32 h-10',    // 128x40
      md: 'w-48 h-16',    // 192x64
      lg: 'w-64 h-22',    // 256x88
      xl: 'w-96 h-32',    // 384x128
    }[props.size]
  }
})

// Color resolution based on the theme
const colors = computed(() => {
  switch (props.theme) {
    case 'white':
      return {
        primary: '#FFFFFF',   // Terracotta equivalent
        secondary: '#FFFFFF', // Sage deep equivalent
        text: '#FFFFFF',
      }
    case 'dark':
      return {
        primary: '#2A251F',
        secondary: '#6F6557',
        text: '#2A251F',
      }
    case 'terracotta':
      return {
        primary: '#C26F4D',
        secondary: '#C26F4D',
        text: '#C26F4D',
      }
    case 'brand':
    default:
      return {
        primary: '#C26F4D',   // Terracotta (action / consequences)
        secondary: '#7CA39D', // Sage deep (success / habits)
        text: '#2A251F',      // Ink (neutral text)
      }
  }
})
</script>

<template>
  <div :class="['inline-block transition-all', sizeClasses]">
    <!-- VARIANT: SYMBOL (Habit Wheel / Roulette) -->
    <svg
      v-if="variant === 'symbol'"
      width="100%"
      height="100%"
      viewBox="0 0 100 100"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      class="select-none"
    >
      <!-- Rueda de hábitos de 8 segmentos (7 completados/squad - Sage deep) -->
      <circle
        cx="45"
        cy="50"
        r="28"
        :stroke="colors.secondary"
        stroke-width="8"
        stroke-dasharray="16 5.9"
        stroke-linecap="round"
        fill="none"
        transform="rotate(-90 45 50)"
      />
      
      <!-- El 8vo segmento destacado (Con consecuencias/ruleta - Terracotta) -->
      <!-- Cierra el trazo de la letra 'd' -->
      <path
        d="M 73,22 L 73,78"
        :stroke="colors.primary"
        stroke-width="8"
        stroke-linecap="round"
      />
      
      <!-- Flecha/Cursor de la ruleta apuntando hacia el centro -->
      <path
        d="M 73,50 L 59,50"
        :stroke="colors.primary"
        stroke-width="8"
        stroke-linecap="round"
      />
      
      <!-- Centro de la ruleta / Eje -->
      <circle
        cx="45"
        cy="50"
        r="4.5"
        :fill="colors.secondary"
      />
    </svg>

    <!-- VARIANT: WORDMARK (dydi Text) -->
    <svg
      v-else-if="variant === 'wordmark'"
      width="100%"
      height="100%"
      viewBox="0 0 240 80"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      class="select-none"
    >
      <text
        x="50%"
        y="58"
        font-family="'Newsreader', 'Georgia', serif"
        font-size="52"
        font-weight="600"
        :fill="colors.text"
        text-anchor="middle"
        letter-spacing="0.05em"
      >
        dydi
      </text>
    </svg>

    <!-- VARIANT: FULL (Symbol + Wordmark) -->
    <svg
      v-else
      width="100%"
      height="100%"
      viewBox="0 0 280 100"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      class="select-none"
    >
      <g transform="translate(10, 10)">
        <!-- Icono lateral (escala a 0.8 para ajustarse con el texto) -->
        <g transform="scale(0.8)">
          <circle
            cx="45"
            cy="50"
            r="28"
            :stroke="colors.secondary"
            stroke-width="8"
            stroke-dasharray="16 5.9"
            stroke-linecap="round"
            fill="none"
            transform="rotate(-90 45 50)"
          />
          <path
            d="M 73,22 L 73,78"
            :stroke="colors.primary"
            stroke-width="8"
            stroke-linecap="round"
          />
          <path
            d="M 73,50 L 59,50"
            :stroke="colors.primary"
            stroke-width="8"
            stroke-linecap="round"
          />
          <circle
            cx="45"
            cy="50"
            r="4.5"
            :fill="colors.secondary"
          />
        </g>
        
        <!-- Texto de la marca 'dydi' al lado -->
        <text
          x="95"
          y="58"
          font-family="'Newsreader', 'Georgia', serif"
          font-size="44"
          font-weight="600"
          :fill="colors.text"
          letter-spacing="0.02em"
        >
          dydi
        </text>
      </g>
    </svg>
  </div>
</template>
