<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { useProposalsStore } from '@/stores/proposals'
import { useAuthStore } from '@/stores/auth'

const router    = useRouter()
const group     = useGroupStore()
const habits    = useHabitsStore()
const store     = useProposalsStore()
const auth      = useAuthStore()

const tab           = ref('catalogo')   // 'catalogo' | 'propuestas'
const loading       = ref(true)
const proposing     = ref(null)          // habitID currently being proposed
const proposeErr    = ref('')
const proposeOk     = ref(null)          // habitID of last successful proposal
const votingID      = ref(null)          // proposalID being voted on
const voteErr       = ref('')

// Habits already in use by the group today (cross-ref with checkins)
const assignedHabitIDs = computed(() => {
  return new Set(habits.todayCheckins.map(c => c.habit_id))
})

const PROPOSAL_LABEL = {
  add_habit:    'Agregar hábito',
  remove_habit: 'Quitar hábito',
  kick_member:  'Expulsar miembro',
  delete_group: 'Disolver grupo',
}

function habitName(habitID) {
  return store.catalog.find(h => h.id === habitID)?.name ?? habitID
}

function voteProgress(p) {
  if (!p.member_count) return 0
  return Math.round((p.vote_count / p.member_count) * 100)
}

function quorumLabel(p) {
  const need = Math.ceil(p.member_count / 2)
  return `${p.vote_count} de ${need} votos necesarios`
}

async function propose(habit) {
  if (proposing.value || proposeOk.value === habit.id) return
  proposing.value = habit.id
  proposeErr.value = ''
  try {
    await store.propose(group.group.id, 'add_habit', habit.id)
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
  <div class="max-w-md mx-auto px-4 pt-4 pb-6">

    <!-- ── Header ─────────────────────────────────────────────────────────── -->
    <header class="flex items-center justify-between mb-6">
      <h1 class="font-serif text-2xl font-semibold text-ink">Propuestas</h1>
      <span class="text-eyebrow">{{ group.group?.name ?? '' }}</span>
    </header>

    <!-- ── Loading ─────────────────────────────────────────────────────────── -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <div class="w-8 h-8 rounded-full border-2 border-sage-deep border-t-transparent animate-spin" />
    </div>

    <template v-else>
      <!-- ── Tabs ─────────────────────────────────────────────────────────── -->
      <div class="flex gap-1 bg-surface rounded-[14px] p-1 mb-6">
        <button
          class="flex-1 rounded-[10px] py-2 text-sm font-semibold transition-all"
          :class="tab === 'catalogo'
            ? 'bg-paper shadow-flat text-ink'
            : 'text-ink-soft'"
          @click="tab = 'catalogo'"
        >
          Catálogo
        </button>
        <button
          class="flex-1 rounded-[10px] py-2 text-sm font-semibold transition-all relative"
          :class="tab === 'propuestas'
            ? 'bg-paper shadow-flat text-ink'
            : 'text-ink-soft'"
          @click="tab = 'propuestas'"
        >
          Propuestas
          <span v-if="store.proposals.length"
            class="absolute -top-0.5 -right-0.5 w-4 h-4 rounded-full bg-terracotta
                   text-paper text-[9px] font-bold flex items-center justify-center">
            {{ store.proposals.length }}
          </span>
        </button>
      </div>

      <!-- ── Catálogo de hábitos ──────────────────────────────────────────── -->
      <div v-if="tab === 'catalogo'">
        <p class="text-xs text-ink-soft mb-4">
          Propón un hábito para todo el squad. Requiere votación mayoritaria.
        </p>

        <p v-if="proposeErr" class="text-sm text-coral mb-4 font-medium">{{ proposeErr }}</p>

        <div v-if="!store.catalog.length"
          class="rounded-card bg-surface py-10 text-center text-sm text-ink-soft">
          No hay hábitos en el catálogo todavía.
        </div>

        <div v-else class="space-y-2">
          <div
            v-for="habit in store.catalog"
            :key="habit.id"
            class="rounded-card bg-paper shadow-flat p-4 flex items-center gap-3"
          >
            <!-- Color dot -->
            <div
              class="w-10 h-10 rounded-full flex-shrink-0 flex items-center justify-center
                     text-paper text-sm font-bold"
              :style="{ backgroundColor: habit.color || '#A8C39A' }"
            >
              {{ habit.name.charAt(0).toUpperCase() }}
            </div>

            <div class="flex-1 min-w-0">
              <p class="font-semibold text-sm text-ink truncate">{{ habit.name }}</p>
              <p v-if="habit.description" class="text-xs text-ink-soft truncate mt-0.5">
                {{ habit.description }}
              </p>
              <span v-if="assignedHabitIDs.has(habit.id)"
                class="inline-block mt-1 text-[10px] font-semibold text-sage-deep
                       bg-sage/20 rounded-full px-2 py-0.5">
                Ya en el grupo
              </span>
            </div>

            <button
              v-if="proposeOk === habit.id"
              class="flex-shrink-0 text-[10px] font-bold text-sage-deep bg-sage/20
                     rounded-full px-3 py-1.5"
            >
              ✓ Propuesto
            </button>
            <button
              v-else
              :disabled="proposing === habit.id"
              class="flex-shrink-0 rounded-pill border border-hairline bg-surface
                     text-ink-soft text-xs font-semibold px-3 py-1.5
                     active:opacity-70 disabled:opacity-40 transition-opacity"
              @click="propose(habit)"
            >
              <span v-if="proposing === habit.id" class="flex items-center gap-1">
                <span class="w-3 h-3 rounded-full border border-ink-soft
                             border-t-transparent animate-spin" />
              </span>
              <span v-else>+ Proponer</span>
            </button>
          </div>
        </div>
      </div>

      <!-- ── Propuestas activas ───────────────────────────────────────────── -->
      <div v-else-if="tab === 'propuestas'">

        <div v-if="!store.proposals.length"
          class="rounded-card bg-surface py-14 text-center">
          <p class="text-3xl mb-3">🗳</p>
          <p class="text-sm text-ink-soft">
            No hay propuestas abiertas.<br>
            Propón un hábito desde el catálogo.
          </p>
        </div>

        <div v-else class="space-y-3">
          <p v-if="voteErr" class="text-sm text-coral font-medium">{{ voteErr }}</p>

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
              </div>
              <span class="rounded-full bg-amber/20 text-amber text-[10px] font-bold
                           px-2.5 py-1 flex-shrink-0">
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
                class="flex-1 rounded-pill bg-sage-deep text-paper py-2.5 font-bold
                       text-sm disabled:opacity-40 active:opacity-80 transition-opacity"
                @click="castVote(p.id, true)"
              >
                <span v-if="votingID === p.id" class="flex items-center justify-center gap-1">
                  <span class="w-3 h-3 rounded-full border-2 border-paper
                               border-t-transparent animate-spin" />
                </span>
                <span v-else>✓ Aprobar</span>
              </button>
              <button
                :disabled="votingID === p.id"
                class="flex-1 rounded-pill border border-hairline text-ink-soft
                       py-2.5 font-bold text-sm disabled:opacity-40
                       active:opacity-80 transition-opacity"
                @click="castVote(p.id, false)"
              >
                ✗ Rechazar
              </button>
            </div>
            <div v-else
              class="rounded-pill bg-sage/10 text-sage-deep text-sm font-semibold
                     py-2.5 text-center">
              ✓ Ya votaste
            </div>
          </div>
        </div>
      </div>
    </template>

  </div>
</template>
