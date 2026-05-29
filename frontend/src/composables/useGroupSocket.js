import { useWebSocket } from '@vueuse/core'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import { usePenaltiesStore } from '@/stores/penalties'

export function useGroupSocket(groupID) {
  const auth = useAuthStore()
  const groupStore = useGroupStore()
  const habitsStore = useHabitsStore()
  const penaltiesStore = usePenaltiesStore()

  const url = `${import.meta.env.VITE_WS_URL}/ws/${groupID}?token=${auth.token}`

  const handlers = {
    checkin:         (p) => habitsStore.updateCheckin(p),
    streak_update:   (p) => habitsStore.updateStreak(p),
    member_online:   (p) => groupStore.setMemberOnline(p.userID),
    member_offline:  (p) => groupStore.setMemberOffline(p.userID),
    roulette_start:  (p) => penaltiesStore.startRoulette(p),
    roulette_result: (p) => penaltiesStore.setRouletteResult(p),
    debt_created:    (p) => penaltiesStore.addDebt(p),
  }

  const { status, close } = useWebSocket(url, {
    autoReconnect: { retries: 10, delay: 3000 },
    onMessage(_, event) {
      try {
        const msg = JSON.parse(event.data)
        handlers[msg.type]?.(msg.payload)
      } catch {}
    },
  })

  return { status, close }
}
