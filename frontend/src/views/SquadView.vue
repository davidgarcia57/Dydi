<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { useGroupSocket } from '@/composables/useGroupSocket'
import { usePenaltiesStore } from '@/stores/penalties'
import PageContainer from '@/components/ui/PageContainer.vue'

const auth = useAuthStore()
const group = useGroupStore()
const habits = useHabitsStore()
const penalties = usePenaltiesStore()
const loaded = ref(false)

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
  return Object.values(byUser).sort((a, b) => a.display_name.localeCompare(b.display_name, 'es'))
})

const COLORS = ['bg-sage-deep', 'bg-terracotta', 'bg-sage', 'bg-amber', 'bg-coral']
const initials = (n = '') =>
  n
    .trim()
    .split(/\s+/)
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
const avatarBg = (n = '') => COLORS[(n?.charCodeAt(0) ?? 0) % COLORS.length]

const STATUS_DOT = {
  done: 'bg-sage',
  pending: 'bg-amber',
  missed: 'bg-coral',
}

let socketDisconnect = null

onMounted(async () => {
  await group.autoLoad()
  if (group.group?.id) {
    await habits.loadToday(group.group.id)
    const ids = [...new Set(habits.todayCheckins.map((c) => c.user_id))]
    await Promise.all([
      ...ids.map((id) => habits.loadStreaks(id)),
      penalties.loadDebts(group.group.id),
    ])
    const { disconnect } = useGroupSocket(group.group.id)
    socketDisconnect = disconnect
  }
  loaded.value = true
})

onUnmounted(() => socketDisconnect?.())
</script>

<template>
  <PageContainer>
    <header class="mb-6">
      <div class="flex items-center justify-between">
        <h1 class="font-serif text-2xl font-semibold text-ink">Squad</h1>
        <span class="text-eyebrow">{{ group.group?.name ?? '' }}</span>
      </div>
      <p class="text-xs text-ink-faint mt-0.5">Presencia y hábitos de hoy</p>
    </header>

    <div
      v-if="!squadRows.length"
      class="rounded-card bg-surface border border-hairline py-10 text-center text-sm text-ink-soft"
    >
      <span v-if="!loaded">Cargando el squad…</span>
      <span v-else>Ningún miembro tiene hábitos asignados todavía.</span>
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-3">
      <div
        v-for="row in squadRows"
        :key="row.user_id"
        class="rounded-card shadow-flat bg-paper p-4"
        :class="{ 'ring-2 ring-sage/40': group.onlineMembers.has(row.user_id) }"
      >
        <div class="flex items-center gap-3 mb-3">
          <!-- Avatar + online dot -->
          <div class="relative">
            <div
              class="w-10 h-10 rounded-full flex items-center justify-center text-paper text-sm font-bold"
              :class="avatarBg(row.display_name)"
            >
              {{ initials(row.display_name) }}
            </div>
            <span
              v-if="group.onlineMembers.has(row.user_id)"
              class="absolute bottom-0 right-0 w-3 h-3 rounded-full bg-sage-deep border-2 border-paper"
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
          <span
            class="rounded-pill px-2.5 py-1 text-[10px] font-bold flex-shrink-0"
            :class="
              row.habits.every((h) => h.status === 'done')
                ? 'bg-sage-soft text-sage-deep'
                : row.habits.some((h) => h.status === 'missed')
                  ? 'bg-coral-soft text-coral-deep'
                  : 'bg-amber-soft text-amber-deep'
            "
          >
            {{
              row.habits.every((h) => h.status === 'done')
                ? '✓ hoy'
                : row.habits.some((h) => h.status === 'missed')
                  ? '✗ falló'
                  : '· pendiente'
            }}
          </span>
        </div>

        <!-- Habits list -->
        <div class="space-y-1.5 pl-[52px]">
          <div v-for="h in row.habits" :key="h.habit_id" class="flex items-center gap-2">
            <span
              class="w-2 h-2 rounded-full flex-shrink-0"
              :class="STATUS_DOT[h.status] ?? 'bg-hairline'"
            />
            <span class="text-xs text-ink truncate">{{ h.habit_name }}</span>
            <span v-if="h.scheduled_time" class="text-[10px] text-ink-faint ml-auto flex-shrink-0">
              {{ h.scheduled_time }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Muro de la Vergüenza -->
    <section v-if="penalties.debts.length > 0" class="mt-8">
      <h2 class="text-eyebrow mb-3 flex items-center gap-2 text-coral-deep">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2.5"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        MURO DE LA VERGÜENZA
      </h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div
          v-for="debt in penalties.debts"
          :key="debt.id"
          class="rounded-card bg-coral-soft/30 border border-coral/30 p-4 flex items-center gap-3"
        >
          <div
            class="w-10 h-10 rounded-full flex-shrink-0 flex items-center justify-center text-paper text-sm font-bold"
            :class="avatarBg(squadRows.find((r) => r.user_id === debt.debtor_id)?.display_name)"
          >
            {{ initials(squadRows.find((r) => r.user_id === debt.debtor_id)?.display_name) }}
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex justify-between items-center mb-0.5">
              <span class="font-semibold text-sm text-ink truncate">
                {{ squadRows.find((r) => r.user_id === debt.debtor_id)?.display_name ?? 'Miembro' }}
              </span>
              <span
                v-if="debt.scope === 'collective'"
                class="rounded-pill bg-coral/20 text-coral-deep text-[10px] font-bold px-2 py-0.5"
              >
                colectiva
              </span>
            </div>
            <p class="text-sm font-semibold text-coral-deep leading-snug">
              {{ debt.punishment_emoji ?? '' }} {{ debt.punishment_text }}
            </p>
          </div>
        </div>
      </div>
    </section>
  </PageContainer>
</template>
