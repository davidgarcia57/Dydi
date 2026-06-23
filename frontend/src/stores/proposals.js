import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

export const useProposalsStore = defineStore('proposals', () => {
  const catalog = ref([]) // Habit[] from /api/habits
  const proposals = ref([]) // Proposal[] for current group
  const voted = ref(new Set()) // proposalIDs the user has already voted on (local)
  const currentGroupID = ref(null) // remembered so vote() can re-fetch the list

  async function loadCatalog() {
    catalog.value = await api('/api/habits')
  }

  async function loadProposals(groupID) {
    currentGroupID.value = groupID
    proposals.value = await api(`/api/groups/${groupID}/proposals`)
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
    await api(`/api/proposals/${proposalID}/vote`, {
      method: 'POST',
      body: JSON.stringify({ approved }),
    })
    voted.value.add(proposalID)
    const p = proposals.value.find((x) => x.id === proposalID)
    if (p && approved) p.vote_count = (p.vote_count ?? 0) + 1
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
