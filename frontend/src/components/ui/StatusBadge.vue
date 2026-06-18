<script setup>
import { computed } from 'vue'

// Pill de estado semántico. Usa color + texto (nunca solo color) para que
// sea legible sin depender de distinguir tonos (accesibilidad).
const props = defineProps({
  status: {
    type: String,
    required: true, // done | pending | missed
    validator: (v) => ['done', 'pending', 'missed'].includes(v),
  },
  label: { type: String, default: '' },
})

const STYLE = {
  done: { cls: 'bg-sage-soft text-sage-deep', label: '✓ hoy' },
  pending: { cls: 'bg-amber-soft text-amber-deep', label: 'pendiente' },
  missed: { cls: 'bg-coral-soft text-coral-deep', label: '✗ hoy' },
}

const cls = computed(() => STYLE[props.status]?.cls ?? 'bg-hairline text-ink-soft')
const text = computed(() => props.label || STYLE[props.status]?.label || props.status)
</script>

<template>
  <span class="rounded-pill px-2.5 py-0.5 text-xs font-semibold" :class="cls">
    {{ text }}
  </span>
</template>
