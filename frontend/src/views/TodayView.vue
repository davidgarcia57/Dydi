<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'

const router = useRouter()
const auth  = useAuthStore()
const group = useGroupStore()
const habits = useHabitsStore()

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
const myCheckin = computed(() =>
  habits.todayCheckins.find(c => c.user_id === auth.user?.id)
)

const myStreak = computed(() => habits.streaks[auth.user?.id] ?? 0)

// ── Squad stats ───────────────────────────────────────────────────────────────
const stats = computed(() => ({
  done:    habits.todayCheckins.filter(c => c.status === 'done').length,
  pending: habits.todayCheckins.filter(c => c.status === 'pending').length,
  missed:  habits.todayCheckins.filter(c => c.status === 'missed').length,
}))

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
  done:    { strip: 'bg-sage',     icon: '✓', iconColor: 'text-sage-deep' },
  pending: { strip: 'bg-amber',    icon: '',  iconColor: '' },
  missed:  { strip: 'bg-coral',    icon: '✗', iconColor: 'text-coral' },
  future:  { strip: 'bg-hairline', icon: '',  iconColor: '' },
}

const STATUS_PILL = {
  done:    { cls: 'bg-sage/30 text-sage-deep',    label: '✓ hoy' },
  pending: { cls: 'bg-amber/30 text-amber',       label: 'pendiente' },
  missed:  { cls: 'bg-coral/30 text-coral',       label: '✗ hoy' },
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────
onMounted(async () => {
  ticker = setInterval(() => { now.value = new Date() }, 30_000)
  await group.autoLoad()
  if (group.group?.id) {
    await habits.loadToday(group.group.id)
    const memberIDs = [...new Set(habits.todayCheckins.map(c => c.user_id))]
    await Promise.all(memberIDs.map(id => habits.loadStreaks(id)))
  }
})

onUnmounted(() => clearInterval(ticker))
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
    <div class="rounded-card bg-[#1C2E28] p-5 mb-4">
      <div class="flex justify-between items-start mb-3">
        <span class="text-eyebrow text-[#7AAF9E]">EL CICLO CIERRA EN</span>
        <span class="text-xs font-semibold text-terracotta">{{ closingLabel }}</span>
      </div>

      <div class="flex items-end gap-2 mb-4">
        <div class="text-center">
          <p class="font-serif text-5xl font-semibold text-paper leading-none">{{ countdown.days }}</p>
          <p class="text-[11px] text-[#7AAF9E] mt-1">días</p>
        </div>
        <span class="font-serif text-4xl text-[#4A6E62] mb-2">:</span>
        <div class="text-center">
          <p class="font-serif text-5xl font-semibold text-paper leading-none">{{ countdown.hours }}</p>
          <p class="text-[11px] text-[#7AAF9E] mt-1">hrs</p>
        </div>
        <span class="font-serif text-4xl text-[#4A6E62] mb-2">:</span>
        <div class="text-center">
          <p class="font-serif text-5xl font-semibold text-paper leading-none">{{ countdown.mins }}</p>
          <p class="text-[11px] text-[#7AAF9E] mt-1">min</p>
        </div>
      </div>

      <!-- Progress bar -->
      <div class="flex justify-between text-xs mb-1.5">
        <span class="text-[#7AAF9E]">Semana {{ weekNumber }}</span>
        <span class="text-terracotta font-semibold">
          {{ stats.done }} de {{ group.members.length || '—' }} al corriente
        </span>
      </div>
      <div class="h-1.5 rounded-full bg-white/10">
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

      <!-- Habit info -->
      <div v-if="myCheckin" class="flex flex-wrap items-center gap-2 mb-4">
        <span class="text-sm font-semibold text-ink">{{ myCheckin.habit_name }}</span>
        <span v-if="myCheckin.scheduled_time"
          class="rounded-pill bg-hairline px-2.5 py-0.5 text-xs text-ink-soft font-medium">
          {{ myCheckin.scheduled_time }}
        </span>
        <span
          class="rounded-pill px-2.5 py-0.5 text-xs font-semibold"
          :class="STATUS_PILL[myCheckin.status]?.cls ?? 'bg-hairline text-ink-soft'"
        >
          {{ STATUS_PILL[myCheckin.status]?.label ?? myCheckin.status }}
        </span>
      </div>
      <p v-else class="text-sm text-ink-soft mb-4">
        No tienes un hábito registrado en este grupo todavía.
      </p>

      <!-- Action button -->
      <button
        v-if="!myCheckin || myCheckin.status === 'pending'"
        class="w-full rounded-pill bg-sage-deep text-paper py-3.5 font-bold text-sm
               active:opacity-80 transition-opacity"
        @click="router.push('/checkin')"
      >
        Hacer mi check-in →
      </button>

      <div
        v-else-if="myCheckin.status === 'done'"
        class="w-full rounded-pill bg-sage/20 text-sage-deep py-3.5 font-bold
               text-sm text-center"
      >
        ✓ &nbsp;Ya cumpliste hoy
      </div>

      <div
        v-else-if="myCheckin.status === 'missed'"
        class="w-full rounded-pill bg-coral/20 text-coral py-3.5 font-bold
               text-sm text-center"
      >
        Se te fue el día
      </div>
    </div>

    <!-- ── Summary numbers ────────────────────────────────────────────────── -->
    <div class="grid grid-cols-3 text-center mb-6">
      <div>
        <p class="font-serif text-3xl font-semibold text-ink">{{ stats.done }}</p>
        <p class="text-xs text-ink-soft mt-0.5">cumplieron</p>
      </div>
      <div class="border-x border-hairline">
        <p class="font-serif text-3xl font-semibold text-ink">{{ stats.pending }}</p>
        <p class="text-xs text-ink-soft mt-0.5">pendientes</p>
      </div>
      <div>
        <p class="font-serif text-3xl font-semibold text-ink">{{ stats.missed }}</p>
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
        Cargando el squad...
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
