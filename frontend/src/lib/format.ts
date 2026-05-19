export function formatCurrency(amount: number, currency = 'USD'): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    maximumFractionDigits: 2,
  }).format(amount)
}

export function formatSignedCurrency(amount: number, currency = 'USD'): string {
  const formatted = formatCurrency(Math.abs(amount), currency)
  if (amount > 0) return `+${formatted}`
  if (amount < 0) return `-${formatted}`
  return formatted
}

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
})

const shortDateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
})

function parseDate(date: string | Date): Date {
  if (typeof date !== 'string') return date
  const [year, month, day] = date.split('-').map(Number)
  return new Date(year!, month! - 1, day!)
}

export function formatDate(date: string | Date): string {
  return dateFormatter.format(parseDate(date))
}

export function formatShortDate(date: string | Date): string {
  return shortDateFormatter.format(parseDate(date))
}

export function formatPercent(value: number, fractionDigits = 1): string {
  return `${value.toFixed(fractionDigits)}%`
}

export function getInitials(email: string): string {
  const local = email.split('@')[0] ?? email
  return local.slice(0, 2).toUpperCase()
}
