<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { useGroupSocket } from '@/composables/useGroupSocket'

const auth   = useAuthStore()
const group  = useGroupStore()
const habits = useHabitsStore()
const loaded = ref(false)

function localDate() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// Group checkins by member
const squadRows = computed(() => {
  const byUser = {}
  for (const c of habits.todayCheckins) {
    if (!byUser[c.user_id]) {
      byUser[c.user_id] = {
        user_id: c.user_id,
        display_name: c.display_name,
        habits: [],
      }
    }
    byUser[c.user_id].habits.push(c)
  }
  return Object.values(byUser).sort((a, b) =>
    a.display_name.localeCompare(b.display_name, 'es')
  )
})

const COLORS = ['bg-sage-deep', 'bg-terracotta', 'bg-sage', 'bg-amber', 'bg-coral']
const initials = (n = '') => n.trim().split(/\s+/).map(w => w[0]).join('').slice(0, 2).toUpperCase()
const avatarBg = (n = '') => COLORS[(n?.charCodeAt(0) ?? 0) % COLORS.length]

const STATUS_DOT = {
  done:    'bg-sage',
  pending: 'bg-amber',
  missed:  'bg-coral',
}

const STATUS_LABEL = {
  done:    { cls: 'bg-sage/30 text-sage-deep',  label: '✓' },
  pending: { cls: 'bg-amber/30 text-amber',     label: '–' },
  missed:  { cls: 'bg-coral/30 text-coral',     label: '✗' },
}

let socketDisconnect = null

onMounted(async () => {
  await group.autoLoad()
  if (group.group?.id) {
    await habits.loadToday(group.group.id)
    const ids = [...new Set(habits.todayCheckins.map(c => c.user_id))]
    await Promise.all(ids.map(id => habits.loadStreaks(id)))
    const { disconnect } = useGroupSocket(group.group.id)
    socketDisconnect = disconnect
  }
  loaded.value = true
})

onUnmounted(() => socketDisconnect?.())
</script>

<template>
  <div class="max-w-md mx-auto px-4 pt-4 pb-6">

    <header class="flex items-center justify-between mb-6">
      <h1 class="font-serif text-2xl font-semibold text-ink">Squad</h1>
      <span class="text-eyebrow">{{ group.group?.name ?? '' }}</span>
    </header>

    <div v-if="!squadRows.length"
      class="rounded-card bg-surface border border-hairline py-10 text-center text-sm text-ink-soft">
      <span v-if="!loaded">Cargando el squad…</span>
      <span v-else>Ningún miembro tiene hábitos asignados todavía.</span>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="row in squadRows"
        :key="row.user_id"
        class="rounded-card shadow-flat bg-surface p-4"
        :class="{ 'ring-2 ring-sage/30': group.onlineMembers.has(row.user_id) }"
      >
        <div class="flex items-center gap-3 mb-3">
          <!-- Avatar + online dot -->
          <div class="relative">
            <div
              class="w-10 h-10 rounded-full flex items-center justify-center
                     text-paper text-sm font-bold"
              :class="avatarBg(row.display_name)"
            >
              {{ initials(row.display_name) }}
            </div>
            <span
              v-if="group.onlineMembers.has(row.user_id)"
              class="absolute bottom-0 right-0 w-3 h-3 rounded-full bg-sage-deep
                     border-2 border-paper"
            />
          </div>

          <!-- Name + streak -->
          <div class="flex-1 min-w-0">
            <div class="flex items-baseline gap-2">
              <span class="font-semibold text-sm text-ink truncate">{{ row.display_name }}</span>
              <span class="text-xs text-terracotta font-medium flex-shrink-0">
                ★ {{ habits.streaks[row.user_id] ?? 0 }}
              </span>
            </div>
            <p class="text-xs text-ink-soft mt-0.5">
              {{ row.user_id === auth.user?.id ? 'Tú' : '' }}
            </p>
          </div>

          <!-- Overall status pill -->
          <div
            class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold"
            :class="row.habits.every(h => h.status === 'done')
              ? 'bg-sage/30 text-sage-deep'
              : row.habits.some(h => h.status === 'missed')
                ? 'bg-coral/30 text-coral'
                : 'bg-amber/30 text-amber'"
          >
            {{ row.habits.every(h => h.status === 'done') ? '✓' : row.habits.some(h => h.status === 'missed') ? '✗' : '–' }}
          </div>
        </div>

        <!-- Habits list -->
        <div class="space-y-1.5 pl-[52px]">
          <div
            v-for="h in row.habits"
            :key="h.habit_id"
            class="flex items-center gap-2"
          >
            <span
              class="w-2 h-2 rounded-full flex-shrink-0"
              :class="STATUS_DOT[h.status] ?? 'bg-hairline'"
            />
            <span class="text-xs text-ink truncate">{{ h.habit_name }}</span>
            <span
              v-if="h.scheduled_time"
              class="text-[10px] text-ink-faint ml-auto flex-shrink-0"
            >
              {{ h.scheduled_time }}
            </span>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>
