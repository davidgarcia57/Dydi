<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { useGroupSocket } from '@/composables/useGroupSocket'

const router = useRouter()
const auth   = useAuthStore()
const group  = useGroupStore()
const habits = useHabitsStore()
const loaded = ref(false)

// ── Countdown ────────────────────────────────────────────────────────────────
const now = ref(new Date())
let ticker

// Cycle closes every Sunday at midnight (end of week)
const cycleEnd = computed(() => {
  const d = new Date()
  const daysLeft = d.getDay() === 0 ? 7 : 7 - d.getDay()
  d.setDate(d.getDate() + daysLeft)
  d.setHours(0, 0, 0, 0)
  return d
})

const countdown = computed(() => {
  const diff = cycleEnd.value - now.value
  if (diff <= 0) return { days: '00', hours: '00', mins: '00' }
  const pad = n => String(n).padStart(2, '0')
  return {
    days:  pad(Math.floor(diff / 86_400_000)),
    hours: pad(Math.floor((diff % 86_400_000) / 3_600_000)),
    mins:  pad(Math.floor((diff % 3_600_000) / 60_000)),
  }
})

const closingLabel = computed(() => {
  const d = cycleEnd.value
  const names = ['dom', 'lun', 'mar', 'mié', 'jue', 'vie', 'sáb']
  return `${names[d.getDay()]} 00:00`
})

const weekNumber = computed(() => {
  const d = new Date()
  const jan1 = new Date(d.getFullYear(), 0, 1)
  return Math.ceil(((d - jan1) / 86_400_000 + jan1.getDay() + 1) / 7)
})

// ── My check-in ──────────────────────────────────────────────────────────────
const myCheckins = computed(() =>
  habits.todayCheckins.filter(c => c.user_id === auth.user?.id)
)

const hasPending = computed(() => myCheckins.value.some(c => c.status === 'pending'))
const allDone    = computed(() => myCheckins.value.length > 0 && myCheckins.value.every(c => c.status === 'done'))
const anyMissed  = computed(() => myCheckins.value.some(c => c.status === 'missed') && !hasPending.value)

const myStreak = computed(() => habits.streaks[auth.user?.id] ?? 0)

// ── Squad stats (per member, not per checkin) ─────────────────────────────────
const stats = computed(() => {
  const byUser = {}
  for (const c of habits.todayCheckins) {
    if (!byUser[c.user_id]) byUser[c.user_id] = []
    byUser[c.user_id].push(c.status)
  }
  let done = 0, pending = 0, missed = 0
  for (const statuses of Object.values(byUser)) {
    if (statuses.every(s => s === 'done'))        done++
    else if (statuses.some(s => s === 'pending')) pending++
    else                                          missed++
  }
  return { done, pending, missed }
})

const progressPct = computed(() => {
  const total = group.members.length
  return total ? Math.round((stats.value.done / total) * 100) : 0
})

// ── Online members ────────────────────────────────────────────────────────────
const onlineAvatars = computed(() =>
  group.members
    .filter(m => group.onlineMembers.has(m.user_id))
    .slice(0, 5)
)

// ── Squad list (each member's checkin row) ────────────────────────────────────
const squadRows = computed(() =>
  habits.todayCheckins.filter(c => c.user_id !== auth.user?.id)
)

// ── Helpers ───────────────────────────────────────────────────────────────────
const AVATAR_COLORS = [
  'bg-sage-deep', 'bg-terracotta', 'bg-sage',
  'bg-amber',     'bg-coral',      'bg-ink-soft',
]

function initials(name = '') {
  return name.trim().split(/\s+/).map(w => w[0]).join('').slice(0, 2).toUpperCase()
}

function avatarBg(name = '') {
  return AVATAR_COLORS[name.charCodeAt(0) % AVATAR_COLORS.length]
}

// L M M J V S D — Monday-first
const DAY_LABELS = ['L', 'M', 'M', 'J', 'V', 'S', 'D']

