<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { usePenaltiesStore } from '@/stores/penalties'

const router    = useRouter()
const auth      = useAuthStore()
const group     = useGroupStore()
const habits    = useHabitsStore()
const penalties = usePenaltiesStore()
const loggingOut  = ref(false)
const leavingGroup = ref(false)
const confirmLeave = ref(false)

async function handleLogout() {
  loggingOut.value = true
  await auth.logout()
  group.reset()
  router.replace('/login')
}

async function handleLeaveGroup() {
  leavingGroup.value = true
  try {
    await group.leaveGroup()
    router.replace('/onboarding')
  } catch (e) {
    confirmLeave.value = false
  } finally {
    leavingGroup.value = false
  }
}

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

    <!-- Salir del grupo -->
    <div v-if="group.group" class="mb-3">
      <div v-if="!confirmLeave">
        <button
          class="w-full rounded-pill border border-hairline text-ink-soft py-3 font-semibold
                 text-sm active:opacity-70 transition-opacity"
          @click="confirmLeave = true"
        >
          Salir del grupo
        </button>
      </div>
      <div v-else class="rounded-card border border-coral/40 bg-coral/5 p-4">
        <p class="text-sm font-semibold text-ink mb-1">
          ¿Seguro que quieres salir de <span class="text-coral">{{ group.group.name }}</span>?
        </p>
        <p class="text-xs text-ink-soft mb-4">
          Perderás tus hábitos y rachas en este grupo.
        </p>
        <div class="flex gap-2">
          <button
            :disabled="leavingGroup"
            class="flex-1 rounded-pill bg-coral text-paper py-2.5 font-bold text-sm
                   disabled:opacity-40 active:opacity-80 transition-opacity"
            @click="handleLeaveGroup"
          >
            {{ leavingGroup ? 'Saliendo…' : 'Sí, salir' }}
          </button>
          <button
            class="flex-1 rounded-pill border border-hairline text-ink-soft py-2.5
                   font-semibold text-sm"
            @click="confirmLeave = false"
          >
            Cancelar
          </button>
        </div>
      </div>
    </div>

    <!-- Cerrar sesión -->
    <button
      :disabled="loggingOut"
      class="w-full rounded-pill border border-hairline text-ink-soft py-3 font-semibold
             text-sm mb-6 disabled:opacity-40 active:opacity-70 transition-opacity"
      @click="handleLogout"
    >
      {{ loggingOut ? 'Cerrando sesión…' : 'Cerrar sesión' }}
    </button>

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
          :class="debt.scope === 'collective' ? 'border-coral' : 'border-terracotta'"
        >
          <div class="flex items-center justify-between mb-2">
            <span
              class="text-[10px] font-bold rounded-pill px-2 py-0.5"
              :class="debt.scope === 'collective'
                ? 'bg-coral/20 text-coral'
                : 'bg-terracotta/20 text-terracotta'"
            >
              {{ debt.scope === 'collective' ? 'COLECTIVA' : 'PERSONAL' }}
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
