<script setup>
// Contenedor de página responsive. Reemplaza el viejo `max-w-md mx-auto` que
// dejaba el contenido como una columna de celular en escritorio.
//   wide   → feeds / grids (hasta ~1152px)
//   narrow → páginas enfocadas: perfil, formularios (hasta ~672px)
defineProps({
  width: {
    type: String,
    default: 'wide',
    validator: (v) => ['wide', 'narrow'].includes(v),
  },
})
</script>

<template>
  <div
    class="mx-auto w-full px-4 pt-4 pb-8 sm:px-6 lg:px-8 lg:pt-8"
    :class="width === 'narrow' ? 'max-w-2xl' : 'max-w-6xl'"
  >
    <slot />
  </div>
</template>
