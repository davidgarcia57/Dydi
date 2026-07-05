// Estado semanal L–V derivado de todayCheckins + weekHistory — espejo del
// composable useWeekStatus del frontend web. La ruleta juzga lunes a viernes:
// estos helpers alimentan el badge "en riesgo" y el pulso del squad sin
// pegarle otra vez a la API.

type CheckinLike = { user_id: string; habit_id: string; status: string };

export function mondayIndex(date = new Date()): number {
  const dow = date.getDay();
  return dow === 0 ? 6 : dow - 1;
}

function dateISO(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(
    d.getDate()
  ).padStart(2, '0')}`;
}

// Fechas L–V de la semana actual que ya pasaron (sin contar hoy).
export function elapsedWeekdays(): string[] {
  const today = new Date();
  const todayIdx = mondayIndex(today);
  const dates: string[] = [];
  for (let i = 0; i < Math.min(todayIdx, 5); i++) {
    const d = new Date(today);
    d.setDate(today.getDate() - todayIdx + i);
    dates.push(dateISO(d));
  }
  return dates;
}

// Fallos acumulados L–V del usuario esta semana: días pasados sin check-in,
// sumados por hábito. Requiere weekHistory cargado.
export function missedThisWeek(
  userID: string,
  todayCheckins: CheckinLike[],
  weekHistory: Record<string, Set<string>>
): number {
  const mine = todayCheckins.filter((c) => c.user_id === userID);
  if (!mine.length) return 0;
  const days = elapsedWeekdays();
  let missed = 0;
  for (const c of mine) {
    const done = weekHistory[`${c.user_id}:${c.habit_id}`];
    for (const day of days) {
      if (!done?.has(day)) missed++;
    }
  }
  return missed;
}

// Estado agregado de hoy para un usuario: done si cumplió todo, pending si le
// falta algo, missed si ya se le fue el día. null si no tiene hábitos.
export function todayStatus(
  userID: string,
  todayCheckins: CheckinLike[]
): 'done' | 'pending' | 'missed' | null {
  const mine = todayCheckins.filter((c) => c.user_id === userID);
  if (!mine.length) return null;
  if (mine.every((c) => c.status === 'done')) return 'done';
  if (mine.some((c) => c.status === 'pending')) return 'pending';
  return 'missed';
}
