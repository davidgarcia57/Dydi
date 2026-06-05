import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

export const usePenaltiesStore = defineStore('penalties', () => {
  const debts = ref([])
  const eligible = ref([])
  const activeEntry = ref(null)
  const suggestions = ref([])

  async function loadDebts(groupID) {
    debts.value = await api(`/api/penalties/${groupID}/debts`)
  }

  async function loadEligible(groupID) {
    eligible.value = await api(`/api/penalties/${groupID}/eligible`)
  }

  async function openRoulette(groupID, debtorID) {
    activeEntry.value = await api('/api/penalties/roulette', {
      method: 'POST',
      body: JSON.stringify({ group_id: groupID, debtor_id: debtorID }),
    })
    return activeEntry.value
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
      if (!debts.value.find(x => x.id === d.id)) debts.value.unshift(d)
    }
  }

  function addDebt(payload) {
    if (!debts.value.find(d => d.id === payload.id)) debts.value.unshift(payload)
  }

  return {
    debts, eligible, activeEntry, suggestions,
    loadDebts, loadEligible, openRoulette, loadSuggestions,
    submitSuggestion, spin, clearEntry,
    setRouletteResult, addDebt,
  }
})
