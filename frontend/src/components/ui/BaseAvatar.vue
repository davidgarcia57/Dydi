<script setup>
import { computed } from 'vue'

// Avatar de iniciales con color derivado del nombre — centraliza la lógica
// que estaba duplicada en TodayView / SquadView / etc.
const props = defineProps({
  name: { type: String, default: '' },
  size: {
    type: String,
    default: 'md', // sm | md | lg
    validator: (v) => ['sm', 'md', 'lg'].includes(v),
  },
})

const COLORS = ['bg-sage-deep', 'bg-terracotta', 'bg-sage', 'bg-amber', 'bg-coral', 'bg-ink-soft']

const SIZES = {
  sm: 'w-6 h-6 text-[9px]',
  md: 'w-10 h-10 text-sm',
  lg: 'w-12 h-12 text-base',
}

const initials = computed(() =>
  (props.name || '')
    .trim()
    .split(/\s+/)
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
)

const bg = computed(() => COLORS[(props.name?.charCodeAt(0) ?? 0) % COLORS.length])
</script>

<template>
  <div
    class="rounded-full flex-shrink-0 flex items-center justify-center text-paper font-bold"
    :class="[SIZES[size], bg]"
  >
    {{ initials }}
  </div>
</template>
