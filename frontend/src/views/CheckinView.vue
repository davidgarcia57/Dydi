<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'

const router = useRouter()
const auth = useAuthStore()
const group = useGroupStore()
const habits = useHabitsStore()

// 'loading' | 'error' | 'no-habit' | 'select' | 'confirm' | 'success' | 'done'
const step = ref('loading')
const selected = ref(null)
const note = ref('')
const submitting = ref(false)
const errMsg = ref('')
const prevStreak = ref(0)
const newStreak = ref(0)
const showPlus = ref(false)

// Live clock
const currentTime = ref(formatTime())
let clockTick = null

function formatTime() {
  const d = new Date()
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const myHabits = computed(() => habits.todayCheckins.filter((c) => c.user_id === auth.user?.id))
const myPending = computed(() => myHabits.value.filter((c) => c.status === 'pending'))

async function load() {
  step.value = 'loading'
  errMsg.value = ''
  try {
    const found = await group.autoLoad()
    if (found && group.group?.id) {
      // Always refresh — relying on cached todayCheckins can show a stale day.
      await habits.loadToday(group.group.id)
    }
    await habits.loadStreaks(auth.user?.id)
    prevStreak.value = habits.streaks[auth.user?.id] ?? 0
    resolve()
  } catch (e) {
    errMsg.value = e?.message || 'No pudimos cargar tus hábitos.'
    step.value = 'error'
  }
}

onMounted(() => {
  clockTick = setInterval(() => {
    currentTime.value = formatTime()
  }, 10_000)
  load()
})

onUnmounted(() => clearInterval(clockTick))

function resolve() {
  if (myHabits.value.length === 0) {
    step.value = 'no-habit'
  } else if (myPending.value.length === 0) {
    step.value = 'done'
  } else if (myPending.value.length === 1) {
    selected.value = myPending.value[0]
    step.value = 'confirm'
  } else {
    step.value = 'select'
  }
}

function pick(habit) {
  selected.value = habit
  step.value = 'confirm'
}

async function submit() {
  if (!selected.value || submitting.value) return
  submitting.value = true
  errMsg.value = ''
  try {
    await habits.checkin(group.group.id, selected.value.habit_id, note.value.trim())
    await habits.loadStreaks(auth.user?.id)
    newStreak.value = habits.streaks[auth.user?.id] ?? 0
    step.value = 'success'
    setTimeout(() => {
      showPlus.value = true
    }, 600)
  } catch (e) {
    errMsg.value = e?.message || 'Algo salió mal, intenta de nuevo.'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-cream flex flex-col relative">
    <!-- ── Loading ─────────────────────────────────────────────────────────── -->
    <div v-if="step === 'loading'" class="flex-1 flex items-center justify-center">
      <div
        class="w-8 h-8 rounded-full border-2 border-sage-deep border-t-transparent animate-spin"
      />
    </div>

    <!-- ── Error ──────────────────────────────────────────────────────────── -->
    <div
      v-else-if="step === 'error'"
      class="flex-1 flex flex-col items-center justify-center px-8 text-center"
    >
      <p class="text-eyebrow text-coral mb-2">ALGO FALLÓ</p>
      <h1 class="font-serif text-2xl font-semibold text-ink mb-2">No pudimos cargar tus hábitos</h1>
      <p class="text-sm text-ink-soft mb-8">{{ errMsg }}</p>
      <button
        class="rounded-pill bg-sage-deep text-paper px-8 py-3.5 font-bold text-sm"
        @click="load"
      >
        Reintentar
      </button>
    </div>

    <!-- ── No habit ───────────────────────────────────────────────────────── -->
    <div
      v-else-if="step === 'no-habit'"
      class="flex-1 flex flex-col items-center justify-center px-8 text-center"
    >
      <div class="w-20 h-20 rounded-full bg-amber-soft flex items-center justify-center mb-6">
        <svg
          class="w-10 h-10 text-amber-deep"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71
               c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898
               0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
          />
        </svg>
      </div>
      <p class="text-eyebrow text-amber-deep mb-2">SIN HÁBITO</p>
      <h1 class="font-serif text-3xl font-semibold text-ink mb-2">No tienes un hábito asignado</h1>
      <p class="text-sm text-ink-soft mb-8">
        Pídele al squad que proponga uno — se aprueba por votación.
      </p>
      <button
        class="rounded-pill bg-ink text-paper px-8 py-3.5 font-bold text-sm"
        @click="router.replace('/today')"
      >
        Volver
      </button>
    </div>

    <!-- ── All done ─────────────────────────────────────────────────────────── -->
    <div
      v-else-if="step === 'done'"
      class="flex-1 flex flex-col items-center justify-center px-8 text-center"
    >
      <div class="w-20 h-20 rounded-full bg-sage-soft flex items-center justify-center mb-6">
        <svg class="w-10 h-10 text-sage-deep" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2.5"
            d="M5 13l4 4L19 7"
          />
        </svg>
      </div>
      <p class="text-eyebrow text-sage-deep mb-2">HOY YA CUMPLISTE</p>
      <h1 class="font-serif text-3xl font-semibold text-ink mb-1">
        {{ myHabits[0]?.habit_name ?? 'Tu hábito' }}
      </h1>
      <p class="text-sm text-ink-soft mb-8">
        Racha actual:
        <span class="font-bold text-terracotta">{{ habits.streaks[auth.user?.id] ?? 0 }} días</span>
      </p>
      <button
        class="rounded-pill bg-ink text-paper px-8 py-3.5 font-bold text-sm"
        @click="router.replace('/today')"
      >
        Volver al squad
      </button>
    </div>

    <!-- ── Select habit ──────────────────────────────────────────────────── -->
    <div v-else-if="step === 'select'" class="flex-1 flex flex-col">
      <!-- Header -->
      <div class="px-6 pt-12 pb-6">
        <button
          class="w-9 h-9 rounded-full bg-surface border border-hairline flex items-center justify-center mb-8"
          @click="router.back()"
        >
          <svg class="w-4 h-4 text-ink" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2.5"
              d="M6 18 18 6M6 6l12 12"
            />
          </svg>
        </button>
        <p class="text-eyebrow mb-2">ELIGE TU HÁBITO DE HOY</p>
        <h1 class="font-serif text-3xl font-semibold text-ink leading-snug">¿Cuál cumpliste?</h1>
      </div>

      <div class="px-6 space-y-3 pb-8">
        <button
          v-for="h in myPending"
          :key="h.habit_id"
          class="w-full rounded-card bg-paper shadow-card p-5 text-left flex items-center gap-4 active:scale-[0.98] transition-transform"
          @click="pick(h)"
        >
          <div
            class="w-12 h-12 rounded-full flex-shrink-0 flex items-center justify-center text-xl font-bold text-paper"
            :style="{ backgroundColor: h.color || '#A8C39A' }"
          >
            {{ h.habit_name?.charAt(0).toUpperCase() }}
          </div>
          <div class="flex-1 min-w-0">
            <p class="font-semibold text-ink truncate">{{ h.habit_name }}</p>
            <p v-if="h.scheduled_time" class="text-xs text-ink-soft mt-0.5">
              {{ h.scheduled_time }}
            </p>
          </div>
          <svg
            class="w-5 h-5 text-ink-faint flex-shrink-0"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="m9 18 6-6-6-6"
            />
          </svg>
        </button>
      </div>
    </div>

    <!-- ── Confirm ──────────────────────────────────────────────────────────── -->
    <div v-else-if="step === 'confirm'" class="flex-1 flex flex-col">
      <!-- Top bar -->
      <div class="flex items-center justify-between px-6 pt-12 mb-10">
        <button
          class="w-9 h-9 rounded-full bg-surface border border-hairline flex items-center justify-center"
          @click="myPending.length > 1 ? (step = 'select') : router.back()"
        >
          <svg class="w-4 h-4 text-ink" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2.5"
              d="M6 18 18 6M6 6l12 12"
            />
          </svg>
        </button>
        <!-- Racha chip -->
        <div
          class="flex items-center gap-1.5 rounded-pill bg-amber-soft text-amber-deep px-3 py-1.5 text-xs font-bold"
        >
          <span>★</span>
          <span>{{ prevStreak }} días de racha</span>
        </div>
      </div>

      <!-- Eyebrow + habit name -->
      <div class="px-8 mb-8">
        <p class="text-eyebrow mb-3">TU HÁBITO DE HOY</p>
        <h1 class="font-serif text-4xl font-semibold text-ink leading-tight mb-5">
          {{ selected?.habit_name }}
        </h1>

        <!-- Time + deadline chips -->
        <div class="flex items-center gap-2 flex-wrap">
          <div
            class="flex items-center gap-1.5 rounded-full bg-surface border border-hairline text-ink-soft px-3 py-1.5 text-xs font-semibold"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2.5"
                d="M12 6v6l4 2m6-2a10 10 0 11-20 0 10 10 0 0120 0z"
              />
            </svg>
            {{ currentTime }}
          </div>
          <div
            v-if="selected?.scheduled_time"
            class="rounded-full bg-amber-soft text-amber-deep px-3 py-1.5 text-xs font-semibold"
          >
            Meta: {{ selected.scheduled_time }}
          </div>
          <div
            v-else
            class="rounded-full bg-sage-soft text-sage-deep px-3 py-1.5 text-xs font-semibold"
          >
            Hoy hasta medianoche
          </div>
        </div>
      </div>

      <!-- BIG CHECK BUTTON — centered -->
      <div class="flex-1 flex flex-col items-center justify-center px-8">
        <button
          :disabled="submitting"
          class="w-24 h-24 rounded-full bg-sage-deep shadow-card flex items-center justify-center mb-6 active:scale-95 transition-all duration-150 disabled:opacity-60 relative overflow-hidden"
          @click="submit"
        >
          <svg
            v-if="!submitting"
            class="w-12 h-12 text-paper"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="3"
              d="M5 13l4 4L19 7"
            />
          </svg>
          <span
            v-else
            class="w-8 h-8 rounded-full border-[3px] border-paper border-t-transparent animate-spin"
          />
        </button>

        <p class="text-sm font-semibold text-ink-soft mb-1">
          {{ submitting ? 'Registrando…' : 'Toca para registrar' }}
        </p>
        <p class="text-xs text-ink-faint">El squad verá tu check-in al instante</p>

        <!-- Optional note -->
        <div class="mt-8 w-full">
          <label class="block">
            <span class="text-eyebrow mb-2 block">NOTA OPCIONAL</span>
            <textarea
              v-model="note"
              rows="2"
              placeholder="¿Algo que quieras contarle al squad?"
              class="w-full rounded-[14px] border border-hairline bg-surface px-4 py-3 text-sm text-ink placeholder-ink-faint resize-none focus:outline-none focus:border-sage-deep transition-colors"
            />
          </label>
        </div>

        <p v-if="errMsg" class="text-sm text-coral mt-4 font-medium">{{ errMsg }}</p>
      </div>

      <div class="h-10" />
    </div>

    <!-- ── Success ──────────────────────────────────────────────────────────── -->
    <div
      v-else-if="step === 'success'"
      class="flex-1 flex flex-col items-center justify-center px-8 text-center"
    >
      <!-- Check ring (animated entrance) -->
      <div class="relative mb-8">
        <!-- Outer glow ring -->
        <div
          class="w-32 h-32 rounded-full bg-sage-soft flex items-center justify-center animate-[ping_0.6s_ease-out_1]"
        />
        <div
          class="absolute inset-0 w-32 h-32 rounded-full bg-sage-soft flex items-center justify-center"
        >
          <div
            class="w-22 h-22 rounded-full bg-sage-deep flex items-center justify-center"
            style="width: 5.5rem; height: 5.5rem"
          >
            <svg class="w-10 h-10 text-paper" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="3"
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>
        </div>
      </div>

      <p class="text-eyebrow text-sage-deep mb-2">¡LO LOGRASTE!</p>
      <h1 class="font-serif text-3xl font-semibold text-ink mb-1">
        {{ selected?.habit_name }}
      </h1>
      <p class="text-sm text-ink-soft mb-6">Check-in registrado hoy</p>

      <!-- Streak card with +1 badge -->
      <div class="relative rounded-card bg-paper shadow-card px-8 py-6 mb-10 w-full max-w-xs">
        <!-- +1 badge -->
        <div
          class="absolute -top-3 -right-2 rounded-full bg-terracotta text-paper text-xs font-bold px-2.5 py-1 shadow-flat transition-all duration-500"
          :class="showPlus ? 'opacity-100 translate-y-0' : 'opacity-0 -translate-y-2'"
        >
          +1 🔥
        </div>

        <p class="text-eyebrow mb-2">TU RACHA</p>
        <p class="font-serif text-6xl font-semibold text-terracotta leading-none mb-1">
          {{ newStreak }}
        </p>
        <p class="text-sm text-ink-soft">
          {{ newStreak === 1 ? 'día — ¡arrancaste!' : `días seguidos` }}
        </p>
      </div>

      <button
        class="w-full max-w-xs rounded-pill bg-ink text-paper py-4 font-bold text-sm"
        @click="router.replace('/today')"
      >
        Ver el squad →
      </button>
    </div>
  </div>
</template>
