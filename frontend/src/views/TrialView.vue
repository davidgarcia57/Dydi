<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { usePenaltiesStore } from '@/stores/penalties'
import { showToast } from '@/composables/useToast'
import PageContainer from '@/components/ui/PageContainer.vue'
import BaseAvatar from '@/components/ui/BaseAvatar.vue'
import RouletteWheel from '@/components/ui/RouletteWheel.vue'
import DoveHero from '@/components/ui/DoveHero.vue'
import { Dices, PartyPopper } from 'lucide-vue-next'

// Espeja spinGraceHours del backend: pasado el deadline + gracia, cualquier
// miembro puede girar por el deudor para que la ruleta nunca muera sin girar.
const SPIN_GRACE_MS = 24 * 3_600_000

const auth = useAuthStore()
const group = useGroupStore()
const penalties = usePenaltiesStore()

// ── State ─────────────────────────────────────────────────────────────────────
const view = ref('list')
const pageLoading = ref(true)
const loading = ref(false)
const spinning = ref(false)
const error = ref(null)
const spinResult = ref(null)
const showForm = ref(false)
const sugText = ref('')
const sugEmoji = ref('')
const confirmComplete = ref(null) // debt.id pendiente de confirmación
const completing = ref(null)
const confirmForgive = ref(null) // debt.id pendiente de confirmación de perdón
const forgiving = ref(null)
const showHistory = ref(false)
const historyLoaded = ref(false)

const DEBT_STATUS_BADGE = {
  completed: { label: 'CUMPLIDA', class: 'bg-sage-soft text-sage-deep' },
  forgiven: { label: 'PERDONADA', class: 'bg-amber-soft text-amber-deep' },
  expired: { label: 'EXPIRÓ', class: 'bg-cream-2 text-ink-faint' },
}

// Carga perezosa del historial de deudas al abrirlo por primera vez.
async function toggleHistory() {
  showHistory.value = !showHistory.value
  if (!showHistory.value || historyLoaded.value || !group.group?.id) return
  try {
    await penalties.loadResolvedDebts(group.group.id)
    historyLoaded.value = true
  } catch (_) {
    // el empty-state cubre el fallo; reintenta al volver a abrir
  }
}

// ── Wheel animation ───────────────────────────────────────────────────────────
const spinDeg = ref(0)

// ── Entry computed ─────────────────────────────────────────────────────────────
const entry = computed(() => penalties.activeEntry)

const deadlinePassed = computed(() =>
  entry.value ? new Date() > new Date(entry.value.suggestion_deadline) : false
)
const isDebtor = computed(() => entry.value?.debtor_id === auth.user?.id)
const graceOver = computed(() =>
  entry.value
    ? Date.now() > new Date(entry.value.suggestion_deadline).getTime() + SPIN_GRACE_MS
    : false
)
const canSpin = computed(
  () => deadlinePassed.value && !entry.value?.spun_at && (isDebtor.value || graceOver.value)
)
const hasSuggested = computed(() =>
  penalties.suggestions.some((s) => s.suggester_id === auth.user?.id)
)
// El deudor nunca escribe su propia penitencia: la propone el resto del squad.
const canSuggest = computed(() => !deadlinePassed.value && !hasSuggested.value && !isDebtor.value)

const deadlineLabel = computed(() => {
  if (!entry.value) return ''
  const diff = new Date(entry.value.suggestion_deadline) - new Date()
  if (diff <= 0) return 'Ventana cerrada'
  const hrs = Math.floor(diff / 3_600_000)
  const mins = Math.floor((diff % 3_600_000) / 60_000)
  if (hrs >= 24) return `${Math.floor(hrs / 24)}d ${hrs % 24}h`
  if (hrs > 0) return `${hrs}h ${mins}min`
  return `${mins}min`
})

const debtorName = computed(() => {
  if (!entry.value) return ''
  return (
    group.members.find((m) => m.user_id === entry.value.debtor_id)?.display_name ??
    penalties.eligible.find((m) => m.user_id === entry.value.debtor_id)?.display_name ??
    entry.value.debtor_name ??
    'miembro'
  )
})

