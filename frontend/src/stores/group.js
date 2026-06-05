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

  function setMemberOnline(userID) {
    onlineMembers.value.add(userID)
  }

  function setMemberOffline(userID) {
    onlineMembers.value.delete(userID)
  }

  return {
    group, members, onlineMembers, myGroups,
    loadMyGroups, loadGroup, autoLoad,
    setMemberOnline, setMemberOffline,
  }
})