function dayStrip(checkin) {
  const dow = new Date().getDay()
  // Convert Sun=0…Sat=6 → Mon=0…Sun=6
  const todayIdx = dow === 0 ? 6 : dow - 1

  return DAY_LABELS.map((label, i) => {
    // TODO: replace `i < todayIdx → 'done'` with real 7-day history from API
    if (i < todayIdx) return { label, status: 'done' }
    if (i === todayIdx) return { label, status: checkin?.status ?? 'pending' }
    return { label, status: 'future' }
  })
}

const STATUS_STYLE = {
  done:    { strip: 'bg-sage',         icon: '✓', iconColor: 'text-sage-deep' },
  pending: { strip: 'bg-amber',        icon: '',  iconColor: '' },
  missed:  { strip: 'bg-coral',        icon: '✗', iconColor: 'text-coral-deep' },
  future:  { strip: 'border border-dashed border-hairline bg-transparent', icon: '', iconColor: '' },
}

const STATUS_PILL = {
  done:    { cls: 'bg-sage-soft text-sage-deep',    label: '✓ hoy' },
  pending: { cls: 'bg-amber-soft text-amber-deep',  label: 'pendiente' },
  missed:  { cls: 'bg-coral-soft text-coral-deep',  label: '✗ hoy' },
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────
let socketDisconnect = null

onMounted(async () => {
  ticker = setInterval(() => { now.value = new Date() }, 30_000)
  try {
    const found = await group.autoLoad()
    if (!found) {
      router.replace('/onboarding')
      return
    }
    await habits.loadToday(group.group.id)
    const memberIDs = [...new Set(habits.todayCheckins.map(c => c.user_id))]
    await Promise.all(memberIDs.map(id => habits.loadStreaks(id)))
    const { disconnect } = useGroupSocket(group.group.id)
    socketDisconnect = disconnect
    loaded.value = true
  } catch (_) {
    router.replace('/onboarding')
  }
})

onUnmounted(() => {
  clearInterval(ticker)
  socketDisconnect?.()
})
</script>

<template>
  <div class="max-w-md mx-auto px-4 pt-4 pb-6">

    <!-- ── Header ─────────────────────────────────────────────────────────── -->
    <header class="flex items-center justify-between mb-4">
      <span class="font-serif text-xl font-semibold text-terracotta tracking-wide">DYDI</span>

      <button class="flex items-center gap-1 text-sm font-bold text-ink rounded-pill
                     border border-hairline px-3 py-1.5 bg-surface">
        {{ group.group?.name ?? 'Mi grupo' }}
        <svg class="w-3.5 h-3.5 text-ink-soft" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="m19 9-7 7-7-7"/>
        </svg>
      </button>

      <div
        class="w-9 h-9 rounded-full flex items-center justify-center
               text-paper text-sm font-bold"
        :class="avatarBg(auth.user?.user_metadata?.display_name ?? '')"
      >
        {{ initials(auth.user?.user_metadata?.display_name ?? auth.user?.email ?? '') }}
      </div>
    </header>

    <!-- ── Live indicator ─────────────────────────────────────────────────── -->
    <div v-if="onlineAvatars.length > 0" class="flex items-center gap-2 mb-4">
      <!-- Stacked avatars -->
      <div class="flex -space-x-2">
        <div
          v-for="m in onlineAvatars"
          :key="m.user_id"
          class="w-6 h-6 rounded-full border-2 border-paper flex items-center
                 justify-center text-paper text-[9px] font-bold"
          :class="avatarBg(m.display_name)"
        >
          {{ initials(m.display_name) }}
        </div>
      </div>
      <span class="inline-flex items-center gap-1 text-eyebrow text-sage-deep">
        <span class="w-1.5 h-1.5 rounded-full bg-sage-deep animate-pulse"></span>
        EN VIVO
      </span>
      <span class="text-xs text-ink-soft">
        {{ onlineAvatars.length }} compas conectados ahora
      </span>
    </div>

    <!-- ── Countdown card ─────────────────────────────────────────────────── -->
    <div class="rounded-card bg-paper shadow-card p-5 mb-4">
      <div class="flex justify-between items-start mb-3">
        <span class="text-eyebrow">EL CICLO CIERRA EN</span>
        <span class="text-xs font-semibold text-terracotta">{{ closingLabel }}</span>
      </div>

      <div class="flex items-end gap-2 mb-4">
        <div class="text-center">
          <p class="font-serif text-5xl font-semibold text-terracotta leading-none">{{ countdown.days }}</p>
          <p class="text-[11px] text-ink-faint mt-1">días</p>
        </div>
        <span class="font-serif text-4xl text-hairline mb-2">:</span>
        <div class="text-center">
          <p class="font-serif text-5xl font-semibold text-terracotta leading-none">{{ countdown.hours }}</p>
          <p class="text-[11px] text-ink-faint mt-1">hrs</p>
        </div>
        <span class="font-serif text-4xl text-hairline mb-2">:</span>
        <div class="text-center">
          <p class="font-serif text-5xl font-semibold text-terracotta leading-none">{{ countdown.mins }}</p>
          <p class="text-[11px] text-ink-faint mt-1">min</p>
        </div>
      </div>

      <!-- Progress bar -->
      <div class="flex justify-between text-xs mb-1.5">
        <span class="text-ink-faint">Semana {{ weekNumber }}</span>
        <span class="text-terracotta font-semibold">
          {{ stats.done }} de {{ group.members.length || '—' }} al corriente
        </span>
      </div>
      <div class="h-1.5 rounded-full bg-hairline">
        <div
          class="h-full rounded-full bg-terracotta transition-all duration-500"
          :style="{ width: progressPct + '%' }"
        />
      </div>
    </div>

    <!-- ── My check-in card ───────────────────────────────────────────────── -->
    <div class="rounded-card shadow-card bg-paper p-5 mb-5">
      <div class="flex justify-between items-start mb-1">
        <span class="text-eyebrow">TU TURNO</span>
        <div class="text-right">
          <p class="font-serif text-3xl font-semibold leading-none text-terracotta">{{ myStreak }}</p>
          <p class="text-eyebrow text-terracotta mt-0.5">RACHA</p>
        </div>
      </div>

      <h2 class="font-serif text-2xl font-semibold text-ink mb-3 leading-snug">
        ¿Ya hiciste el tuyo?
      </h2>

      <!-- Habit list (one row per assigned habit) -->
      <div v-if="myCheckins.length" class="space-y-2 mb-4">
        <div
          v-for="c in myCheckins"
          :key="c.habit_id"
          class="flex flex-wrap items-center gap-2"
        >
          <span class="text-sm font-semibold text-ink">{{ c.habit_name }}</span>
          <span v-if="c.scheduled_time"
            class="rounded-pill bg-hairline px-2.5 py-0.5 text-xs text-ink-soft font-medium">
            {{ c.scheduled_time }}
          </span>
          <span
            class="rounded-pill px-2.5 py-0.5 text-xs font-semibold"
            :class="STATUS_PILL[c.status]?.cls ?? 'bg-hairline text-ink-soft'"
          >
            {{ STATUS_PILL[c.status]?.label ?? c.status }}
          </span>
        </div>
      </div>
      <p v-else class="text-sm text-ink-soft mb-4">
        No tienes un hábito registrado en este grupo todavía.
      </p>

      <!-- Action button -->
      <button
        v-if="hasPending || !myCheckins.length"
        class="w-full rounded-pill bg-sage-deep text-paper py-3.5 font-bold text-sm
               active:opacity-80 transition-opacity"
        @click="router.push('/checkin')"
      >
        Hacer mi check-in →
      </button>

      <div
        v-else-if="allDone"
        class="w-full rounded-pill bg-sage-soft text-sage-deep py-3.5 font-bold
               text-sm text-center"
      >
        ✓ &nbsp;Ya cumpliste hoy
      </div>

      <div
        v-else-if="anyMissed"
        class="w-full rounded-pill bg-coral-soft text-coral-deep py-3.5 font-bold
               text-sm text-center"
      >
        Se te fue el día
      </div>
    </div>

    <!-- ── Summary numbers ────────────────────────────────────────────────── -->
    <div class="rounded-card bg-paper shadow-flat grid grid-cols-3 text-center mb-6 overflow-hidden">
      <div class="py-4">
        <p class="font-serif text-3xl font-semibold text-sage-deep">{{ stats.done }}</p>
        <p class="text-xs text-ink-soft mt-0.5">cumplieron</p>
      </div>
      <div class="border-x border-hairline py-4">
        <p class="font-serif text-3xl font-semibold text-amber-deep">{{ stats.pending }}</p>
        <p class="text-xs text-ink-soft mt-0.5">pendientes</p>
      </div>
      <div class="py-4">
        <p class="font-serif text-3xl font-semibold text-coral-deep">{{ stats.missed }}</p>
        <p class="text-xs text-ink-soft mt-0.5">fallaron</p>
      </div>
    </div>

    <!-- ── Squad hoy ──────────────────────────────────────────────────────── -->
    <div>
      <div class="flex justify-between items-center mb-3">
        <h3 class="font-semibold text-sm text-ink">El squad hoy</h3>
        <span class="text-xs text-ink-soft">Lun → Dom</span>
      </div>

      <!-- Empty / loading state -->
      <div v-if="squadRows.length === 0"
        class="rounded-card bg-surface py-10 text-center text-sm text-ink-soft">
        <span v-if="!loaded">Cargando el squad…</span>
        <span v-else>Propón un hábito en la pestaña Votar para ver al squad aquí.</span>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="row in squadRows"
          :key="row.user_id"
          class="rounded-card bg-surface p-4 flex items-start gap-3"
        >
          <!-- Avatar -->
          <div
            class="w-10 h-10 rounded-full flex-shrink-0 flex items-center justify-center
                   text-paper text-sm font-bold"
            :class="avatarBg(row.display_name ?? '')"
          >
            {{ initials(row.display_name ?? '') }}
          </div>

          <!-- Info -->
          <div class="flex-1 min-w-0">
            <div class="flex justify-between items-center mb-0.5">
              <div class="flex items-baseline gap-1.5">
                <span class="font-semibold text-sm text-ink truncate">{{ row.display_name }}</span>
                <span class="text-xs text-terracotta font-medium">
                  ★ {{ habits.streaks[row.user_id] ?? 0 }}
                </span>
              </div>
              <span
                v-if="STATUS_PILL[row.status]"
                class="rounded-pill px-2 py-0.5 text-[10px] font-semibold ml-2 flex-shrink-0"
                :class="STATUS_PILL[row.status].cls"
              >
                {{ STATUS_PILL[row.status].label }}
              </span>
            </div>

            <p class="text-xs text-ink-soft mb-2 truncate">{{ row.habit_name }}</p>

            <!-- 7-day strip -->
            <div class="flex gap-1">
              <div
                v-for="(day, i) in dayStrip(row)"
                :key="i"
                class="flex flex-col items-center gap-0.5"
              >
                <div
                  class="w-7 h-7 rounded-md flex items-center justify-center"
                  :class="STATUS_STYLE[day.status].strip"
                >
                  <span
                    v-if="STATUS_STYLE[day.status].icon"
                    class="text-xs font-bold"
                    :class="STATUS_STYLE[day.status].iconColor"
                  >
                    {{ STATUS_STYLE[day.status].icon }}
                  </span>
                </div>
                <span class="text-[9px] text-ink-faint font-medium">{{ day.label }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>
