<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { usePenaltiesStore } from '@/stores/penalties'

const auth      = useAuthStore()
const group     = useGroupStore()
const penalties = usePenaltiesStore()

// ── State ─────────────────────────────────────────────────────────────────────
const view       = ref('list')  // 'list' | 'entry'
const loading    = ref(false)
const spinning   = ref(false)
const error      = ref(null)
const spinResult = ref(null)
const showForm   = ref(false)
const sugText    = ref('')
const sugEmoji   = ref('')

// ── Entry computed ─────────────────────────────────────────────────────────────
const entry = computed(() => penalties.activeEntry)

const deadlinePassed = computed(() =>
  entry.value ? new Date() > new Date(entry.value.suggestion_deadline) : false
)
const isDebtor    = computed(() => entry.value?.debtor_id === auth.user?.id)
const canSpin     = computed(() => isDebtor.value && deadlinePassed.value && !entry.value?.spun_at)
const hasSuggested = computed(() =>
  penalties.suggestions.some(s => s.suggester_id === auth.user?.id)
)
const canSuggest  = computed(() => !deadlinePassed.value && !hasSuggested.value)

const deadlineLabel = computed(() => {
  if (!entry.value) return ''
  const diff = new Date(entry.value.suggestion_deadline) - new Date()
  if (diff <= 0) return 'Ventana cerrada'
  const hrs  = Math.floor(diff / 3_600_000)
  const mins = Math.floor((diff % 3_600_000) / 60_000)
  if (hrs >= 24) return `${Math.floor(hrs / 24)}d ${hrs % 24}h`
  if (hrs > 0)   return `${hrs}h ${mins}min`
  return `${mins}min`
})

const debtorName = computed(() => {
  if (!entry.value) return ''
  return (
    group.members.find(m => m.user_id === entry.value.debtor_id)?.display_name
    ?? penalties.eligible.find(m => m.user_id === entry.value.debtor_id)?.display_name
    ?? 'miembro'
  )
})

const spinDebts = computed(() =>
  spinResult.value
    ? (Array.isArray(spinResult.value) ? spinResult.value : [spinResult.value])
    : []
)

// ── Helpers ───────────────────────────────────────────────────────────────────
const COLORS = ['bg-sage-deep', 'bg-terracotta', 'bg-sage', 'bg-amber', 'bg-coral']
const initials  = (n = '') => n.trim().split(/\s+/).map(w => w[0]).join('').slice(0, 2).toUpperCase()
const avatarBg  = (n = '') => COLORS[(n?.charCodeAt(0) ?? 0) % COLORS.length]
const memberName = id => group.members.find(m => m.user_id === id)?.display_name ?? '?'
const shortDate  = iso  => new Date(iso).toLocaleDateString('es-MX', { month: 'short', day: 'numeric' })

// ── Actions ───────────────────────────────────────────────────────────────────
async function openRoulette(member) {
  loading.value = true
  error.value   = null
  try {
    await penalties.openRoulette(group.group.id, member.user_id)
    await penalties.loadSuggestions(penalties.activeEntry.id)
    view.value = 'entry'
  } catch (e) {
    error.value = e?.error ?? 'No se pudo abrir la ruleta'
  } finally {
    loading.value = false
  }
}

async function submitSuggestion() {
  const text = sugText.value.trim()
  if (!text) return
  error.value = null
  try {
    await penalties.submitSuggestion(entry.value.id, text, sugEmoji.value.trim() || null)
    sugText.value  = ''
    sugEmoji.value = ''
    showForm.value = false
  } catch (e) {
    error.value = e?.error ?? 'No se pudo enviar'
  }
}

async function doSpin() {
  spinning.value = true
  error.value    = null
  try {
    spinResult.value = await penalties.spin(entry.value.id)
  } catch (e) {
    error.value = e?.error ?? 'Error al girar'
  } finally {
    spinning.value = false
  }
}

function backToList() {
  view.value       = 'list'
  spinResult.value = null
  showForm.value   = false
  error.value      = null
  penalties.clearEntry()
}

onMounted(async () => {
  await group.autoLoad()
  if (group.group?.id) {
    await Promise.all([
      penalties.loadEligible(group.group.id),
      penalties.loadDebts(group.group.id),
    ])
  }
})
</script>

