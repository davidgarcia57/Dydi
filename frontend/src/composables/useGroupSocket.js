import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { usePenaltiesStore } from '@/stores/penalties'

// Not a Vue composable — call from anywhere (including onMounted).
// Returns { disconnect } to call in the parent's onUnmounted.
export function useGroupSocket(groupID) {
  const auth = useAuthStore()
  const groupStore = useGroupStore()
  const habitsStore = useHabitsStore()
  const penaltiesStore = usePenaltiesStore()

  // Each handler receives the full Event object (type, groupID, userID, payload).
  // member_online/offline carry userID at the top level; data events carry info in .payload.
  const handlers = {
    // El streak_update llega justo después del checkin con la racha ya
    // recalculada por el backend — aquí ya no se re-consulta /streaks.
    checkin: (msg) => habitsStore.updateCheckin(msg.payload),
    streak_update: (msg) => habitsStore.updateStreak(msg.payload),
    member_online: (msg) => groupStore.setMemberOnline(msg.userID),
    member_offline: (msg) => groupStore.setMemberOffline(msg.userID),
    roulette_start: (msg) => penaltiesStore.addOpenEntry(msg.payload),
    roulette_result: (msg) => penaltiesStore.setRouletteResult(msg.payload),
    debt_updated: (msg) => penaltiesStore.updateDebt(msg.payload),
  }

  let ws = null
  let reconnectTimer = null
  let closed = false
  let attempts = 0
  const MAX_ATTEMPTS = 10

  function scheduleReconnect() {
    if (closed || attempts >= MAX_ATTEMPTS) return
    const wait = Math.min(1000 * 2 ** attempts, 30_000) // exponential backoff, capped
    attempts++
    reconnectTimer = setTimeout(connect, wait)
  }

  async function connect() {
    if (closed) return
    // Build the URL on every connect so the token is always fresh.
    // Supabase rotates access_tokens (~1h); a stale token would fail the
    // handshake silently and burn through all reconnection attempts.
    const accessToken = await auth.getAccessToken().catch(() => null)
    if (!accessToken) {
      scheduleReconnect()
      return
    }
    const url = `${import.meta.env.VITE_WS_URL}/ws/${groupID}?token=${encodeURIComponent(accessToken)}`
    ws = new WebSocket(url)

    ws.onopen = () => {
      attempts = 0
      clearTimeout(reconnectTimer)
      groupStore.setRealtimeState('connected')
    }

    ws.onmessage = ({ data }) => {
      try {
        const msg = JSON.parse(data)
        handlers[msg.type]?.(msg)
      } catch (e) {
        console.warn('[dydi] ws message handler error:', e)
      }
    }

    ws.onclose = () => {
      if (!closed) {
        groupStore.setRealtimeState('reconnecting')
        scheduleReconnect()
      }
    }

    ws.onerror = () => ws.close()
  }

  function disconnect() {
    closed = true
    clearTimeout(reconnectTimer)
    groupStore.setRealtimeState(null)
    if (ws) {
      ws.onclose = null
      ws.close()
    }
  }

  connect()
  return { disconnect }
}
