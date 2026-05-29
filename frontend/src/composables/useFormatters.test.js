import { describe, it, expect } from 'vitest'
import { useFormatters } from './useFormatters'

describe('useFormatters', () => {
  const { formatStreak, formatPercent, formatDate } = useFormatters()

  describe('formatStreak', () => {
    it('singular para 1 día', () => {
      expect(formatStreak(1)).toBe('1 day')
    })
    it('plural para más de 1 día', () => {
      expect(formatStreak(7)).toBe('7 days')
    })
    it('cero días', () => {
      expect(formatStreak(0)).toBe('0 days')
    })
  })

  describe('formatPercent', () => {
    it('calcula porcentaje correctamente', () => {
      expect(formatPercent(3, 4)).toBe('75%')
    })
    it('retorna 0% si total es 0', () => {
      expect(formatPercent(0, 0)).toBe('0%')
    })
    it('redondea correctamente', () => {
      expect(formatPercent(1, 3)).toBe('33%')
    })
  })
})
