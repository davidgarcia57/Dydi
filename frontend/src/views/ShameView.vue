<script setup>
import { computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { usePenaltiesStore } from '@/stores/penalties'

const auth      = useAuthStore()
const group     = useGroupStore()
const habits    = useHabitsStore()
const penalties = usePenaltiesStore()

const displayName = computed(() =>
  auth.user?.user_metadata?.display_name
  ?? auth.user?.email?.split('@')[0]
  ?? 'Tú'
)

const myStreak = computed(() => habits.streaks[auth.user?.id] ?? 0)

const myDebts = computed(() =>
  penalties.debts.filter(d => d.debtor_id === auth.user?.id)
)

const COLORS = ['bg-sage-deep', 'bg-terracotta', 'bg-sage', 'bg-amber', 'bg-coral']
const initials  = (n = '') => n.trim().split(/\s+/).map(w => w[0]).join('').slice(0, 2).toUpperCase()
const avatarBg  = (n = '') => COLORS[(n?.charCodeAt(0) ?? 0) % COLORS.length]
const shortDate = iso => new Date(iso).toLocaleDateString('es-MX', { month: 'long', day: 'numeric' })

onMounted(async () => {
  await group.autoLoad()
  if (group.group?.id && auth.user?.id) {
    await Promise.all([
      habits.loadStreaks(auth.user.id),
      penalties.loadDebts(group.group.id),
    ])
  }
})
</script>

<template>
  <div class="max-w-md mx-auto px-4 pt-4 pb-6">

    <!-- Perfil -->
    <div class="rounded-card shadow-card bg-paper p-5 mb-6 flex items-center gap-4">
      <div
        class="w-16 h-16 rounded-full flex items-center justify-center
               text-paper text-xl font-bold flex-shrink-0"
        :class="avatarBg(displayName)"
      >
        {{ initials(displayName) }}
      </div>
      <div class="flex-1 min-w-0">
        <h1 class="font-serif text-xl font-semibold text-ink leading-tight truncate">
          {{ displayName }}
        </h1>
        <p class="text-xs text-ink-soft mt-0.5">{{ group.group?.name ?? 'Sin grupo' }}</p>
      </div>
      <div class="text-right flex-shrink-0">
        <p class="font-serif text-4xl font-semibold text-terracotta leading-none">{{ myStreak }}</p>
        <p class="text-eyebrow text-terracotta mt-1">RACHA</p>
      </div>
    </div>

    <!-- Mis deudas -->
    <section>
      <h2 class="text-eyebrow mb-3">MIS DEUDAS ACTIVAS</h2>

      <div v-if="!myDebts.length"
        class="rounded-card border border-sage/30 bg-sage/10 px-4 py-6 text-center">
        <p class="font-serif text-3xl mb-1">✓</p>
        <p class="text-sm font-semibold text-sage-deep">Sin deudas pendientes</p>
        <p class="text-xs text-ink-soft mt-1">¡Estás al corriente!</p>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="debt in myDebts"
          :key="debt.id"
          class="rounded-card shadow-flat bg-paper p-4 border-l-4"
          :class="debt.is_collective ? 'border-coral' : 'border-terracotta'"
        >
          <div class="flex items-center justify-between mb-2">
            <span
              class="text-[10px] font-bold rounded-pill px-2 py-0.5"
              :class="debt.is_collective
                ? 'bg-coral/20 text-coral'
                : 'bg-terracotta/20 text-terracotta'"
            >
              {{ debt.is_collective ? 'COLECTIVA' : 'PERSONAL' }}
            </span>
            <span class="text-[10px] text-ink-faint">
              expira {{ shortDate(debt.expires_at) }}
            </span>
          </div>
          <p class="text-base font-semibold text-ink">
            {{ debt.punishment_emoji ?? '' }} {{ debt.punishment_text }}
          </p>
        </div>
      </div>
    </section>

  </div>
</template>
