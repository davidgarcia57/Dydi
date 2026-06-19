<script setup>
defineProps({
  size: {
    type: String,
    default: 'md', // sm | md | lg | xl
    validator: (v) => ['sm', 'md', 'lg', 'xl'].includes(v),
  },
  showText: {
    type: Boolean,
    default: true,
  }
})

const TEXT_SIZES = {
  sm: 'text-xl',
  md: 'text-2xl',
  lg: 'text-4xl',
  xl: 'text-5xl',
}

const LOGO_SIZES = {
  sm: 'h-6 w-6',
  md: 'h-8 w-8',
  lg: 'h-12 w-12',
  xl: 'h-16 w-16',
}
</script>

<template>
  <div class="flex items-center gap-2">
    <!-- SVG Logo -->
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 240 240" :class="LOGO_SIZES[size]" class="shrink-0">
      <g transform="translate(120, 120)">
        <path d="M 0 -24 L 0 -88 A 88 88 0 0 1 76.21 44 L 20.78 12 Z" class="fill-hairline stroke-surface" stroke-width="6" stroke-linejoin="round" transform="rotate(240)" />
        <path d="M 0 -24 L 0 -88 A 88 88 0 0 1 76.21 44 L 20.78 12 Z" class="fill-terracotta stroke-surface" stroke-width="6" stroke-linejoin="round" transform="rotate(0)" />
        <g transform="translate(0, 12)">
          <path d="M 0 -24 L 0 -88 A 88 88 0 0 1 76.21 44 L 20.78 12 Z" class="fill-ink opacity-10" transform="rotate(120) translate(0, 4)" />
          <path d="M 0 -24 L 0 -88 A 88 88 0 0 1 76.21 44 L 20.78 12 Z" class="fill-sage-deep stroke-surface" stroke-width="6" stroke-linejoin="round" transform="rotate(120)" />
        </g>
        <circle cx="82" cy="-35" r="7" class="fill-ink" />
        <circle cx="94" cy="-15" r="4" class="fill-ink-soft" />
      </g>
    </svg>
    
    <span
      v-if="showText"
      class="font-serif font-semibold lowercase tracking-tight text-terracotta"
      :class="TEXT_SIZES[size]"
    >
      dydi
    </span>
  </div>
</template>
