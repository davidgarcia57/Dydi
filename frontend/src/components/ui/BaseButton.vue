<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: {
    type: String,
    default: 'primary', // primary | secondary | ghost | ink
    validator: (v) => ['primary', 'secondary', 'ghost', 'ink'].includes(v),
  },
  size: {
    type: String,
    default: 'md', // sm | md | lg
    validator: (v) => ['sm', 'md', 'lg'].includes(v),
  },
  type: { type: String, default: 'button' },
  loading: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  block: { type: Boolean, default: false },
})

const VARIANTS = {
  primary: 'bg-sage-deep text-paper',
  secondary: 'bg-terracotta text-paper',
  ink: 'bg-ink text-paper',
  ghost: 'border border-ink/20 text-ink bg-transparent',
}

const SIZES = {
  sm: 'px-4 py-2 text-xs',
  md: 'px-6 py-3 text-sm',
  lg: 'px-8 py-3.5 text-sm',
}

const classes = computed(() => [
  'inline-flex items-center justify-center gap-2 rounded-pill font-bold',
  'transition-all active:opacity-80 active:scale-[0.99]',
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sage-deep/50 focus-visible:ring-offset-2 focus-visible:ring-offset-cream',
  'disabled:opacity-60 disabled:pointer-events-none',
  VARIANTS[props.variant],
  SIZES[props.size],
  props.block ? 'w-full' : '',
])
</script>

<template>
  <button :type="type" :class="classes" :disabled="disabled || loading">
    <span
      v-if="loading"
      class="w-4 h-4 rounded-full border-2 border-current border-t-transparent animate-spin"
      aria-hidden="true"
    />
    <slot v-else />
  </button>
</template>
