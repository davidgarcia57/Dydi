// Formateadores de presentacion del movil. Espejo de
// frontend/src/composables/useFormatters.js: si cambia la copia en un lado,
// cambiala en el otro.
//
// Existe porque las pantallas interpolaban `{prevStreak} dias` a mano y una
// racha de 1 se mostraba como "1 dias".
export function formatStreak(n: number): string {
  return n === 1 ? '1 día' : `${n} días`;
}

export function formatPercent(done: number, total: number): string {
  if (!total) return '0%';
  return `${Math.round((done / total) * 100)}%`;
}
