import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

export const useGroupStore = defineStore('group', () => {
  const group = ref(null)
  const members = ref([])
  const onlineMembers = ref(new Set())

  async function loadGroup(id) {
    group.value = await api(`/api/groups/${id}`)
    members.value = await api(`/api/groups/${id}/members`)
  }

  function setMemberOnline(userID) {
    onlineMembers.value.add(userID)
  }

  function setMemberOffline(userID) {
    onlineMembers.value.delete(userID)
  }

  return { group, members, onlineMembers, loadGroup, setMemberOnline, setMemberOffline }
})
