export type PeriodKey = 'this-month' | 'last-month' | 'last-3-months' | 'this-year' | 'last-year' | 'custom'

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

export function historyStartIso(): string {
  return `${new Date().getFullYear() - 3}-01-01`
}

export function periodToRange(key: PeriodKey, customFrom?: string, customTo?: string): PeriodRange {
  const today = new Date()
  if (key === 'this-month') {
    return {
      key,
      from: `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-01`,
      to: toIsoDate(today),
    }
  }
  if (key === 'last-month') {
    const from = new Date(today.getFullYear(), today.getMonth() - 1, 1)
    return {
      key,
      from: toIsoDate(from),
      to: toIsoDate(today),
    }
  }
  if (key === 'last-3-months') {
    const from = new Date(today.getFullYear(), today.getMonth() - 3, 1)
    return {
      key,
      from: toIsoDate(from),
      to: toIsoDate(today),
    }
  }
  if (key === 'this-year') {
    return {
      key,
      from: `${today.getFullYear()}-01-01`,
      to: toIsoDate(today),
    }
  }
  if (key === 'last-year') {
    const y = today.getFullYear() - 1
    return {
      key,
      from: `${y}-01-01`,
      to: `${y}-12-31`,
    }
  }
  return {
    key: 'custom',
    from: customFrom ?? toIsoDate(today),
    to: customTo ?? toIsoDate(today),
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
    case 'this-year':
      return 'This year'
    case 'last-year':
      return 'Last year'
    case 'custom':
      return 'Custom'
  }
}
