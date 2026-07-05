import { useHabitsStore } from '@/stores/habits'

// Estado semanal L–V derivado de todayCheckins + weekHistory. La ruleta juzga
// lunes a viernes: estos helpers alimentan el badge "en riesgo" y el pulso del
// squad sin pegarle otra vez a la API.

function dateISO(d) {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// Índice Mon-first (0..6) del día de hoy.
export function mondayIndex(date = new Date()) {
  const dow = date.getDay()
  return dow === 0 ? 6 : dow - 1
}

// Fechas L–V de la semana actual que ya pasaron (sin contar hoy).
export function elapsedWeekdays() {
  const today = new Date()
  const todayIdx = mondayIndex(today)
  const dates = []
  for (let i = 0; i < Math.min(todayIdx, 5); i++) {
    const d = new Date(today)
    d.setDate(today.getDate() - todayIdx + i)
    dates.push(dateISO(d))
  }
  return dates
}

// Fallos acumulados L–V del usuario esta semana: días pasados sin check-in,
// sumados por hábito. Requiere weekHistory cargado (habits.loadWeekHistory).
export function missedThisWeek(userID) {
  const habits = useHabitsStore()
  const mine = habits.todayCheckins.filter((c) => c.user_id === userID)
  if (!mine.length) return 0
  const days = elapsedWeekdays()
  let missed = 0
  for (const c of mine) {
    const done = habits.weekHistory[`${c.user_id}:${c.habit_id}`]
    for (const day of days) {
      if (!done?.has(day)) missed++
    }
  }
  return missed
}

// Estado agregado de hoy para un usuario: done si cumplió todo, pending si le
// falta algo, missed si ya se le fue el día. null si no tiene hábitos.
export function todayStatus(userID) {
  const habits = useHabitsStore()
  const mine = habits.todayCheckins.filter((c) => c.user_id === userID)
  if (!mine.length) return null
  if (mine.every((c) => c.status === 'done')) return 'done'
  if (mine.some((c) => c.status === 'pending')) return 'pending'
  return 'missed'
}
