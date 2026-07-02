<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import BrandWordmark from '@/components/ui/BrandWordmark.vue'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { useGroupSocket } from '@/composables/useGroupSocket'
import { showToast } from '@/composables/useToast'
import PageContainer from '@/components/ui/PageContainer.vue'

const router = useRouter()
const auth = useAuthStore()
const group = useGroupStore()
const habits = useHabitsStore()
const loaded = ref(false)
const loadError = ref(false)

async function shareInvite() {
  if (!group.group?.invite_code || !group.group?.id) return
  // Join expects the full "{groupID}:{inviteCode}" code (see OnboardingView).
  const code = `${group.group.id}:${group.group.invite_code}`
  const text = `¡Únete a mi squad "${group.group.name}" en Dydi!\nCódigo de invitación: ${code}`
  if (navigator.share) {
    try {
      await navigator.share({
        title: 'Únete a Dydi',
        text: text,
      })
    } catch (e) {}
  } else {
    await navigator.clipboard.writeText(code)
    showToast(`Código copiado: ${code}`)
  }
}

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
  const pad = (n) => String(n).padStart(2, '0')
  return {
    days: pad(Math.floor(diff / 86_400_000)),
    hours: pad(Math.floor((diff % 86_400_000) / 3_600_000)),
    mins: pad(Math.floor((diff % 3_600_000) / 60_000)),
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
const myCheckins = computed(() => habits.todayCheckins.filter((c) => c.user_id === auth.user?.id))

const hasPending = computed(() => myCheckins.value.some((c) => c.status === 'pending'))
const allDone = computed(
  () => myCheckins.value.length > 0 && myCheckins.value.every((c) => c.status === 'done')
)
const anyMissed = computed(
  () => myCheckins.value.some((c) => c.status === 'missed') && !hasPending.value
)

const myStreak = computed(() => habits.streaks[auth.user?.id] ?? 0)

// ── Squad stats (per member, not per checkin) ─────────────────────────────────
const stats = computed(() => {
  const byUser = {}
  for (const c of habits.todayCheckins) {
    if (!byUser[c.user_id]) byUser[c.user_id] = []
    byUser[c.user_id].push(c.status)
  }
  let done = 0,
    pending = 0,
    missed = 0
  for (const statuses of Object.values(byUser)) {
    if (statuses.every((s) => s === 'done')) done++
    else if (statuses.some((s) => s === 'pending')) pending++
    else missed++
  }
  return { done, pending, missed }
})

const progressPct = computed(() => {
  const total = group.members.length
  return total ? Math.round((stats.value.done / total) * 100) : 0
})

// ── Online members ────────────────────────────────────────────────────────────
const onlineAvatars = computed(() =>
  group.members.filter((m) => group.onlineMembers.has(m.user_id)).slice(0, 5)
)

// ── Helpers ───────────────────────────────────────────────────────────────────
const AVATAR_COLORS = [
  'bg-sage-deep',
  'bg-terracotta',
  'bg-sage',
  'bg-amber',
  'bg-coral',
  'bg-ink-soft',
]

function initials(name = '') {
  return name
    .trim()
    .split(/\s+/)
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

function avatarBg(name = '') {
  return AVATAR_COLORS[name.charCodeAt(0) % AVATAR_COLORS.length]
}

const STATUS_PILL = {
  done: { cls: 'bg-sage-soft text-sage-deep', label: '✓ hoy' },
  pending: { cls: 'bg-amber-soft text-amber-deep', label: 'pendiente' },
  missed: { cls: 'bg-coral-soft text-coral-deep', label: '✗ hoy' },
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────
let socketDisconnect = null

async function load() {
  loadError.value = false
  try {
    const found = await group.autoLoad()
    // No group at all → onboarding. A *failure* to load is a different thing.
    if (!found) {
      router.replace('/onboarding')
      return
    }
    await habits.loadToday(group.group.id)
    const memberIDs = [...new Set(habits.todayCheckins.map((c) => c.user_id))]
    await Promise.all(memberIDs.map((id) => habits.loadStreaks(id)))
    const { disconnect } = useGroupSocket(group.group.id)
    socketDisconnect = disconnect
    loaded.value = true
  } catch (_) {
    loadError.value = true
  }
}

onMounted(() => {
  ticker = setInterval(() => {
    now.value = new Date()
  }, 30_000)
  load()
})

onUnmounted(() => {
  clearInterval(ticker)
  socketDisconnect?.()
})
</script>

<template>
  <PageContainer>
    <!-- ── Error state ────────────────────────────────────────────────────── -->
    <div
      v-if="loadError"
      class="rounded-card bg-coral-soft/40 border border-coral/40 p-4 mb-4 flex flex-wrap items-center justify-between gap-3"
    >
      <p class="text-sm font-medium text-coral-deep">
        No pudimos cargar tu grupo. Revisa tu conexión.
      </p>
      <button
        class="rounded-pill bg-coral text-paper px-4 py-2 text-sm font-bold active:opacity-80 transition-opacity"
        @click="load"
      >
        Reintentar
      </button>
    </div>

    <!-- ── Header ─────────────────────────────────────────────────────────── -->
    <header class="flex items-center justify-between mb-4">
      <BrandWordmark size="sm" />

      <button
        class="flex items-center gap-1.5 text-sm font-bold text-ink rounded-pill border border-hairline px-3 py-1.5 bg-surface active:opacity-70 transition-opacity"
        @click="shareInvite"
      >
        {{ group.group?.name ?? 'Mi grupo' }}
        <svg
          class="w-3.5 h-3.5 text-ink-soft"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2.5"
            d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
          />
        </svg>
      </button>

      <div
        class="w-9 h-9 rounded-full flex items-center justify-center text-paper text-sm font-bold"
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
          class="w-6 h-6 rounded-full border-2 border-paper flex items-center justify-center text-paper text-[9px] font-bold"
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
    <div class="lg:grid lg:grid-cols-3 lg:gap-6 lg:items-start">
      <div class="lg:col-span-2">
        <div class="rounded-card bg-paper shadow-card p-5 mb-4">
          <div class="flex justify-between items-start mb-3">
            <span class="text-eyebrow">EL CICLO CIERRA EN</span>
            <span class="text-xs font-semibold text-terracotta">{{ closingLabel }}</span>
          </div>

          <div class="flex items-end gap-2 mb-4">
            <div class="text-center">
              <p class="font-serif text-5xl font-semibold text-terracotta leading-none">
                {{ countdown.days }}
              </p>
              <p class="text-[11px] text-ink-faint mt-1">días</p>
            </div>
            <span class="font-serif text-4xl text-hairline mb-2">:</span>
            <div class="text-center">
              <p class="font-serif text-5xl font-semibold text-terracotta leading-none">
                {{ countdown.hours }}
              </p>
              <p class="text-[11px] text-ink-faint mt-1">hrs</p>
            </div>
            <span class="font-serif text-4xl text-hairline mb-2">:</span>
            <div class="text-center">
              <p class="font-serif text-5xl font-semibold text-terracotta leading-none">
                {{ countdown.mins }}
              </p>
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
              <p class="font-serif text-3xl font-semibold leading-none text-terracotta">
                {{ myStreak }}
              </p>
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
              <span
                v-if="c.scheduled_time"
                class="rounded-pill bg-hairline px-2.5 py-0.5 text-xs text-ink-soft font-medium"
              >
                {{ c.scheduled_time }}
              </span>
              <span
                class="rounded-pill px-2.5 py-0.5 text-xs font-semibold"
                :class="STATUS_PILL[c.status]?.cls ?? 'bg-hairline text-ink-soft'"
              >
                {{ STATUS_PILL[c.status]?.label ?? c.status }}
              </span>
              <p v-if="c.note" class="w-full text-xs text-ink-soft italic mt-0.5">“{{ c.note }}”</p>
            </div>
          </div>
          <p v-else class="text-sm text-ink-soft mb-4">
            No tienes un hábito registrado en este grupo todavía.
          </p>

          <!-- Action button -->
          <button
            v-if="hasPending || !myCheckins.length"
            class="w-full rounded-pill bg-sage-deep text-paper py-3.5 font-bold text-sm active:opacity-80 transition-opacity"
            @click="router.push('/checkin')"
          >
            Hacer mi check-in →
          </button>

          <div
            v-else-if="allDone"
            class="w-full rounded-pill bg-sage-soft text-sage-deep py-3.5 font-bold text-sm text-center"
          >
            ✓ &nbsp;Ya cumpliste hoy
          </div>

          <div
            v-else-if="anyMissed"
            class="w-full rounded-pill bg-coral-soft text-coral-deep py-3.5 font-bold text-sm text-center"
          >
            Se te fue el día
          </div>
        </div>
      </div>

      <!-- ── El squad (resumen) ─────────────────────────────────────────────── -->
      <div class="lg:col-span-1">
        <div class="rounded-card bg-paper shadow-flat p-5">
          <div class="flex justify-between items-center mb-4">
            <h3 class="font-semibold text-sm text-ink">El squad</h3>
            <RouterLink
              to="/squad"
              class="text-xs font-semibold text-sage-deep hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sage-deep/50 rounded"
            >
              Ver squad →
            </RouterLink>
          </div>

          <!-- Online now -->
          <div v-if="onlineAvatars.length" class="flex items-center gap-2 mb-5">
            <div class="flex -space-x-2">
              <div
                v-for="m in onlineAvatars"
                :key="m.user_id"
                class="w-7 h-7 rounded-full border-2 border-paper flex items-center justify-center text-paper text-[10px] font-bold"
                :class="avatarBg(m.display_name)"
              >
                {{ initials(m.display_name) }}
              </div>
            </div>
            <span class="text-xs text-ink-soft">{{ onlineAvatars.length }} en línea</span>
          </div>
          <p v-else class="text-xs text-ink-faint mb-5">Nadie conectado ahora mismo.</p>

          <!-- Squad pulse -->
          <div class="grid grid-cols-3 text-center rounded-card bg-surface overflow-hidden">
            <div class="py-3">
              <p class="font-serif text-2xl font-semibold text-sage-deep leading-none">
                {{ stats.done }}
              </p>
              <p class="text-[10px] text-ink-soft mt-1">al corriente</p>
            </div>
            <div class="border-x border-hairline py-3">
              <p class="font-serif text-2xl font-semibold text-amber-deep leading-none">
                {{ stats.pending }}
              </p>
              <p class="text-[10px] text-ink-soft mt-1">pendientes</p>
            </div>
            <div class="py-3">
              <p class="font-serif text-2xl font-semibold text-coral-deep leading-none">
                {{ stats.missed }}
              </p>
              <p class="text-[10px] text-ink-soft mt-1">fallaron</p>
            </div>
          </div>

          <p v-if="!loaded" class="text-xs text-ink-faint mt-4 text-center">Cargando el squad…</p>
        </div>
      </div>
    </div>
  </PageContainer>
</template>
