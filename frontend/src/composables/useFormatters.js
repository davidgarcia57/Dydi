// Formateadores de presentación. Estaban en inglés y sin usar por nadie: las
// vistas interpolaban `{{ n }} días` a mano, y por eso una racha de 1 se
// mostraba como "1 días".
export function useFormatters() {
  function formatStreak(n) {
    return n === 1 ? '1 día' : `${n} días`
  }

  function formatPercent(done, total) {
    if (!total) return '0%'
    return `${Math.round((done / total) * 100)}%`
  }

  function formatDate(iso) {
    return new Date(iso).toLocaleDateString('es-MX', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
    })
  }

  return { formatStreak, formatPercent, formatDate }
}