// Miembros en el bote que aún no tienen ruleta abierta (los que ya la tienen
// aparecen en la sección "Ruletas abiertas").
const eligibleWithoutEntry = computed(() =>
  penalties.eligible.filter((m) => !penalties.openEntries.some((e) => e.debtor_id === m.user_id))
)

// Nada en juego: ni ruletas abiertas, ni bote, ni deudas → la ruleta duerme.
const rouletteAsleep = computed(
  () =>
    !penalties.openEntries.length && !eligibleWithoutEntry.value.length && !penalties.debts.length
)

const isWeekend = computed(() => {
  const dow = new Date().getDay()
  return dow === 6 || dow === 0
})

// Deudas a punto de caducar (<48 h) se marcan en coral.
const expiresSoon = (debt) => new Date(debt.expires_at) - Date.now() < 2 * 86_400_000

function entryCountdown(e) {
  const diff = new Date(e.suggestion_deadline) - new Date()
  if (diff <= 0) return '¡Lista para girar!'
  const hrs = Math.floor(diff / 3_600_000)
  const mins = Math.floor((diff % 3_600_000) / 60_000)
  if (hrs >= 24) return `Sugerencias por ${Math.floor(hrs / 24)}d ${hrs % 24}h`
  if (hrs > 0) return `Sugerencias por ${hrs}h ${mins}min`
  return `Sugerencias por ${mins}min`
}

function entryDebtorName(e) {
  return e.debtor_name ?? memberName(e.debtor_id)
}

const spinDebts = computed(() =>
  spinResult.value ? (Array.isArray(spinResult.value) ? spinResult.value : [spinResult.value]) : []
)

// ── Ruleta ────────────────────────────────────────────────────────────────────
// Los colores viven aquí (se comparten entre la rueda y los chips de
// sugerencias); la geometría y el aspecto viven en RouletteWheel.
const WHEEL_COLORS = [
  '#C26F4D',
  '#A8C39A',
  '#5C7650',
  '#E9C281',
  '#EDA48F',
  '#BC5C42',
  '#7CA39D',
  '#A57B33',
]

// Con menos de 2 sugerencias la rueda se rellena a 8 segmentos vacíos.
const wheelCount = computed(() =>
  penalties.suggestions.length >= 2 ? penalties.suggestions.length : 8
)

// ── Helpers ───────────────────────────────────────────────────────────────────
const memberName = (id) => group.members.find((m) => m.user_id === id)?.display_name ?? '?'
const shortDate = (iso) =>
  new Date(iso).toLocaleDateString('es-MX', { month: 'short', day: 'numeric' })

