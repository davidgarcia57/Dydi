import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

function localDateISO() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export const useHabitsStore = defineStore('habits', () => {
  const todayCheckins = ref([])
  const streaks = ref({})   // { [userID]: maxCurrentStreak }

  async function loadToday(groupID) {
    todayCheckins.value = await api(`/api/habits/checkins/${groupID}/today?date=${localDateISO()}`)
  }

  async function loadStreaks(userID) {
    const list = await api(`/api/habits/streaks/${userID}`)
    const best = Array.isArray(list)
      ? list.reduce((max, s) => Math.max(max, s.current ?? 0), 0)
      : 0
    streaks.value = { ...streaks.value, [userID]: best }
  }

  async function checkin(groupID, habitID, note = '') {
    const body = { group_id: groupID, habit_id: habitID, checked_on: localDateISO() }
    if (note) body.note = note
    await api('/api/habits/checkins', {
      method: 'POST',
      body: JSON.stringify(body),
    })
    await loadToday(groupID)
  }

  function updateCheckin(payload) {
    const idx = todayCheckins.value.findIndex(
      c => c.user_id === payload.user_id && c.habit_id === payload.habit_id
    )
    if (idx >= 0) todayCheckins.value[idx] = { ...todayCheckins.value[idx], ...payload }
  }

  function updateStreak(payload) {
    if (payload.userID != null) {
      streaks.value = { ...streaks.value, [payload.userID]: payload.streak }
    }
  }

  return { todayCheckins, streaks, loadToday, loadStreaks, checkin, updateCheckin, updateStreak }
})
