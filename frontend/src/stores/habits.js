import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

function dateISO(d) {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function localDateISO() {
  return dateISO(new Date())
}

export const useHabitsStore = defineStore('habits', () => {
  const todayCheckins = ref([])
  const streaks = ref({}) // { [userID]: maxCurrentStreak }
  const weekHistory = ref({}) // { "userID:habitID": Set<"YYYY-MM-DD"> }

  async function loadToday(groupID) {
    todayCheckins.value = await api(`/api/habits/checkins/${groupID}/today?date=${localDateISO()}`)
  }

  // Loads the last 7 days of check-ins so the squad strips reflect real history.
  async function loadWeekHistory(groupID) {
    const to = new Date()
    const from = new Date()
    from.setDate(to.getDate() - 6)
    const list = await api(`/api/habits/history/${groupID}?from=${dateISO(from)}&to=${dateISO(to)}`)
    const map = {}
    for (const e of list ?? []) {
      const key = `${e.user_id}:${e.habit_id}`
      ;(map[key] ??= new Set()).add(e.checked_on)
    }
    weekHistory.value = map
  }

  async function loadStreaks(userID) {
    const list = await api(`/api/habits/streaks/${userID}`)
    const best = Array.isArray(list) ? list.reduce((max, s) => Math.max(max, s.current ?? 0), 0) : 0
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
      (c) => c.user_id === payload.user_id && c.habit_id === payload.habit_id
    )
    if (idx >= 0) todayCheckins.value[idx] = { ...todayCheckins.value[idx], ...payload }
  }

  function updateStreak(payload) {
    if (payload.userID != null) {
      streaks.value = { ...streaks.value, [payload.userID]: payload.streak }
    }
  }

  return {
    todayCheckins,
    streaks,
    weekHistory,
    loadToday,
    loadWeekHistory,
    loadStreaks,
    checkin,
    updateCheckin,
    updateStreak,
  }
})