// ── Actions ───────────────────────────────────────────────────────────────────
async function openRoulette(member) {
  loading.value = true
  error.value = null
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

// Entra a una ruleta ya abierta (sin POST: re-abrir exige elegibilidad vigente).
async function enterEntry(e) {
  loading.value = true
  error.value = null
  try {
    penalties.enterEntry(e)
    await penalties.loadSuggestions(e.id)
    view.value = 'entry'
  } catch (err) {
    error.value = err?.error ?? 'No se pudo abrir la ruleta'
  } finally {
    loading.value = false
  }
}

async function completeDebt(debt) {
  if (confirmComplete.value !== debt.id) {
    confirmComplete.value = debt.id
    return
  }
  completing.value = debt.id
  try {
    await penalties.completeDebt(debt.id)
    showToast('Penitencia cumplida')
  } catch (e) {
    error.value = e?.error ?? 'No se pudo marcar la deuda'
  } finally {
    completing.value = null
    confirmComplete.value = null
  }
}

// Perdonar es del squad, nunca del deudor (el backend también lo exige).
async function forgiveDebt(debt) {
  if (confirmForgive.value !== debt.id) {
    confirmForgive.value = debt.id
    return
  }
  forgiving.value = debt.id
  try {
    await penalties.forgiveDebt(debt.id)
    showToast('Deuda perdonada')
  } catch (e) {
    error.value = e?.error ?? 'No se pudo perdonar la deuda'
  } finally {
    forgiving.value = null
    confirmForgive.value = null
  }
}

async function submitSuggestion() {
  const text = sugText.value.trim()
  if (!text) return
  error.value = null
  try {
    await penalties.submitSuggestion(entry.value.id, text, sugEmoji.value.trim() || null)
    sugText.value = ''
    sugEmoji.value = ''
    showForm.value = false
  } catch (e) {
    error.value = e?.error ?? 'No se pudo enviar'
  }
}

async function doSpin() {
  if (spinning.value) return
  spinning.value = true
  error.value = null

  let result = null
  try {
    // Pedimos el ganador real primero
    result = await penalties.spin(entry.value.id)
  } catch (e) {
    error.value = e?.error ?? 'Error al girar'
    spinning.value = false
    return
  }

  // 1. Encontrar el segmento ganador
  const items =
    penalties.suggestions.length >= 2 ? penalties.suggestions : Array.from({ length: 8 })
  let winnerIndex = items.findIndex((s) => s.id === result.winning_suggestion_id)
  if (winnerIndex === -1) winnerIndex = Math.floor(Math.random() * items.length)

  // 2. Calcular ángulo exacto para que el segmento ganador aterrice arriba (0 grados)
  const anglePer = 360 / items.length
  const centerDeg = (winnerIndex + 0.5) * anglePer
  // Desplazamiento aleatorio dentro del segmento para que no se vea robótico
  const offset = (Math.random() - 0.5) * (anglePer * 0.8)
  const targetLanding = centerDeg + offset

  // 3. Aplicar rotaciones extra y ajustar grado final
  const currentRotations = Math.floor(spinDeg.value / 360)
  const extraRotations = 5 // 5 vueltas completas de suspenso
  const newDeg = (currentRotations + extraRotations) * 360 + (360 - targetLanding)

  spinDeg.value = newDeg

  // Esperar a que la transición CSS (4.2s) termine, y retener el reveal un
  // beat más: la ruleta paró… ¿y? Ese medio segundo ES el suspenso.
  setTimeout(() => {
    setTimeout(() => {
      spinResult.value = result
      spinning.value = false

      revealResult()
    }, 650)
  }, 4200)
}

function revealResult() {
  // Celebración via CDN
  const triggerConfetti = () => {
    if (window.confetti) {
      window.confetti({
        particleCount: 120,
        spread: 70,
        origin: { y: 0.6 },
        colors: ['#C26F4D', '#A8C39A', '#5C7650', '#E9C281'],
      })
    }
  }

  if (!window.confetti) {
    const script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/npm/canvas-confetti@1.9.2/dist/confetti.browser.min.js'
    script.onload = triggerConfetti
    document.head.appendChild(script)
  } else {
    triggerConfetti()
  }
}

function backToList() {
  view.value = 'list'
  spinResult.value = null
  showForm.value = false
  error.value = null
  spinDeg.value = 0
  penalties.clearEntry()
}

onMounted(async () => {
  try {
    await group.autoLoad()
    if (group.group?.id) {
      await Promise.all([
        penalties.loadEligible(group.group.id),
        penalties.loadDebts(group.group.id),
        penalties.loadOpenEntries(group.group.id),
      ])
    }
  } catch (e) {
    error.value = e?.error ?? 'No se pudo cargar la ruleta'
  } finally {
    pageLoading.value = false
  }
})
</script>

<template>
  <PageContainer>
    <!-- ═══════════════════ LIST VIEW ═══════════════════════════════════════ -->
    <template v-if="view === 'list'">
      <header class="mb-6">
        <div class="flex items-center justify-between">
          <h1 class="font-serif text-2xl font-semibold text-ink">Ruleta</h1>
          <span class="text-eyebrow">{{ group.group?.name ?? '' }}</span>
        </div>
        <p class="text-xs text-ink-faint mt-0.5">Penitencias del ciclo actual</p>
      </header>

      <div
        v-if="error"
        class="mb-4 rounded-card bg-coral/20 text-coral px-4 py-3 text-sm font-medium"
      >
        {{ error }}
      </div>

      <!-- Cargando -->
      <div v-if="pageLoading" class="flex items-center justify-center py-20">
        <div
          class="w-8 h-8 rounded-full border-2 border-sage-deep border-t-transparent animate-spin"
        />
      </div>

      <template v-else>
        <!-- Nada en juego: la ruleta duerme hasta el sábado -->
        <div
          v-if="rouletteAsleep"
          class="rounded-card bg-paper shadow-card px-6 py-12 text-center mb-7"
        >
          <RouletteWheel
            :count="8"
            :colors="WHEEL_COLORS"
            :size="160"
            sleeping
            class="mx-auto mb-4"
          />
          <template v-if="isWeekend">
            <p class="font-serif text-2xl font-semibold text-ink mb-1">
              La ruleta se queda con hambre
            </p>
            <p class="text-sm text-ink-soft">Nadie falló esta semana. Squad limpio.</p>
          </template>
          <template v-else>
            <p class="font-serif text-2xl font-semibold text-ink mb-1">La ruleta duerme…</p>
            <p class="text-sm text-ink-soft">Despierta el sábado. Que no te encuentre.</p>
          </template>
        </div>

        <template v-else>
          <!-- Ruletas abiertas: cualquiera del squad puede entrar a sugerir -->
          <section v-if="penalties.openEntries.length" class="mb-7">
            <h2 class="text-eyebrow mb-3">RULETAS ABIERTAS</h2>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div
                v-for="e in penalties.openEntries"
                :key="e.id"
                class="rounded-card shadow-flat bg-surface p-4 flex items-center gap-3 border-l-4 border-l-terracotta"
              >
                <BaseAvatar :name="entryDebtorName(e)" size="md" />
                <div class="flex-1 min-w-0">
                  <p class="font-semibold text-sm text-ink truncate">{{ entryDebtorName(e) }}</p>
                  <p class="text-xs text-ink-soft mt-0.5">{{ entryCountdown(e) }}</p>
                </div>
                <button
                  class="rounded-pill bg-terracotta text-paper px-4 py-2 text-xs font-bold transition-all duration-200 hover:opacity-90 active:scale-95 active:opacity-80 flex-shrink-0"
                  :disabled="loading"
                  @click="enterEntry(e)"
                >
                  {{ loading ? '...' : 'Entrar →' }}
                </button>
              </div>
            </div>
          </section>

          <!-- En deuda esta semana -->
          <section class="mb-7">
            <h2 class="text-eyebrow mb-3">EN EL BOTE ESTA SEMANA</h2>

            <div
              v-if="!eligibleWithoutEntry.length"
              class="rounded-card border border-sage/30 bg-sage-soft px-4 py-8 text-center"
            >
              <template v-if="penalties.openEntries.length">
                <Dices class="w-12 h-12 mx-auto mb-3 text-sage-deep opacity-80" />
                <p class="text-sm font-semibold text-sage-deep">
                  Todos los del bote ya tienen ruleta
                </p>
                <p class="text-xs text-ink-soft mt-1">Entra arriba a proponer su penitencia.</p>
              </template>
              <template v-else>
                <PartyPopper class="w-12 h-12 mx-auto mb-3 text-sage-deep opacity-80" />
                <p class="text-sm font-semibold text-sage-deep">Squad limpio esta semana</p>
                <p class="text-xs text-ink-soft mt-1">Nadie falló ningún hábito.</p>
              </template>
            </div>

            <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div
                v-for="m in eligibleWithoutEntry"
                :key="m.user_id"
                class="rounded-card shadow-flat bg-surface p-4 flex items-center gap-3"
              >
                <BaseAvatar :name="m.display_name" size="md" />
                <div class="flex-1 min-w-0">
                  <p class="font-semibold text-sm text-ink truncate">{{ m.display_name }}</p>
                  <p class="text-xs text-ink-soft mt-0.5">Falló esta semana</p>
                </div>
                <button
                  class="rounded-pill bg-terracotta text-paper px-4 py-2 text-xs font-bold transition-all duration-200 hover:opacity-90 active:scale-95 active:opacity-80 flex-shrink-0"
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

            <div
              v-if="!penalties.debts.length"
              class="rounded-card bg-surface border border-hairline px-4 py-5 text-center text-sm text-ink-soft"
            >
              Sin deudas activas en el grupo.
            </div>

            <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div
                v-for="debt in penalties.debts"
                :key="debt.id"
                class="relative overflow-hidden rounded-card shadow-flat bg-paper border border-hairline p-4 pl-5"
              >
                <div class="absolute inset-y-0 left-0 w-1.5 bg-terracotta" />
                <div class="flex items-center justify-between gap-2 mb-2">
                  <div class="flex items-center gap-2 min-w-0">
                    <BaseAvatar :name="memberName(debt.debtor_id)" size="sm" />
                    <span class="text-xs font-semibold text-ink truncate">
                      {{ memberName(debt.debtor_id) }}
                    </span>
                    <span
                      v-if="debt.scope === 'collective'"
                      class="rounded-pill bg-coral-soft text-coral-deep text-[10px] font-bold px-2 py-0.5 flex-shrink-0"
                    >
                      colectiva
                    </span>
                  </div>
                  <span
                    class="rounded-pill text-[10px] font-bold px-2 py-0.5 flex-shrink-0"
                    :class="
                      expiresSoon(debt)
                        ? 'bg-coral-soft text-coral-deep'
                        : 'bg-amber-soft text-amber-deep'
                    "
                  >
                    expira {{ shortDate(debt.expires_at) }}
                  </span>
                </div>
                <p class="text-eyebrow text-terracotta mb-1">LA RULETA DICTÓ</p>
                <p class="font-serif text-lg font-semibold text-ink leading-snug">
                  {{ debt.punishment_emoji ?? '' }} {{ debt.punishment_text }}
                </p>
                <button
                  v-if="debt.debtor_id === auth.user?.id"
                  class="mt-3 w-full rounded-pill py-2 text-xs font-bold transition-colors"
                  :class="
                    confirmComplete === debt.id
                      ? 'bg-sage-deep text-paper'
                      : 'border border-sage-deep text-sage-deep'
                  "
                  :disabled="completing === debt.id"
                  @click="completeDebt(debt)"
                >
                  {{
                    completing === debt.id
                      ? 'Guardando…'
                      : confirmComplete === debt.id
                        ? '¿Seguro? El squad lo verá'
                        : '✓ Ya cumplí mi penitencia'
                  }}
                </button>
                <DoveHero v-if="confirmForgive === debt.id" :size="76" class="mx-auto mt-3" />
                <button
                  v-else-if="debt.debtor_id !== auth.user?.id"
                  class="mt-3 w-full rounded-pill py-2 text-xs font-bold transition-colors border border-hairline text-ink-soft"
                  :disabled="forgiving === debt.id"
                  @click="forgiveDebt(debt)"
                >
                  Perdonar
                </button>
                <button
                  v-if="confirmForgive === debt.id"
                  class="mt-2 w-full rounded-pill py-2 text-xs font-bold transition-colors bg-amber-deep text-paper"
                  :disabled="forgiving === debt.id"
                  @click="forgiveDebt(debt)"
                >
                  {{ forgiving === debt.id ? 'Guardando…' : '¿Seguro? La deuda muere aquí' }}
                </button>
              </div>
            </div>
          </section>
        </template>

        <!-- Historial de deudas -->
        <section class="mt-7">
          <button
            class="flex items-center gap-2 text-eyebrow text-ink-soft mb-3 active:opacity-70"
            @click="toggleHistory"
          >
            HISTORIAL
            <span class="text-ink-faint">{{ showHistory ? '▲' : '▼' }}</span>
          </button>

          <template v-if="showHistory">
            <div
              v-if="!penalties.resolvedDebts.length"
              class="rounded-card bg-surface border border-hairline px-4 py-5 text-center text-sm text-ink-soft"
            >
              Sin deudas pasadas. El historial del squad aparecerá aquí.
            </div>

            <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div
                v-for="debt in penalties.resolvedDebts"
                :key="debt.id"
                class="rounded-card bg-surface border border-hairline p-4"
              >
                <div class="flex items-center justify-between mb-2">
                  <div class="flex items-center gap-2">
                    <BaseAvatar :name="memberName(debt.debtor_id)" size="sm" class="opacity-70" />
                    <span class="text-xs font-semibold text-ink">{{
                      memberName(debt.debtor_id)
                    }}</span>
                  </div>
                  <span
                    class="rounded-full text-[10px] font-bold px-2.5 py-1"
                    :class="DEBT_STATUS_BADGE[debt.status]?.class ?? 'bg-cream-2 text-ink-faint'"
                  >
                    {{ DEBT_STATUS_BADGE[debt.status]?.label ?? debt.status }}
                  </span>
                </div>
                <p class="text-sm text-ink-soft">
                  {{ debt.punishment_emoji ?? '' }} {{ debt.punishment_text }}
                </p>
              </div>
            </div>
          </template>
        </section>
      </template>
    </template>

    <!-- ═══════════════════ ENTRY VIEW ════════════════════════════════════════ -->
    <template v-else-if="view === 'entry' && entry">
      <div class="max-w-md mx-auto">
        <header class="flex items-center gap-3 mb-5">
          <button
            class="w-9 h-9 rounded-full flex items-center justify-center bg-surface border border-hairline"
            @click="backToList"
          >
            <svg class="w-4 h-4 text-ink" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2.5"
                d="M15 19l-7-7 7-7"
              />
            </svg>
          </button>
          <div>
            <h1 class="font-serif text-xl font-semibold text-ink leading-tight">
              Ruleta de {{ debtorName }}
            </h1>
            <p class="text-xs text-ink-soft">Semana actual</p>
          </div>
        </header>

        <div
          v-if="error"
          class="mb-4 rounded-card bg-coral/20 text-coral px-4 py-3 text-sm font-medium"
        >
          {{ error }}
        </div>

        <!-- ── RESULTADO DEL SPIN ─────────────────────────────────────────────── -->
        <template v-if="spinResult">
          <!-- Rueda detenida, atenuada -->
          <div class="flex justify-center mb-5 opacity-40">
            <RouletteWheel :count="wheelCount" :colors="WHEEL_COLORS" :size="150" />
          </div>

          <div class="rounded-card shadow-card bg-paper p-6 mb-5 text-center">
            <p class="text-eyebrow text-terracotta mb-2">
              {{ spinDebts[0]?.scope === 'collective' ? 'DEUDA COLECTIVA' : 'LE TOCÓ A' }}
            </p>
            <BaseAvatar
              :name="memberName(spinDebts[0]?.debtor_id)"
              size="lg"
              class="mx-auto mb-3"
            />
            <h2 class="font-serif text-2xl font-semibold text-ink mb-4">
              {{ memberName(spinDebts[0]?.debtor_id) }}
            </h2>
            <div class="rounded-[14px] bg-terracotta/10 border border-terracotta/20 px-5 py-4 mb-3">
              <p class="text-eyebrow text-terracotta mb-1">LA RULETA HA HABLADO</p>
              <p class="font-serif text-xl font-semibold text-ink">
                {{ spinDebts[0]?.punishment_emoji ?? '' }} {{ spinDebts[0]?.punishment_text }}
              </p>
            </div>
            <p v-if="spinDebts[0]?.scope === 'collective'" class="text-xs text-ink-soft mb-1">
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

        <!-- ── PRE-SPIN ──────────────────────────────────────────────────────── -->
        <template v-else>
          <!-- La rueda -->
          <div class="flex flex-col items-center mb-5">
            <RouletteWheel
              :count="wheelCount"
              :colors="WHEEL_COLORS"
              :size="240"
              :deg="spinDeg"
              :spinning="spinning"
            />

            <!-- Countdown chip -->
            <div
              v-if="!deadlinePassed"
              class="mt-3 flex items-center gap-1.5 rounded-pill bg-amber-soft text-amber-deep px-3 py-1 text-xs font-semibold"
            >
              <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2.5"
                  d="M12 6v6l4 2m6-2a10 10 0 11-20 0 10 10 0 0120 0z"
                />
              </svg>
              Gira en {{ deadlineLabel }}
            </div>
            <div
              v-else-if="!entry.spun_at"
              class="mt-3 rounded-pill bg-coral-soft text-coral-deep px-3 py-1 text-xs font-semibold"
            >
              ¡La ruleta está lista!
            </div>
          </div>

          <!-- Suggestions (penitencias en juego) -->
          <section class="mb-4">
            <div class="flex items-center justify-between mb-3">
              <h2 class="text-eyebrow">PENITENCIAS EN JUEGO</h2>
              <span class="text-xs text-ink-soft">{{ penalties.suggestions.length }}</span>
            </div>

            <div
              v-if="!penalties.suggestions.length"
              class="rounded-card bg-surface border border-hairline px-4 py-5 text-center text-sm text-ink-soft mb-3"
            >
              Nadie ha propuesto una penitencia aún.
            </div>

            <div v-else class="flex flex-wrap gap-2 mb-3">
              <span
                v-for="(s, i) in penalties.suggestions"
                :key="s.id"
                class="rounded-full px-3 py-1.5 text-xs font-semibold text-paper"
                :style="{ backgroundColor: WHEEL_COLORS[i % WHEEL_COLORS.length] }"
              >
                {{ s.emoji ? s.emoji + ' ' : '' }}{{ s.text }}
              </span>
            </div>

            <!-- Suggestion form -->
            <template v-if="canSuggest">
              <button
                v-if="!showForm"
                class="w-full rounded-pill border border-sage-deep text-sage-deep py-3 font-bold text-sm transition-all duration-200 hover:opacity-90 active:scale-95 active:opacity-80"
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
                    class="w-14 rounded-xl border border-hairline bg-paper px-3 py-2.5 text-center text-lg focus:outline-none focus:border-sage-deep"
                  />
                  <input
                    v-model="sugText"
                    type="text"
                    placeholder="Ej: 30 sentadillas en público"
                    class="flex-1 rounded-xl border border-hairline bg-paper px-3 py-2.5 text-sm focus:outline-none focus:border-sage-deep"
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

            <div
              v-else-if="!deadlinePassed && isDebtor"
              class="rounded-card bg-surface border border-hairline px-4 py-3 text-sm text-ink-soft text-center"
            >
              Tu squad escribe tus penitencias… tú solo giras.
            </div>

            <div
              v-else-if="!deadlinePassed && hasSuggested"
              class="rounded-pill bg-sage-soft text-sage-deep px-4 py-3 text-sm font-semibold text-center"
            >
              ✓ Ya propusiste tu penitencia
            </div>
          </section>

          <!-- GIRAR -->
          <button
            v-if="canSpin"
            class="w-full rounded-pill bg-terracotta text-paper py-4 font-bold text-base active:opacity-80 transition-all flex items-center justify-center gap-2"
            :class="{ 'opacity-60': spinning }"
            :disabled="spinning"
            @click="doSpin"
          >
            <svg v-if="spinning" class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              />
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
              />
            </svg>
            {{ spinning ? 'Girando…' : isDebtor ? 'Girar la ruleta' : `Girar por ${debtorName}` }}
          </button>

          <div
            v-else-if="deadlinePassed && !isDebtor && !entry.spun_at"
            class="rounded-card bg-amber-soft border border-amber/30 p-4 text-center"
          >
            <p class="text-sm font-semibold text-amber-deep">Esperando que {{ debtorName }} gire</p>
            <p class="text-xs text-ink-soft mt-1">
              Si no gira en 24h, cualquiera del squad podrá girar por él.
            </p>
          </div>

          <div
            v-else-if="entry.spun_at"
            class="rounded-card bg-surface border border-hairline p-4 text-center text-sm text-ink-soft"
          >
            Esta ruleta ya fue girada.
          </div>
        </template>
      </div>
    </template>
  </PageContainer>
</template>
