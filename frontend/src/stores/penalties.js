import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

export const usePenaltiesStore = defineStore('penalties', () => {
  const debts = ref([])
  const roulette = ref(null)

  async function loadDebts(groupID) {
    debts.value = await api(`/api/penalties/${groupID}`)
  }

  async function spin(groupID) {
    await api('/api/penalties/spin', {
      method: 'POST',
      body: JSON.stringify({ group_id: groupID }),
    })
  }

  function startRoulette(payload) {
    roulette.value = { ...payload, result: null, spinning: true }
  }

  function setRouletteResult(payload) {
    roulette.value = { ...roulette.value, ...payload, spinning: false }
  }

  function addDebt(payload) {
    debts.value.unshift(payload)
  }

  return { debts, roulette, loadDebts, spin, startRoulette, setRouletteResult, addDebt }
})
