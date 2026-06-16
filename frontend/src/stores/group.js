import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

export const useGroupStore = defineStore('group', () => {
  const group = ref(null)         // { id, name, invite_code, created_at }
  const members = ref([])         // [{ user_id, display_name, ... }]
  const onlineMembers = ref(new Set())
  const myGroups = ref([])        // lightweight list: [{ id, name }]

  async function loadMyGroups() {
    myGroups.value = await api('/api/groups')
  }

  // GET /api/groups/:id returns GroupWithMembers (fields promoted to top level)
  async function loadGroup(id) {
    const data = await api(`/api/groups/${id}`)
    const { members: mems, ...groupData } = data
    group.value = groupData
    members.value = mems ?? []
  }

  // Loads the first group automatically. Returns true if a group was found.
  async function autoLoad() {
    if (group.value?.id) return true
    await loadMyGroups()
    if (!myGroups.value?.length) return false
    await loadGroup(myGroups.value[0].id)
    return true
  }

  async function createGroup(name) {
    const data = await api('/api/groups', {
      method: 'POST',
      body: JSON.stringify({ name }),
    })
    group.value = data
    members.value = []
    myGroups.value = [data]
    return data
  }

  async function joinGroup(groupID, inviteCode) {
    await api(`/api/groups/${groupID}/join`, {
      method: 'POST',
      body: JSON.stringify({ invite_code: inviteCode }),
    })
    await loadGroup(groupID)
    myGroups.value = [group.value]
  }

  async function leaveGroup() {
    if (!group.value?.id) return
    await api(`/api/groups/${group.value.id}/leave`, { method: 'DELETE' })
    reset()
  }

  function reset() {
    group.value = null
    members.value = []
    myGroups.value = []
    onlineMembers.value = new Set()
  }

  // Reassign the ref (not mutate in place) so Vue's reactivity fires and the
  // squad/today views re-render when presence changes over WebSocket.
  function setMemberOnline(userID) {
    const next = new Set(onlineMembers.value)
    next.add(userID)
    onlineMembers.value = next
  }

  function setMemberOffline(userID) {
    const next = new Set(onlineMembers.value)
    next.delete(userID)
    onlineMembers.value = next
  }

  return {
    group, members, onlineMembers, myGroups,
    loadMyGroups, loadGroup, autoLoad,
    createGroup, joinGroup, leaveGroup, reset,
    setMemberOnline, setMemberOffline,
  }
})
