<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'

const router = useRouter()
const auth   = useAuthStore()
const group  = useGroupStore()
const habits = useHabitsStore()

// 'loading' | 'select' | 'confirm' | 'success' | 'done'
const step      = ref('loading')
const selected  = ref(null)
const note      = ref('')
const submitting = ref(false)
const errMsg    = ref('')
const newStreak = ref(0)

const myHabits = computed(() =>
  habits.todayCheckins.filter(c => c.user_id === auth.user?.id)
)
const myPending = computed(() =>
  myHabits.value.filter(c => c.status === 'pending')
)

onMounted(async () => {
  await group.autoLoad()
  if (group.group?.id && habits.todayCheckins.length === 0) {
    await habits.loadToday(group.group.id)
  }
  resolve()
})

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
  } catch (e) {
    errMsg.value = e?.message || 'Algo salió mal, intenta de nuevo.'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-cream flex flex-col">

    <!-- Close button -->
    <button
      v-if="step !== 'success'"
      class="absolute top-4 left-4 z-10 w-9 h-9 rounded-full bg-surface
             flex items-center justify-center text-ink-soft shadow-flat"
      @click="router.back()"
    >
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5"
          d="M6 18 18 6M6 6l12 12"/>
      </svg>
    </button>

    <!-- ── Loading ─────────────────────────────────────────────────────────── -->
    <div v-if="step === 'loading'"
      class="flex-1 flex items-center justify-center">
      <div class="w-8 h-8 rounded-full border-2 border-sage-deep border-t-transparent animate-spin" />
    </div>

    <!-- ── No habit assigned ──────────────────────────────────────────────── -->
    <div v-else-if="step === 'no-habit'"
      class="flex-1 flex flex-col items-center justify-center px-8 text-center">
      <div class="w-20 h-20 rounded-full bg-amber/20 flex items-center justify-center mb-6">
        <svg class="w-10 h-10 text-amber" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71
               c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898
               0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"/>
        </svg>
      </div>
      <p class="text-eyebrow tracking-eyebrow text-amber mb-2">SIN HÁBITO</p>
      <h1 class="font-serif text-3xl font-semibold text-ink mb-2">
        No tienes un hábito asignado
      </h1>
      <p class="text-sm text-ink-soft mb-8">
        Pídele al admin del grupo que te asigne uno.
      </p>
      <button
        class="rounded-pill bg-ink text-paper px-8 py-3.5 font-bold text-sm"
        @click="router.replace('/today')"
      >
        Volver
      </button>
    </div>

    <!-- ── All done ────────────────────────────────────────────────────────── -->
    <div v-else-if="step === 'done'"
      class="flex-1 flex flex-col items-center justify-center px-8 text-center">
      <div class="w-20 h-20 rounded-full bg-sage/20 flex items-center justify-center mb-6">
        <svg class="w-10 h-10 text-sage-deep" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5"
            d="M5 13l4 4L19 7"/>
        </svg>
      </div>
      <p class="text-eyebrow tracking-eyebrow text-sage-deep mb-2">HOY YA CUMPLISTE</p>
      <h1 class="font-serif text-3xl font-semibold text-ink mb-2">
        {{ myHabits.length > 0 ? myHabits[0].habit_name : 'Tu hábito' }}
      </h1>
      <p class="text-sm text-ink-soft mb-8">
        Racha actual:
        <span class="font-bold text-terracotta">
          {{ habits.streaks[auth.user?.id] ?? 0 }} días
        </span>
      </p>
      <button
        class="rounded-pill bg-ink text-paper px-8 py-3.5 font-bold text-sm"
        @click="router.replace('/today')"
      >
        Volver al squad
      </button>
    </div>

    <!-- ── Select habit ────────────────────────────────────────────────────── -->
    <div v-else-if="step === 'select'"
      class="flex-1 flex flex-col px-6 pt-16 pb-8">
      <p class="text-eyebrow tracking-eyebrow text-ink-soft mb-3">ELIGE TU HÁBITO</p>
      <h1 class="font-serif text-3xl font-semibold text-ink mb-8 leading-snug">
        ¿Cuál cumpliste hoy?
      </h1>

      <div class="space-y-3 flex-1">
        <button
          v-for="h in myPending"
          :key="h.habit_id"
          class="w-full rounded-card bg-paper shadow-card p-5 text-left
                 flex items-center gap-4 active:scale-[0.98] transition-transform"
          @click="pick(h)"
        >
          <!-- Color dot -->
          <div
            class="w-12 h-12 rounded-full flex-shrink-0 flex items-center justify-center
                   text-xl font-bold text-paper"
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
          <svg class="w-5 h-5 text-ink-faint flex-shrink-0" fill="none"
            viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="m9 18 6-6-6-6"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- ── Confirm check-in ────────────────────────────────────────────────── -->
    <div v-else-if="step === 'confirm'"
      class="flex-1 flex flex-col px-6 pt-16 pb-8">

      <p class="text-eyebrow tracking-eyebrow text-ink-soft mb-3">CHECK-IN</p>
      <h1 class="font-serif text-3xl font-semibold text-ink mb-1 leading-snug">
        {{ selected?.habit_name }}
      </h1>
      <p v-if="selected?.scheduled_time" class="text-sm text-ink-soft mb-8">
        Hora programada: {{ selected.scheduled_time }}
      </p>
      <div v-else class="mb-8" />

      <!-- Note textarea -->
      <label class="block mb-6">
        <span class="text-xs font-semibold text-ink-soft uppercase tracking-eyebrow mb-2 block">
          Nota (opcional)
        </span>
        <textarea
          v-model="note"
          rows="3"
          placeholder="¿Algo que quieras compartir con el squad?"
          class="w-full rounded-[14px] border border-hairline bg-surface px-4 py-3
                 text-sm text-ink placeholder-ink-faint resize-none
                 focus:outline-none focus:border-sage-deep transition-colors"
        />
      </label>

      <!-- Error -->
      <p v-if="errMsg" class="text-sm text-coral mb-4 font-medium">{{ errMsg }}</p>

      <!-- Actions -->
      <div class="mt-auto space-y-3">
        <button
          :disabled="submitting"
          class="w-full rounded-pill bg-sage-deep text-paper py-4 font-bold text-sm
                 disabled:opacity-50 transition-opacity active:opacity-80"
          @click="submit"
        >
          <span v-if="submitting" class="flex items-center justify-center gap-2">
            <span class="w-4 h-4 rounded-full border-2 border-paper border-t-transparent animate-spin" />
            Registrando...
          </span>
          <span v-else>Registrar check-in ✓</span>
        </button>

        <button
          v-if="myPending.length > 1"
          class="w-full rounded-pill border border-hairline text-ink-soft py-3.5
                 text-sm font-medium"
          @click="step = 'select'"
        >
          ← Cambiar hábito
        </button>
      </div>
    </div>

    <!-- ── Success ─────────────────────────────────────────────────────────── -->
    <div v-else-if="step === 'success'"
      class="flex-1 flex flex-col items-center justify-center px-8 text-center">

      <!-- Celebration ring -->
      <div class="relative mb-8">
        <div class="w-28 h-28 rounded-full bg-sage/20 flex items-center justify-center">
          <div class="w-20 h-20 rounded-full bg-sage-deep flex items-center justify-center">
            <svg class="w-10 h-10 text-paper" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3"
                d="M5 13l4 4L19 7"/>
            </svg>
          </div>
        </div>
      </div>

      <p class="text-eyebrow tracking-eyebrow text-sage-deep mb-2">¡LO LOGRASTE!</p>
      <h1 class="font-serif text-3xl font-semibold text-ink mb-1">
        {{ selected?.habit_name }}
      </h1>
      <p class="text-sm text-ink-soft mb-8">Check-in registrado hoy</p>

      <!-- Streak highlight -->
      <div class="rounded-card bg-paper shadow-card px-8 py-5 mb-10 w-full max-w-xs">
        <p class="text-eyebrow tracking-eyebrow text-ink-soft mb-1">TU RACHA</p>
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
