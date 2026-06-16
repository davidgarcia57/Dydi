import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { usePenaltiesStore } from '@/stores/penalties'

// Not a Vue composable — call from anywhere (including onMounted).
// Returns { disconnect } to call in the parent's onUnmounted.
export function useGroupSocket(groupID) {
  const auth         = useAuthStore()
  const groupStore   = useGroupStore()
  const habitsStore  = useHabitsStore()
  const penaltiesStore = usePenaltiesStore()

  const url = `${import.meta.env.VITE_WS_URL}/ws/${groupID}?token=${auth.token}`

  // Each handler receives the full Event object (type, groupID, userID, payload).
  // member_online/offline carry userID at the top level; data events carry info in .payload.
  const handlers = {
    checkin:              (msg) => { habitsStore.updateCheckin(msg.payload); if (msg.userID) habitsStore.loadStreaks(msg.userID) },
    streak_update:        (msg) => habitsStore.updateStreak(msg.payload),
    member_online:        (msg) => groupStore.setMemberOnline(msg.userID),
    member_offline:       (msg) => groupStore.setMemberOffline(msg.userID),
    roulette_result:      (msg) => penaltiesStore.setRouletteResult(msg.payload),
    collective_punishment:(msg) => penaltiesStore.setRouletteResult(msg.payload),
    debt_created:         (msg) => penaltiesStore.addDebt(msg.payload),
  }

  let ws             = null
  let reconnectTimer = null
  let closed         = false
  let attempts       = 0
  const MAX_ATTEMPTS = 10

  function scheduleReconnect() {
    if (closed || attempts >= MAX_ATTEMPTS) return
    const wait = Math.min(1000 * 2 ** attempts, 30_000) // exponential backoff, capped
    attempts++
    reconnectTimer = setTimeout(connect, wait)
  }

  function connect() {
    if (closed) return
    ws = new WebSocket(url)

    ws.onopen = () => { attempts = 0; clearTimeout(reconnectTimer) }

    ws.onmessage = ({ data }) => {
      try {
        const msg = JSON.parse(data)
        // Delivery latency for the paper: receive time minus server emit time.
        if (msg.emittedAt) {
          const latency = Date.now() - new Date(msg.emittedAt).getTime()
          if (latency >= 0) console.debug(`[dydi] ws ${msg.type} delivery ${latency}ms`)
        }
        handlers[msg.type]?.(msg)
      } catch {}
    }

    ws.onclose = () => { if (!closed) scheduleReconnect() }

    ws.onerror = () => ws.close()
  }

  function disconnect() {
    closed = true
    clearTimeout(reconnectTimer)
    if (ws) {
      ws.onclose = null
      ws.close()
    }
  }

  connect()
  return { disconnect }
}
