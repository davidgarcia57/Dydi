<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useGroupStore } from '@/stores/group'

const router = useRouter()
const group = useGroupStore()

onMounted(async () => {
  if (!group.myGroups.length) {
    try {
      await group.loadMyGroups()
    } catch {
      // sin lista no hay switcher; la vista actual ya maneja sus errores
    }
  }
})

async function onChange(event) {
  const value = event.target.value
  if (value === '__new__') {
    // Restaura la selección visual y manda al onboarding a crear/unirse.
    event.target.value = group.group?.id ?? ''
    router.push('/onboarding')
    return
  }
  if (!value || value === group.group?.id) return
  await group.switchGroup(value)
  // Recarga dura: todas las vistas y el WebSocket renacen contra el grupo nuevo.
  window.location.reload()
}
</script>

<template>
  <label class="block">
    <span class="sr-only">Grupo activo</span>
    <select
      class="w-full rounded-xl border border-hairline bg-paper px-3 py-2 text-sm font-semibold text-ink focus:outline-none focus:border-sage-deep"
      :value="group.group?.id ?? ''"
      @change="onChange"
    >
      <option v-if="!group.group" value="" disabled>Sin grupo</option>
      <option v-for="g in group.myGroups" :key="g.id" :value="g.id">
        {{ g.name }}
      </option>
      <option value="__new__">+ Crear o unirme a otro…</option>
    </select>
  </label>
</template>
