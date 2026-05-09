import { CreditCard, Landmark, PiggyBank, TrendingUp } from 'lucide-react'
import type { Account } from '@/types/models'
import { formatCurrency } from '@/lib/format'
import { cn } from '@/lib/utils'

const icons = {
  checking: Landmark,
  savings: PiggyBank,
  credit: CreditCard,
  investment: TrendingUp,
}

const typeLabels = {
  checking: 'Checking',
  savings: 'Savings',
  credit: 'Credit card',
  investment: 'Investment',
}

type Props = {
  account: Account
  variant?: 'card' | 'row'
  onClick?: () => void
  className?: string
}

export function AccountCard({ account, variant = 'card', onClick, className }: Props) {
  const Icon = icons[account.type]
  const balance = formatCurrency(account.balance, account.currency)
  const negative = account.balance < 0

  if (variant === 'row') {
    return (
      <button
        type="button"
        onClick={onClick}
        className={cn(
          'flex w-full items-center justify-between gap-4 rounded-md border border-border bg-card px-4 py-3 text-left transition-colors hover:bg-muted',
          className,
        )}
      >
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-md bg-primary/10 text-primary">
            <Icon className="h-4 w-4" />
          </div>
          <div>
            <div className="text-sm font-medium text-foreground">{account.name}</div>
            <div className="text-xs text-muted-foreground">{typeLabels[account.type]}</div>
          </div>
        </div>
        <div className={cn('font-mono text-sm tabular-nums', negative ? 'text-expense' : 'text-foreground')}>
          {balance}
        </div>
      </button>
    )
  }

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'flex w-56 shrink-0 flex-col gap-3 rounded-lg border border-border bg-card p-4 text-left shadow-sm transition-colors hover:bg-muted/40',
        className,
      )}
    >
      <div className="flex items-center justify-between">
        <div className="flex h-9 w-9 items-center justify-center rounded-md bg-primary/10 text-primary">
          <Icon className="h-4 w-4" />
        </div>
        <span className="text-xs uppercase tracking-wide text-muted-foreground">
          {typeLabels[account.type]}
        </span>
      </div>
      <div className="space-y-1">
        <div className="text-sm text-muted-foreground">{account.name}</div>
        <div
          className={cn(
            'font-mono text-2xl tabular-nums',
            negative ? 'text-expense' : 'text-foreground',
          )}
        >
          {balance}
        </div>
      </div>
    </button>
  )
}
