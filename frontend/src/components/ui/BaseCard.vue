<script setup>
import { computed } from 'vue'

// Card del sistema Dydi. `elevated` = card flotante (shadow-card sobre paper);
// `flat` = card de fondo (shadow-flat sobre surface). El padding lo decide
// quien la use (pasa p-4 / p-5 como class, Vue las fusiona).
const props = defineProps({
  variant: {
    type: String,
    default: 'elevated', // elevated | flat
    validator: (v) => ['elevated', 'flat'].includes(v),
  },
})

const classes = computed(() =>
  props.variant === 'elevated'
    ? 'rounded-card shadow-card bg-paper'
    : 'rounded-card shadow-flat bg-surface'
)
</script>

<template>
  <div :class="classes">
    <slot />
  </div>
</template>
