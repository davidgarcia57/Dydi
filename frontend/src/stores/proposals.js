import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

export const useProposalsStore = defineStore('proposals', () => {
  const catalog = ref([]) // Habit[] from /api/habits
  const proposals = ref([]) // Proposal[] for current group
  const voted = ref(new Set()) // proposalIDs the user has already voted on
  const currentGroupID = ref(null) // remembered so vote() can re-fetch the list

  async function loadCatalog() {
    catalog.value = await api('/api/habits')
  }

  async function loadProposals(groupID) {
    currentGroupID.value = groupID
    const list = await api(`/api/groups/${groupID}/proposals`)
    proposals.value = list
    // Sync voted set from server — survives page reloads
    voted.value = new Set(list.filter((p) => p.user_voted).map((p) => p.id))
  }

  async function propose(groupID, type, habitID = null) {
    const body = { type }
    if (habitID) body.habit_id = habitID
    const p = await api(`/api/groups/${groupID}/proposals`, {
      method: 'POST',
      body: JSON.stringify(body),
    })
    proposals.value.unshift(p)
    return p
  }

  async function vote(proposalID, approved) {
    try {
      await api(`/api/proposals/${proposalID}/vote`, {
        method: 'POST',
        body: JSON.stringify({ approved }),
      })
    } catch (e) {
      // 409 = already voted, proposal closed, or expired — just re-sync
      if (e?.status === 409) {
        voted.value.add(proposalID)
        if (currentGroupID.value) await loadProposals(currentGroupID.value)
        return
      }
      throw e
    }
    voted.value.add(proposalID)
    // Re-fetch the authoritative list: a proposal that just reached quorum flips
    // to status != 'open' server-side and drops out, so it stops showing here.
    if (currentGroupID.value) await loadProposals(currentGroupID.value)
  }

  function reset() {
    proposals.value = []
    voted.value = new Set()
  }

  return { catalog, proposals, voted, loadCatalog, loadProposals, propose, vote, reset }
})

