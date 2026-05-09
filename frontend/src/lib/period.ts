export type PeriodKey = 'this-month' | 'last-month' | 'last-3-months' | 'custom'

export type PeriodRange = {
  key: PeriodKey
  from: string
  to: string
}

function toIsoDate(d: Date): string {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function startOfMonth(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), 1)
}

function endOfMonth(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth() + 1, 0)
}

export function periodToRange(key: PeriodKey, customFrom?: string, customTo?: string): PeriodRange {
  const today = new Date()
  if (key === 'this-month') {
    return {
      key,
      from: toIsoDate(startOfMonth(today)),
      to: toIsoDate(endOfMonth(today)),
    }
  }
  if (key === 'last-month') {
    const lastMonth = new Date(today.getFullYear(), today.getMonth() - 1, 1)
    return {
      key,
      from: toIsoDate(startOfMonth(lastMonth)),
      to: toIsoDate(endOfMonth(lastMonth)),
    }
  }
  if (key === 'last-3-months') {
    const start = new Date(today.getFullYear(), today.getMonth() - 2, 1)
    return {
      key,
      from: toIsoDate(startOfMonth(start)),
      to: toIsoDate(endOfMonth(today)),
    }
  }
  return {
    key: 'custom',
    from: customFrom ?? toIsoDate(startOfMonth(today)),
    to: customTo ?? toIsoDate(endOfMonth(today)),
  }
}

export function periodLabel(key: PeriodKey): string {
  switch (key) {
    case 'this-month':
      return 'This month'
    case 'last-month':
      return 'Last month'
    case 'last-3-months':
      return 'Last 3 months'
    case 'custom':
      return 'Custom'
  }
}
