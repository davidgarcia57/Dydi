<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { useProposalsStore } from '@/stores/proposals'
import PageContainer from '@/components/ui/PageContainer.vue'
import HabitIcon from '@/components/ui/HabitIcon.vue'

const router = useRouter()
const group = useGroupStore()
const habits = useHabitsStore()
const store = useProposalsStore()

const tab = ref('catalogo') // 'catalogo' | 'propuestas' | 'historial'
const loading = ref(true)
const historyLoaded = ref(false)
const proposing = ref(null) // habitID currently being proposed
const proposeErr = ref('')
const proposeOk = ref(null) // habitID of last successful proposal
const votingID = ref(null) // proposalID being voted on
const voteErr = ref('')

// Habits already in use by the group today (cross-ref with checkins)
const assignedHabitIDs = computed(() => {
  return new Set(habits.todayCheckins.map((c) => c.habit_id))
})

const availableHabits = computed(() => {
  return store.catalog.filter((h) => !assignedHabitIDs.value.has(h.id))
})

const activeHabits = computed(() => {
  return store.catalog.filter((h) => assignedHabitIDs.value.has(h.id))
})

const PROPOSAL_LABEL = {
  add_habit: 'Agregar hábito',
  remove_habit: 'Quitar hábito',
  kick_member: 'Expulsar miembro',
  delete_group: 'Disolver grupo',
}

const STATUS_BADGE = {
  approved: { label: 'APROBADA', class: 'bg-sage-soft text-sage-deep' },
  rejected: { label: 'RECHAZADA', class: 'bg-coral-soft text-coral-deep' },
  expired: { label: 'EXPIRÓ', class: 'bg-cream-2 text-ink-faint' },
}

// Carga perezosa: el historial solo se pide al abrir su tab.
async function openHistory() {
  tab.value = 'historial'
  if (historyLoaded.value || !group.group?.id) return
  try {
    await store.loadResolved(group.group.id)
    historyLoaded.value = true
  } catch (_) {
    // el empty-state del tab cubre el fallo; reintenta al volver a entrar
  }
}

function habitName(habitID) {
  return store.catalog.find((h) => h.id === habitID)?.name ?? habitID
}

function memberName(userID) {
  return group.members.find((m) => m.user_id === userID)?.display_name ?? 'Miembro'
}

function voteProgress(p) {
  if (!p.member_count) return 0
  return Math.round((p.vote_count / p.member_count) * 100)
}

function quorumLabel(p) {
  const need = Math.ceil(p.member_count / 2)
  return `${p.vote_count} de ${need} votos necesarios`
}

async function propose(habit, type = 'add_habit') {
  if (proposing.value || proposeOk.value === habit.id) return
  proposing.value = habit.id
  proposeErr.value = ''
  try {
    await store.propose(group.group.id, type, { habitID: habit.id })
    proposeOk.value = habit.id
    tab.value = 'propuestas'
  } catch (e) {
    proposeErr.value = e?.error ?? e?.message ?? 'No se pudo crear la propuesta.'
  } finally {
    proposing.value = null
  }
}

async function castVote(proposalID, approved) {
  votingID.value = proposalID
  voteErr.value = ''
  try {
    await store.vote(proposalID, approved)
  } catch (e) {
    voteErr.value = e?.error ?? e?.message ?? 'No se pudo registrar el voto.'
  } finally {
    votingID.value = null
  }
}

