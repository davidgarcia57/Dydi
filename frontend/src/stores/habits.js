import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

export const useHabitsStore = defineStore('habits', () => {
  const todayCheckins = ref([])
  const streaks = ref({})

  async function loadToday(groupID) {
    todayCheckins.value = await api(`/api/habits/checkins/${groupID}/today`)
  }

  async function loadStreaks(userID) {
    streaks.value = await api(`/api/habits/streaks/${userID}`)
  }

  async function checkin(userHabitID, note = '') {
    await api('/api/habits/checkins', {
      method: 'POST',
      body: JSON.stringify({ user_habit_id: userHabitID, note }),
    })
    await loadToday(userHabitID) // refresh
  }

  function updateCheckin(payload) {
    const idx = todayCheckins.value.findIndex(c => c.user_habit_id === payload.user_habit_id)
    if (idx >= 0) todayCheckins.value[idx] = { ...todayCheckins.value[idx], ...payload }
  }

  function updateStreak(payload) {
    streaks.value[payload.userID] = payload.streak
  }

  return { todayCheckins, streaks, loadToday, loadStreaks, checkin, updateCheckin, updateStreak }
})
