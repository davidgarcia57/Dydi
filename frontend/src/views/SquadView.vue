<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { useProposalsStore } from '@/stores/proposals'
import { useGroupSocket } from '@/composables/useGroupSocket'
import { showToast } from '@/composables/useToast'
import { mondayIndex } from '@/composables/useWeekStatus'
import PageContainer from '@/components/ui/PageContainer.vue'
import BaseAvatar from '@/components/ui/BaseAvatar.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import SparkleGlyph from '@/components/ui/SparkleGlyph.vue'

const auth = useAuthStore()
const group = useGroupStore()
const habits = useHabitsStore()
const proposals = useProposalsStore()
const loaded = ref(false)
const loadError = ref(false)
const confirmKick = ref(null) // user_id pending kick confirmation
const kicking = ref(null) // user_id whose kick proposal is in flight

// Group checkins by member — includes me, so my own week shows here too.
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

// ── 7-day strip (Monday-first) — same data source as TodayView ────────────────
const DAY_LABELS = ['L', 'M', 'M', 'J', 'V', 'S', 'D']

// Monday-first index (0..6) → the YYYY-MM-DD date for that day of the current week.
function dateForIdx(i, todayIdx) {
  const d = new Date()
  d.setDate(d.getDate() - (todayIdx - i))
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function dayStrip(checkin) {
  const dow = new Date().getDay()
  // Convert Sun=0…Sat=6 → Mon=0…Sun=6
  const todayIdx = dow === 0 ? 6 : dow - 1
  const key = checkin ? `${checkin.user_id}:${checkin.habit_id}` : ''
  const dates = checkin ? habits.weekHistory[key] : null

  return DAY_LABELS.map((label, i) => {
    if (i > todayIdx) return { label, status: 'future' }
    if (i === todayIdx) return { label, status: checkin?.status ?? 'pending', note: checkin?.note }
    // Past day: real history — a check-in that date means done, otherwise missed.
    const date = dateForIdx(i, todayIdx)
    const done = dates ? dates.has(date) : false
    return { label, status: done ? 'done' : 'missed', note: habits.weekNotes[`${key}:${date}`] }
  })
}

const STATUS_STYLE = {
  done: { strip: 'bg-sage', icon: '✓', iconColor: 'text-sage-deep' },
  pending: { strip: 'bg-amber', icon: '', iconColor: '' },
  missed: { strip: 'bg-coral', icon: '✗', iconColor: 'text-coral-deep' },
  future: { strip: 'border border-dashed border-hairline bg-transparent', icon: '', iconColor: '' },
}

function memberStatus(row) {
  if (row.habits.every((h) => h.status === 'done')) return 'done'
  if (row.habits.some((h) => h.status === 'pending')) return 'pending'
  return 'missed'
}

// ── Matriz semanal L–V: filas = miembros, columnas = días ────────────────────
// El patrón desktop del grupo completo de un vistazo. Solo L–V porque eso es
// lo que juzga la ruleta; sábado y domingo son su fin de semana.
const MATRIX_LABELS = ['L', 'M', 'M', 'J', 'V']

const MATRIX_CELL = {
  done: 'bg-sage text-sage-deep',
  partial: 'bg-sage/50 text-sage-deep',
  pending: 'bg-amber text-amber-deep',
  missed: 'bg-coral text-coral-deep',
  future: 'border border-dashed border-hairline text-transparent',
}

const MATRIX_ICON = { done: '✓', partial: '~', missed: '✗' }

function weekMatrixRow(row) {
  const todayIdx = mondayIndex()
  return MATRIX_LABELS.map((label, i) => {
    if (i > todayIdx) return { label, status: 'future' }
    if (i === todayIdx) return { label, status: memberStatus(row) }
    const date = dateForIdx(i, todayIdx)
    let done = 0
    for (const h of row.habits) {
      if (habits.weekHistory[`${row.user_id}:${h.habit_id}`]?.has(date)) done++
    }
    if (done === row.habits.length) return { label, status: 'done' }
    return { label, status: done > 0 ? 'partial' : 'missed' }
  })
}

// Semana perfecta: todos los días L–V transcurridos (incluido hoy) en verde.
function isPerfectWeek(row) {
  if (!row.habits.length) return false
  const cells = weekMatrixRow(row)
  return (
    cells.some((c) => c.status === 'done') &&
    cells.every((c) => c.status === 'done' || c.status === 'future')
  )
}

async function proposeKick(row) {
  if (kicking.value) return
  kicking.value = row.user_id
  try {
    await proposals.propose(group.group.id, 'kick_member', { targetUserID: row.user_id })
    confirmKick.value = null
    showToast('Propuesta enviada. El squad la vota en Votar.')
  } catch (e) {
    showToast(e?.error ?? e?.message ?? 'No se pudo proponer la expulsión.')
  } finally {
    kicking.value = null
  }
}

let socketDisconnect = null

async function load() {
  loadError.value = false
  loaded.value = false
  try {
    await group.autoLoad()
    if (group.group?.id) {
      await habits.loadToday(group.group.id)
      await habits.loadWeekHistory(group.group.id)
      const ids = [...new Set(habits.todayCheckins.map((c) => c.user_id))]
      await Promise.all(ids.map((id) => habits.loadStreaks(id)))
      const { disconnect } = useGroupSocket(group.group.id)
      socketDisconnect = disconnect
    }
    loaded.value = true
  } catch (_) {
    loadError.value = true
  }
}

onMounted(load)

onUnmounted(() => socketDisconnect?.())
</script>

<template>
  <PageContainer>
    <header class="mb-6">
      <div class="flex items-center justify-between">
        <h1 class="font-serif text-2xl font-semibold text-ink">Squad</h1>
        <span class="text-eyebrow">{{ group.group?.name ?? '' }}</span>
      </div>
      <p class="text-xs text-ink-faint mt-0.5">La semana del equipo · presencia en vivo</p>
    </header>

    <div
      v-if="loadError"
      class="rounded-card bg-coral-soft/40 border border-coral/40 py-10 text-center"
    >
      <p class="text-sm font-medium text-coral-deep mb-3">No pudimos cargar el squad.</p>
      <button
        class="rounded-pill bg-coral text-paper px-4 py-2 text-sm font-bold active:opacity-80 transition-opacity"
        @click="load"
      >
        Reintentar
      </button>
    </div>

    <div
      v-else-if="!squadRows.length"
      class="rounded-card bg-surface border border-hairline py-10 text-center text-sm text-ink-soft"
    >
      <span v-if="!loaded">Cargando el squad…</span>
      <span v-else>Ningún miembro tiene hábitos asignados todavía.</span>
    </div>

    <template v-else>
      <!-- ── La semana del squad: matriz miembros × L–V ───────────────────── -->
      <section class="rounded-card bg-paper shadow-card p-4 sm:p-5 mb-5 overflow-x-auto">
        <div class="flex items-center justify-between gap-3 mb-4">
          <h2 class="text-eyebrow">LA SEMANA DEL SQUAD</h2>
          <span class="text-[10px] text-ink-faint">L–V cuentan para la ruleta</span>
        </div>

        <div class="min-w-[19rem]">
          <!-- Encabezado de días -->
          <div class="flex items-center gap-3 px-2 mb-1">
            <div class="w-6 flex-shrink-0" />
            <span class="flex-1" />
            <div class="flex gap-1">
              <span
                v-for="(l, i) in MATRIX_LABELS"
                :key="i"
                class="w-6 text-center text-[9px] text-ink-faint font-medium"
              >
                {{ l }}
              </span>
            </div>
          </div>

          <div class="space-y-1">
            <div
              v-for="row in squadRows"
              :key="row.user_id"
              class="flex items-center gap-3 rounded-xl px-2 py-1.5"
              :class="{ 'bg-gradient-to-r from-amber-soft/70 to-transparent': isPerfectWeek(row) }"
            >
              <BaseAvatar :name="row.display_name" size="sm" />
              <span class="flex-1 min-w-0 text-xs font-semibold text-ink truncate">
                {{ row.user_id === auth.user?.id ? 'Tú' : row.display_name }}
                <span
                  v-if="isPerfectWeek(row)"
                  class="text-[10px] font-bold text-amber-deep inline-flex items-center gap-0.5"
                >
                  <SparkleGlyph :size="11" />
                  semana perfecta
                </span>
              </span>
              <div class="flex gap-1">
                <div
                  v-for="(cell, i) in weekMatrixRow(row)"
                  :key="i"
                  class="w-6 h-6 rounded-md flex items-center justify-center text-[10px] font-bold"
                  :class="MATRIX_CELL[cell.status]"
                >
                  {{ MATRIX_ICON[cell.status] ?? '' }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-3">
        <div
          v-for="row in squadRows"
          :key="row.user_id"
          class="rounded-card shadow-flat bg-paper p-4"
          :class="{ 'ring-2 ring-sage/40': group.onlineMembers.has(row.user_id) }"
        >
          <!-- Member header -->
          <div class="flex items-center gap-3 mb-3">
            <div class="relative">
              <BaseAvatar :name="row.display_name" size="md" />
              <span
                v-if="group.onlineMembers.has(row.user_id)"
                class="absolute bottom-0 right-0 w-3 h-3 rounded-full bg-sage-deep border-2 border-paper"
              />
            </div>

            <div class="flex-1 min-w-0">
              <div class="flex items-baseline gap-2">
                <span class="font-semibold text-sm text-ink truncate">{{ row.display_name }}</span>
                <span class="text-xs text-terracotta font-medium flex-shrink-0">
                  ★ {{ habits.streaks[row.user_id] ?? 0 }}
                </span>
              </div>
              <p v-if="row.user_id === auth.user?.id" class="text-xs text-ink-soft mt-0.5">Tú</p>
            </div>

            <StatusBadge :status="memberStatus(row)" class="flex-shrink-0" />
          </div>

          <!-- One block per assigned habit: name + status + 7-day strip -->
          <div class="space-y-3">
            <div v-for="h in row.habits" :key="h.habit_id">
              <div class="flex justify-between items-center gap-2 mb-1">
                <p class="text-xs text-ink-soft truncate">
                  {{ h.habit_name }}
                  <span
                    v-if="habits.streakByHabit[`${row.user_id}:${h.habit_id}`]"
                    class="text-terracotta font-semibold"
                  >
                    ★ {{ habits.streakByHabit[`${row.user_id}:${h.habit_id}`] }}
                  </span>
                </p>
                <StatusBadge :status="h.status" class="flex-shrink-0" />
              </div>

              <div class="flex gap-1">
                <div
                  v-for="(day, i) in dayStrip(h)"
                  :key="i"
                  class="flex flex-col items-center gap-0.5"
                  :title="day.note ? `“${day.note}”` : undefined"
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

          <!-- Propose kicking this member (never yourself) -->
          <div v-if="row.user_id !== auth.user?.id" class="mt-3 pt-3 border-t border-hairline">
            <button
              v-if="confirmKick !== row.user_id"
              class="text-xs font-semibold text-ink-faint hover:text-coral-deep transition-colors"
              @click="confirmKick = row.user_id"
            >
              Proponer expulsión
            </button>
            <div v-else class="space-y-2">
              <p class="text-xs text-ink-soft">
                ¿Proponer expulsar a
                <span class="font-semibold text-ink">{{ row.display_name }}</span
                >? El squad lo vota por mayoría.
              </p>
              <div class="flex gap-2">
                <button
                  :disabled="kicking === row.user_id"
                  class="flex-1 rounded-pill bg-coral text-paper py-2 text-xs font-bold disabled:opacity-40 active:opacity-80 transition-opacity"
                  @click="proposeKick(row)"
                >
                  {{ kicking === row.user_id ? 'Enviando…' : 'Sí, proponer' }}
                </button>
                <button
                  class="flex-1 rounded-pill border border-hairline text-ink-soft py-2 text-xs font-semibold active:opacity-70 transition-opacity"
                  @click="confirmKick = null"
                >
                  Cancelar
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </PageContainer>
</template>
