import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

export const usePenaltiesStore = defineStore('penalties', () => {
  const debts = ref([])
  const eligible = ref([])
  const openEntries = ref([])
  const activeEntry = ref(null)
  const suggestions = ref([])

  async function loadDebts(groupID) {
    debts.value = await api(`/api/penalties/${groupID}/debts`)
  }

  async function loadEligible(groupID) {
    eligible.value = await api(`/api/penalties/${groupID}/eligible`)
  }

  // Ruletas ya abiertas y sin girar: visibles para TODO el grupo, aunque la
  // elegibilidad de la semana ya haya expirado (abrirla de nuevo daría 409).
  async function loadOpenEntries(groupID) {
    openEntries.value = await api(`/api/penalties/${groupID}/roulette`)
  }

  // Entra a una ruleta ya abierta sin re-abrirla (POST exige elegibilidad).
  function enterEntry(entry) {
    activeEntry.value = entry
  }

  async function openRoulette(groupID, debtorID) {
    activeEntry.value = await api('/api/penalties/roulette', {
      method: 'POST',
      body: JSON.stringify({ group_id: groupID, debtor_id: debtorID }),
    })
    if (!openEntries.value.find((e) => e.id === activeEntry.value.id)) {
      openEntries.value.unshift(activeEntry.value)
    }
    return activeEntry.value
  }

  async function completeDebt(debtID) {
    const debt = await api(`/api/penalties/debts/${debtID}/complete`, { method: 'POST' })
    debts.value = debts.value.filter((d) => d.id !== debt.id)
    return debt
  }

  async function loadSuggestions(entryID) {
    suggestions.value = await api(`/api/penalties/roulette/${entryID}/suggestions`)
  }

  async function submitSuggestion(entryID, text, emoji = null) {
    const s = await api(`/api/penalties/roulette/${entryID}/suggestions`, {
      method: 'POST',
      body: JSON.stringify({ text, ...(emoji ? { emoji } : {}) }),
    })
    suggestions.value.push(s)
    return s
  }

  // Returns a single Debt (normal) or Debt[] (collective — no suggestions).
  async function spin(entryID) {
    const result = await api(`/api/penalties/roulette/${entryID}/spin`, {
      method: 'POST',
    })
    const added = Array.isArray(result) ? result : [result]
    debts.value.unshift(...added)
    return result
  }

  function clearEntry() {
    activeEntry.value = null
    suggestions.value = []
  }

  // WebSocket event handlers
  function setRouletteResult(payload) {
    if (activeEntry.value) {
      activeEntry.value = { ...activeEntry.value, spun_at: new Date().toISOString() }
    }
    const added = Array.isArray(payload) ? payload : [payload]
    for (const d of added) {
      if (!debts.value.find((x) => x.id === d.id)) debts.value.unshift(d)
      openEntries.value = openEntries.value.filter((e) => e.id !== d.roulette_entry_id)
    }
  }

  function addDebt(payload) {
    if (!debts.value.find((d) => d.id === payload.id)) debts.value.unshift(payload)
  }

  function addOpenEntry(payload) {
    if (!openEntries.value.find((e) => e.id === payload.id)) {
      openEntries.value.unshift(payload)
    }
  }

  function updateDebt(payload) {
    // Las deudas activas solo listan status=pending: al completarse, sale.
    if (payload.status !== 'pending') {
      debts.value = debts.value.filter((d) => d.id !== payload.id)
      return
    }
    debts.value = debts.value.map((d) => (d.id === payload.id ? payload : d))
  }

  return {
    debts,
    eligible,
    openEntries,
    activeEntry,
    suggestions,
    loadDebts,
    loadEligible,
    loadOpenEntries,
    enterEntry,
    openRoulette,
    completeDebt,
    loadSuggestions,
    submitSuggestion,
    spin,
    clearEntry,
    setRouletteResult,
    addDebt,
    addOpenEntry,
    updateDebt,
  }
})
