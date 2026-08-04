<script setup>
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import BaseAvatar from '@/components/ui/BaseAvatar.vue'
import TargetGlyph from '@/components/ui/TargetGlyph.vue'
import { missedThisWeek, todayStatus } from '@/composables/useWeekStatus'

// Pulso del squad: de un vistazo, quién ya cumplió hoy (anillo salvia), quién
// va pendiente (ámbar), quién dejó ir el día (coral), quién está en riesgo de
// ruleta (🎯) y quién anda conectado ahora (punto verde).
const auth = useAuthStore()
const group = useGroupStore()

const RING = {
  done: 'ring-sage-deep',
  pending: 'ring-amber',
  missed: 'ring-coral',
}

const STATUS_HINT = {
  done: 'ya cumplió hoy',
  pending: 'aún no hace check-in',
  missed: 'se le fue el día',
}

const pulse = computed(() => {
  const me = auth.user?.id
  return [...group.members]
    .sort((a, b) => {
      if (a.user_id === me) return -1
      if (b.user_id === me) return 1
      return a.display_name.localeCompare(b.display_name, 'es')
    })
    .map((m) => {
      const status = todayStatus(m.user_id)
      return {
        ...m,
        status,
        atRisk: missedThisWeek(m.user_id) > 0,
        online: group.onlineMembers.has(m.user_id),
        isMe: m.user_id === me,
      }
    })
})

function hint(m) {
  const parts = [m.status ? STATUS_HINT[m.status] : 'sin hábitos asignados']
  if (m.atRisk) parts.push('en riesgo de ruleta')
  if (m.online) parts.push('en línea')
  return `${m.display_name}: ${parts.join(' · ')}`
}

// Tope de miembros por grupo (lo impone un trigger en la BD). Los lugares libres
// se dibujan como anillos punteados: con un squad chico esta tarjeta era un avatar
// y un mar de espacio muerto, y así el vacío pasa de defecto a información —
// "1 de 8" — que es la tercera guía de NN/g para estados vacíos.
const SQUAD_CAPACITY = 8
const freeSeats = computed(() => Math.max(0, SQUAD_CAPACITY - pulse.value.length))
</script>

<template>
  <div v-if="pulse.length" class="flex flex-wrap gap-x-4 gap-y-3">
    <div
      v-for="m in pulse"
      :key="m.user_id"
      class="flex flex-col items-center gap-1.5 w-14"
      :title="hint(m)"
    >
      <div class="relative">
        <div
          class="rounded-full ring-2 ring-offset-2 ring-offset-paper"
          :class="m.status ? RING[m.status] : 'ring-hairline'"
        >
          <BaseAvatar :name="m.display_name" size="md" />
        </div>
        <span
          v-if="m.online"
          class="absolute bottom-0 -right-0.5 w-3 h-3 rounded-full bg-sage-deep border-2 border-paper"
        />
        <span
          v-if="m.atRisk"
          class="absolute -top-1.5 -right-2 w-5 h-5 rounded-full bg-paper shadow-flat flex items-center justify-center text-coral-deep"
        >
          <TargetGlyph :size="12" />
        </span>
      </div>
      <span class="text-[10px] font-semibold text-ink-soft truncate w-full text-center">
        {{ m.isMe ? 'Tú' : m.display_name.split(' ')[0] }}
      </span>
    </div>

    <div
      v-for="i in freeSeats"
      :key="`seat-${i}`"
      class="flex flex-col items-center gap-1.5 w-14"
      :title="`Lugar libre (${pulse.length} de ${SQUAD_CAPACITY} ocupados)`"
    >
      <!-- w-12 y no w-10: el avatar md mide 40px pero lleva ring-2 con
           ring-offset-2, así que su huella visual es de 48px. -->
      <div
        class="w-12 h-12 rounded-full border-2 border-dashed"
        :class="i === 1 ? 'border-hairline' : 'border-hairline/60'"
      />
      <span v-if="i === 1" class="text-[10px] font-semibold text-ink-faint">libre</span>
    </div>
  </div>

  <p v-if="pulse.length && freeSeats" class="text-[10px] text-ink-faint mt-3">
    {{ pulse.length }} de {{ SQUAD_CAPACITY }} lugares · quedan {{ freeSeats }}
  </p>
</template>