onMounted(async () => {
  loading.value = true
  try {
    await group.autoLoad()
    if (!group.group?.id) {
      router.replace('/onboarding')
      return
    }
    await Promise.all([
      store.loadCatalog(),
      store.loadProposals(group.group.id),
      habits.todayCheckins.length === 0 ? habits.loadToday(group.group.id) : Promise.resolve(),
    ])
  } catch (_) {
    // errors surfaced inline; page renders with whatever loaded
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <PageContainer>
    <!-- ── Header ─────────────────────────────────────────────────────────── -->
    <header class="mb-6">
      <div class="flex items-center justify-between">
        <h1 class="font-serif text-2xl font-semibold text-ink">Propuestas</h1>
        <span class="text-eyebrow">{{ group.group?.name ?? '' }}</span>
      </div>
      <p class="text-xs text-ink-faint mt-0.5">Propón y vota hábitos para el squad</p>
    </header>

    <!-- ── Loading ─────────────────────────────────────────────────────────── -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <div
        class="w-8 h-8 rounded-full border-2 border-sage-deep border-t-transparent animate-spin"
      />
    </div>

    <template v-else>
      <!-- ── Tabs ─────────────────────────────────────────────────────────── -->
      <div class="flex gap-1 bg-cream-2 rounded-[14px] p-1 mb-6 max-w-md">
        <button
          class="flex-1 rounded-[10px] py-2 text-sm font-semibold transition-all"
          :class="tab === 'catalogo' ? 'bg-paper shadow-flat text-ink' : 'text-ink-soft'"
          @click="tab = 'catalogo'"
        >
          Catálogo
        </button>
        <button
          class="flex-1 rounded-[10px] py-2 text-sm font-semibold transition-all relative"
          :class="tab === 'propuestas' ? 'bg-paper shadow-flat text-ink' : 'text-ink-soft'"
          @click="tab = 'propuestas'"
        >
          Propuestas
          <span
            v-if="store.proposals.length"
            class="absolute -top-0.5 -right-0.5 w-4 h-4 rounded-full bg-terracotta text-paper text-[9px] font-bold flex items-center justify-center"
          >
            {{ store.proposals.length }}
          </span>
        </button>
        <button
          class="flex-1 rounded-[10px] py-2 text-sm font-semibold transition-all"
          :class="tab === 'historial' ? 'bg-paper shadow-flat text-ink' : 'text-ink-soft'"
          @click="openHistory"
        >
          Historial
        </button>
      </div>

      <!-- ── Catálogo de hábitos ──────────────────────────────────────────── -->
      <div v-if="tab === 'catalogo'">
        <p class="text-xs text-ink-soft mb-4">
          Propón un hábito para todo el squad. Requiere votación mayoritaria.
        </p>

        <p v-if="proposeErr" class="text-sm text-coral mb-4 font-medium">{{ proposeErr }}</p>

        <div
          v-if="!store.catalog.length"
          class="rounded-card bg-surface py-10 text-center text-sm text-ink-soft"
        >
          No hay hábitos en el catálogo todavía.
        </div>

        <div v-else class="space-y-6">
          <!-- HÁBITOS DISPONIBLES -->
          <div v-if="availableHabits.length > 0">
            <h3 class="text-eyebrow text-ink-soft mb-3">DISPONIBLES PARA AÑADIR</h3>
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-2">
              <div
                v-for="habit in availableHabits"
                :key="habit.id"
                class="rounded-card bg-paper shadow-flat p-4 flex items-center gap-3"
              >
                <div
                  class="w-10 h-10 rounded-full flex-shrink-0 flex items-center justify-center text-paper"
                  :style="{ backgroundColor: habit.color || 'var(--color-sage)' }"
                >
                  <HabitIcon :icon-key="habit.icon_key" :size="22" />
                </div>

                <div class="flex-1 min-w-0">
                  <p class="font-semibold text-sm text-ink truncate">{{ habit.name }}</p>
                  <p v-if="habit.description" class="text-xs text-ink-soft truncate mt-0.5">
                    {{ habit.description }}
                  </p>
                </div>

                <button
                  v-if="proposeOk === habit.id"
                  class="flex-shrink-0 text-[10px] font-bold text-sage-deep bg-sage-soft rounded-full px-3 py-1.5"
                >
                  ✓ Propuesto
                </button>
                <button
                  v-else
                  :disabled="proposing === habit.id"
                  class="flex-shrink-0 rounded-pill border border-hairline bg-surface text-ink-soft text-xs font-semibold px-3 py-1.5 active:opacity-70 disabled:opacity-40 transition-opacity"
                  @click="propose(habit, 'add_habit')"
                >
                  <span v-if="proposing === habit.id" class="flex items-center gap-1">
                    <span
                      class="w-3 h-3 rounded-full border border-ink-soft border-t-transparent animate-spin"
                    />
                  </span>
                  <span v-else>+ Proponer</span>
                </button>
              </div>
            </div>
          </div>

          <!-- HÁBITOS ACTIVOS -->
          <div v-if="activeHabits.length > 0">
            <h3 class="text-eyebrow text-ink-soft mb-3">ACTIVOS EN EL SQUAD</h3>
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-2">
              <div
                v-for="habit in activeHabits"
                :key="habit.id"
                class="rounded-card bg-surface border border-hairline p-4 flex items-center gap-3"
              >
                <div
                  class="w-10 h-10 rounded-full flex-shrink-0 flex items-center justify-center text-paper opacity-60"
                  :style="{ backgroundColor: habit.color || 'var(--color-sage)' }"
                >
                  <HabitIcon :icon-key="habit.icon_key" :size="22" />
                </div>

                <div class="flex-1 min-w-0">
                  <p class="font-semibold text-sm text-ink truncate">{{ habit.name }}</p>
                  <span
                    class="inline-block mt-1 text-[10px] font-semibold text-sage-deep bg-sage-soft rounded-full px-2 py-0.5"
                  >
                    Ya en el grupo
                  </span>
                </div>

                <button
                  v-if="proposeOk === habit.id"
                  class="flex-shrink-0 text-[10px] font-bold text-sage-deep bg-sage-soft rounded-full px-3 py-1.5"
                >
                  ✓ Propuesto
                </button>
                <button
                  v-else
                  :disabled="proposing === habit.id"
                  class="flex-shrink-0 rounded-pill border border-coral/30 bg-coral-soft/50 text-coral-deep text-xs font-semibold px-3 py-1.5 active:opacity-70 disabled:opacity-40 transition-opacity"
                  @click="propose(habit, 'remove_habit')"
                >
                  <span v-if="proposing === habit.id" class="flex items-center gap-1">
                    <span
                      class="w-3 h-3 rounded-full border border-coral-deep border-t-transparent animate-spin"
                    />
                  </span>
                  <span v-else>- Quitar</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── Propuestas activas ───────────────────────────────────────────── -->
      <div v-else-if="tab === 'propuestas'">
        <div
          v-if="!store.proposals.length"
          class="rounded-card bg-paper shadow-flat py-14 text-center"
        >
          <p class="text-eyebrow text-ink-faint mb-2">SIN PROPUESTAS</p>
          <p class="font-serif text-xl font-semibold text-ink mb-1">Todo tranquilo</p>
          <p class="text-sm text-ink-soft mt-1">Propón un hábito desde el catálogo.</p>
        </div>

        <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-3 items-start">
          <p v-if="voteErr" class="text-sm text-coral font-medium lg:col-span-2">{{ voteErr }}</p>

          <div
            v-for="p in store.proposals"
            :key="p.id"
            class="rounded-card bg-paper shadow-flat p-5"
          >
            <!-- Type badge -->
            <div class="flex items-start justify-between gap-2 mb-3">
              <div>
                <span class="text-eyebrow text-ink-soft">
                  {{ PROPOSAL_LABEL[p.type] ?? p.type }}
                </span>
                <p v-if="p.habit_id" class="font-semibold text-sm text-ink mt-0.5">
                  {{ habitName(p.habit_id) }}
                </p>
                <p v-else-if="p.target_user_id" class="font-semibold text-sm text-ink mt-0.5">
                  {{ memberName(p.target_user_id) }}
                </p>
                <p v-else-if="p.type === 'delete_group'" class="text-xs text-ink-soft mt-0.5">
                  Si gana, el grupo se elimina para todos
                </p>
              </div>
              <span
                class="rounded-full bg-amber-soft text-amber-deep text-[10px] font-bold px-2.5 py-1 flex-shrink-0"
              >
                ABIERTA
              </span>
            </div>

            <!-- Vote progress -->
            <div class="mb-4">
              <div class="flex justify-between text-xs text-ink-soft mb-1">
                <span>{{ quorumLabel(p) }}</span>
                <span>{{ voteProgress(p) }}%</span>
              </div>
              <div class="h-1.5 rounded-full bg-hairline">
                <div
                  class="h-full rounded-full bg-sage-deep transition-all duration-500"
                  :style="{ width: voteProgress(p) + '%' }"
                />
              </div>
            </div>

            <!-- Vote buttons -->
            <div v-if="!store.voted.has(p.id)" class="flex gap-2">
              <button
                :disabled="votingID === p.id"
                class="flex-1 rounded-pill bg-sage-deep text-paper py-2.5 font-bold text-sm disabled:opacity-40 active:opacity-80 transition-opacity"
                @click="castVote(p.id, true)"
              >
                <span v-if="votingID === p.id" class="flex items-center justify-center gap-1">
                  <span
                    class="w-3 h-3 rounded-full border-2 border-paper border-t-transparent animate-spin"
                  />
                </span>
                <span v-else>✓ Aprobar</span>
              </button>
              <button
                :disabled="votingID === p.id"
                class="flex-1 rounded-pill border border-hairline text-ink-soft py-2.5 font-bold text-sm disabled:opacity-40 active:opacity-80 transition-opacity"
                @click="castVote(p.id, false)"
              >
                ✗ Rechazar
              </button>
            </div>
            <div
              v-else
              class="rounded-pill bg-sage-soft text-sage-deep text-sm font-semibold py-2.5 text-center"
            >
              ✓ Ya votaste
            </div>
          </div>
        </div>
      </div>

      <!-- ── Historial de decisiones ──────────────────────────────────────── -->
      <div v-else-if="tab === 'historial'">
        <div
          v-if="!store.resolved.length"
          class="rounded-card bg-paper shadow-flat py-14 text-center"
        >
          <p class="text-eyebrow text-ink-faint mb-2">SIN HISTORIAL</p>
          <p class="font-serif text-xl font-semibold text-ink mb-1">Nada decidido aún</p>
          <p class="text-sm text-ink-soft mt-1">Las propuestas cerradas aparecerán aquí.</p>
        </div>

        <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-3 items-start">
          <div
            v-for="p in store.resolved"
            :key="p.id"
            class="rounded-card bg-surface border border-hairline p-5"
          >
            <div class="flex items-start justify-between gap-2 mb-3">
              <div>
                <span class="text-eyebrow text-ink-soft">
                  {{ PROPOSAL_LABEL[p.type] ?? p.type }}
                </span>
                <p v-if="p.habit_id" class="font-semibold text-sm text-ink mt-0.5">
                  {{ habitName(p.habit_id) }}
                </p>
                <p v-else-if="p.target_user_id" class="font-semibold text-sm text-ink mt-0.5">
                  {{ memberName(p.target_user_id) }}
                </p>
                <p v-else-if="p.type === 'delete_group'" class="text-xs text-ink-soft mt-0.5">
                  Si gana, el grupo se elimina para todos
                </p>
              </div>
              <span
                class="rounded-full text-[10px] font-bold px-2.5 py-1 flex-shrink-0"
                :class="STATUS_BADGE[p.status]?.class ?? 'bg-cream-2 text-ink-faint'"
              >
                {{ STATUS_BADGE[p.status]?.label ?? p.status }}
              </span>
            </div>
            <p class="text-xs text-ink-soft">
              {{ p.vote_count }} de {{ p.member_count }} votos a favor ·
              {{
                new Date(p.created_at).toLocaleDateString('es-MX', {
                  month: 'short',
                  day: 'numeric',
                })
              }}
            </p>
          </div>
        </div>
      </div>
    </template>
  </PageContainer>
</template>
