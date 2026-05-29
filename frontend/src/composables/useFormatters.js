export function useFormatters() {
  function formatStreak(n) {
    return n === 1 ? '1 day' : `${n} days`
  }

  function formatPercent(done, total) {
    if (!total) return '0%'
    return `${Math.round((done / total) * 100)}%`
  }

  function formatDate(iso) {
    return new Date(iso).toLocaleDateString('en-US', {
      weekday: 'short', month: 'short', day: 'numeric',
    })
  }

  return { formatStreak, formatPercent, formatDate }
}
