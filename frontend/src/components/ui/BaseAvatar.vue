<script setup>
import { computed, ref } from 'vue'

// Avatar: muestra foto (src) si se proporciona, con fallback a iniciales de color.
const props = defineProps({
  name: { type: String, default: '' },
  src: { type: String, default: '' }, // URL de foto de perfil (opcional)
  size: {
    type: String,
    default: 'md', // sm | md | lg
    validator: (v) => ['sm', 'md', 'lg'].includes(v),
  },
})

const imgError = ref(false)
const showImg = computed(() => !!props.src && !imgError.value)

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
    class="rounded-full flex-shrink-0 overflow-hidden flex items-center justify-center text-paper font-bold"
    :class="[SIZES[size], showImg ? '' : bg]"
  >
    <img
      v-if="showImg"
      :src="src"
      :alt="name"
      class="w-full h-full object-cover"
      @error="imgError = true"
    />
    <span v-else>{{ initials }}</span>
  </div>
</template>