<template>
  <div class="max-w-md mx-auto px-4 pt-4 pb-6">

    <!-- ═══════════════════ LIST VIEW ═══════════════════════════════════════ -->
    <template v-if="view === 'list'">

      <header class="flex items-center justify-between mb-6">
        <h1 class="font-serif text-2xl font-semibold text-ink">Ruleta</h1>
        <span class="text-eyebrow">{{ group.group?.name ?? '' }}</span>
      </header>

      <div v-if="error"
        class="mb-4 rounded-card bg-coral/20 text-coral px-4 py-3 text-sm font-medium">
        {{ error }}
      </div>

      <!-- En deuda esta semana -->
      <section class="mb-7">
        <h2 class="text-eyebrow mb-3">EN DEUDA ESTA SEMANA</h2>

        <div v-if="!penalties.eligible.length"
          class="rounded-card border border-sage/30 bg-sage/10 px-4 py-6 text-center">
          <p class="font-serif text-3xl mb-1">🎉</p>
          <p class="text-sm font-semibold text-sage-deep">Squad limpio esta semana</p>
          <p class="text-xs text-ink-soft mt-1">Nadie falló ningún hábito.</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="m in penalties.eligible"
            :key="m.user_id"
            class="rounded-card shadow-flat bg-surface p-4 flex items-center gap-3"
          >
            <div
              class="w-10 h-10 rounded-full flex-shrink-0 flex items-center
                     justify-center text-paper text-sm font-bold"
              :class="avatarBg(m.display_name)"
            >
              {{ initials(m.display_name) }}
            </div>
            <div class="flex-1 min-w-0">
              <p class="font-semibold text-sm text-ink truncate">{{ m.display_name }}</p>
              <p class="text-xs text-ink-soft mt-0.5">Falló esta semana</p>
            </div>
            <button
              class="rounded-pill bg-terracotta text-paper px-4 py-2 text-xs font-bold
                     active:opacity-80 transition-opacity flex-shrink-0"
              :disabled="loading"
              @click="openRoulette(m)"
            >
              {{ loading ? '...' : 'Abrir →' }}
            </button>
          </div>
        </div>
      </section>

      <!-- Deudas activas -->
      <section>
        <h2 class="text-eyebrow mb-3">DEUDAS ACTIVAS</h2>

        <div v-if="!penalties.debts.length"
          class="rounded-card bg-surface border border-hairline px-4 py-5 text-center text-sm text-ink-soft">
          Sin deudas activas en el grupo.
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="debt in penalties.debts"
            :key="debt.id"
            class="rounded-card shadow-flat bg-surface p-4"
          >
            <div class="flex items-center justify-between mb-2">
              <div class="flex items-center gap-2">
                <div
                  class="w-7 h-7 rounded-full flex items-center justify-center
                         text-paper text-[10px] font-bold"
                  :class="avatarBg(memberName(debt.debtor_id))"
                >
                  {{ initials(memberName(debt.debtor_id)) }}
                </div>
                <span class="text-xs font-semibold text-ink">{{ memberName(debt.debtor_id) }}</span>
                <span v-if="debt.is_collective"
                  class="rounded-pill bg-coral/20 text-coral text-[10px] font-bold px-2 py-0.5">
                  colectiva
                </span>
              </div>
              <span class="text-[10px] text-ink-faint">exp. {{ shortDate(debt.expires_at) }}</span>
            </div>
            <p class="text-sm font-semibold text-ink">
              {{ debt.punishment_emoji ?? '' }} {{ debt.punishment_text }}
            </p>
          </div>
        </div>
      </section>
    </template>

    <!-- ═══════════════════ ENTRY VIEW ════════════════════════════════════════ -->
    <template v-else-if="view === 'entry' && entry">

      <header class="flex items-center gap-3 mb-6">
        <button
          class="w-9 h-9 rounded-full flex items-center justify-center
                 bg-surface border border-hairline"
          @click="backToList"
        >
          <svg class="w-4 h-4 text-ink" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M15 19l-7-7 7-7"/>
          </svg>
        </button>
        <div>
          <h1 class="font-serif text-xl font-semibold text-ink leading-tight">
            Ruleta de {{ debtorName }}
          </h1>
          <p class="text-xs text-ink-soft">Semana actual</p>
        </div>
      </header>

      <div v-if="error"
        class="mb-4 rounded-card bg-coral/20 text-coral px-4 py-3 text-sm font-medium">
        {{ error }}
      </div>

      <!-- RESULTADO DEL SPIN -->
      <template v-if="spinResult">
        <div class="rounded-card shadow-card bg-paper p-8 mb-5 text-center">
          <p class="text-6xl mb-4">{{ spinDebts[0]?.punishment_emoji ?? '🎲' }}</p>
          <p class="text-eyebrow text-terracotta mb-2">
            {{ spinDebts[0]?.is_collective ? 'DEUDA COLECTIVA' : 'PENITENCIA ASIGNADA' }}
          </p>
          <h2 class="font-serif text-2xl font-semibold text-ink leading-snug mb-3">
            {{ spinDebts[0]?.punishment_text }}
          </h2>
          <p v-if="spinDebts[0]?.is_collective" class="text-sm text-ink-soft mb-3">
            Nadie propuso penitencia — el squad completo paga.
          </p>
          <p class="text-xs text-ink-faint">
            Expira el {{ shortDate(spinDebts[0]?.expires_at) }}
          </p>
        </div>
        <button
          class="w-full rounded-pill bg-sage-deep text-paper py-3.5 font-bold text-sm"
          @click="backToList"
        >
          Volver a la ruleta
        </button>
      </template>

      <!-- PRE-SPIN -->
      <template v-else>

        <!-- Deadline strip -->
        <div class="rounded-card bg-surface border border-hairline p-4 mb-4
                    flex justify-between items-center">
          <div>
            <p class="text-eyebrow">VENTANA DE SUGERENCIAS</p>
            <p class="font-semibold text-sm text-ink mt-0.5">
              {{ deadlinePassed ? 'Cerrada' : 'Abierta' }}
            </p>
          </div>
          <p class="font-serif text-lg font-semibold"
            :class="deadlinePassed ? 'text-coral' : 'text-sage-deep'">
            {{ deadlineLabel }}
          </p>
        </div>

        <!-- Suggestions -->
        <section class="mb-4">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-eyebrow">PENITENCIAS PROPUESTAS</h2>
            <span class="text-xs text-ink-soft">{{ penalties.suggestions.length }}</span>
          </div>

          <div v-if="!penalties.suggestions.length"
            class="rounded-card bg-surface border border-hairline px-4 py-5
                   text-center text-sm text-ink-soft mb-3">
            Nadie ha propuesto una penitencia aún.
          </div>

          <div v-else class="space-y-2 mb-3">
            <div
              v-for="s in penalties.suggestions"
              :key="s.id"
              class="rounded-card bg-surface border border-hairline p-3 flex items-center gap-3"
            >
              <div
                class="w-7 h-7 rounded-full flex-shrink-0 flex items-center
                       justify-center text-paper text-[10px] font-bold"
                :class="avatarBg(memberName(s.suggester_id))"
              >
                {{ initials(memberName(s.suggester_id)) }}
              </div>
              <p class="text-sm text-ink">
                <span v-if="s.emoji" class="mr-1">{{ s.emoji }}</span>{{ s.text }}
              </p>
            </div>
          </div>

          <!-- Suggestion form -->
          <template v-if="canSuggest">
            <button
              v-if="!showForm"
              class="w-full rounded-pill border border-sage-deep text-sage-deep py-3
                     font-bold text-sm active:opacity-80 transition-opacity"
              @click="showForm = true"
            >
              + Proponer penitencia
            </button>
            <div v-else class="rounded-card bg-surface border border-hairline p-4">
              <p class="text-eyebrow mb-3">TU PROPUESTA</p>
              <div class="flex gap-2 mb-3">
                <input
                  v-model="sugEmoji"
                  type="text"
                  placeholder="😈"
                  maxlength="2"
                  class="w-14 rounded-xl border border-hairline bg-paper px-3 py-2.5
                         text-center text-lg focus:outline-none focus:border-sage-deep"
                />
                <input
                  v-model="sugText"
                  type="text"
                  placeholder="Ej: 30 sentadillas en público"
                  class="flex-1 rounded-xl border border-hairline bg-paper px-3 py-2.5
                         text-sm focus:outline-none focus:border-sage-deep"
                />
              </div>
              <div class="flex gap-2">
                <button
                  class="flex-1 rounded-pill bg-sage-deep text-paper py-2.5 font-bold text-sm"
                  @click="submitSuggestion"
                >
                  Enviar
                </button>
                <button
                  class="rounded-pill border border-hairline text-ink-soft px-4 py-2.5 text-sm font-bold"
                  @click="showForm = false"
                >
                  Cancelar
                </button>
              </div>
            </div>
          </template>

          <div v-else-if="!deadlinePassed && hasSuggested"
            class="rounded-pill bg-sage/20 text-sage-deep px-4 py-3 text-sm font-semibold text-center">
            ✓ Ya propusiste tu penitencia
          </div>
        </section>

        <!-- GIRAR -->
        <button
          v-if="canSpin"
          class="w-full rounded-pill bg-terracotta text-paper py-4 font-bold text-base
                 active:opacity-80 transition-all"
          :class="{ 'opacity-60': spinning }"
          :disabled="spinning"
          @click="doSpin"
        >
          {{ spinning ? 'Girando...' : '¡GIRAR LA RULETA!' }}
        </button>

        <div
          v-else-if="deadlinePassed && !isDebtor && !entry.spun_at"
          class="rounded-card bg-amber/10 border border-amber/30 p-4 text-center"
        >
          <p class="text-sm font-semibold text-amber">
            Esperando que {{ debtorName }} gire
          </p>
          <p class="text-xs text-ink-soft mt-1">La ventana de sugerencias ya cerró.</p>
        </div>

        <div v-else-if="entry.spun_at"
          class="rounded-card bg-surface border border-hairline p-4 text-center text-sm text-ink-soft">
          Esta ruleta ya fue girada.
        </div>
      </template>

    </template>
  </div>
</template>
