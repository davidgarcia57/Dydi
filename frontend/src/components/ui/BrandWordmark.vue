<script setup>
// Wordmark de texto provisional (sin logo todavía). Cuando exista un logo real,
// se reemplaza solo este componente.
defineProps({
  size: {
    type: String,
    default: 'md', // sm | md | lg | xl
    validator: (v) => ['sm', 'md', 'lg', 'xl'].includes(v),
  },
})

const SIZES = {
  sm: 'text-xl',
  md: 'text-2xl',
  lg: 'text-4xl',
  xl: 'text-5xl',
}
</script>

<template>
  <span
    class="font-serif font-semibold lowercase tracking-tight text-terracotta"
    :class="SIZES[size]"
  >
    dydi
  </span>
</template>
