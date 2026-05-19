import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Wallet } from 'lucide-react'
import { useAccounts } from '@/hooks/useAccounts'
import { useAccountHistory } from '@/hooks/useAccountHistory'
import { PageWrapper } from '@/components/layout/PageWrapper'
import { AccountList } from '@/components/accounts/AccountList'
import { AccountsOverTime } from '@/components/charts/AccountsOverTime'
import { LoadingScreen } from '@/components/shared/LoadingSpinner'
import { EmptyState } from '@/components/shared/EmptyState'
import { Button } from '@/components/ui/button'
import { formatCurrency } from '@/lib/format'
import { historyStartIso } from '@/lib/period'
import { cn } from '@/lib/utils'
import type { AccountBalanceSnapshot } from '@/types/models'

function todayIso(): string {
  return new Date().toISOString().slice(0, 10)
}

function balanceAtDate(
  accountId: string,
  date: string,
  snapshots: AccountBalanceSnapshot[],
): number | null {
  const relevant = snapshots
    .filter((s) => s.account_id === accountId && s.date <= date)
    .sort((a, b) => a.date.localeCompare(b.date))
  return relevant.length > 0 ? relevant[relevant.length - 1]!.balance : null
}

export default function AccountsPage() {
  const navigate = useNavigate()
  const today = todayIso()
  const [selectedDate, setSelectedDate] = useState(today)

  const historyStart = historyStartIso()
  const accountsQuery = useAccounts()
  const historyQuery = useAccountHistory({ from: historyStart, to: today, interval: 'week' })

  if (accountsQuery.isLoading) {
    return (
      <PageWrapper>
        <LoadingScreen label="Loading accounts" />
      </PageWrapper>
    )
  }

  const accounts = accountsQuery.data ?? []

  if (accounts.length === 0) {
    return (
      <PageWrapper>
        <EmptyState
          icon={Wallet}
          title="Welcome — let's add your first account"
          description="You'll see your balances and recent activity here once you add an account."
          action={<Button onClick={() => navigate('/onboarding')}>Get started</Button>}
        />
      </PageWrapper>
    )
  }

  const snapshots = historyQuery.data?.snapshots ?? []
  const isHistorical = selectedDate !== today

  // Override each account's balance with the snapshot value for the selected date
  const accountsAtDate = accounts.map((acc) => {
    const historical = balanceAtDate(acc.id, selectedDate, snapshots)
    return historical !== null ? { ...acc, balance: historical } : acc
  })

  const netWorth = accountsAtDate.reduce((sum, acc) => sum + acc.balance, 0)

  return (
    <PageWrapper className="space-y-8">
      {/* Top section: accounts with date picker */}
      <div className="space-y-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="font-serif text-3xl">Accounts</h1>
            <p className="text-sm text-muted-foreground">
              {isHistorical ? 'Historical balances' : 'Your current balances'}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <label
              htmlFor="balance-date"
              className="text-sm text-muted-foreground whitespace-nowrap"
            >
              Balances as of
            </label>
            <input
              id="balance-date"
              type="date"
              value={selectedDate}
              max={today}
              min={historyStart}
              onChange={(e) => setSelectedDate(e.target.value || today)}
              className="h-8 rounded-md border border-input bg-background px-3 text-sm focus:outline-none focus:ring-1 focus:ring-ring"
            />
            {isHistorical && (
              <Button
                variant="ghost"
                size="sm"
                className="h-8 text-xs"
                onClick={() => setSelectedDate(today)}
              >
                Today
              </Button>
            )}
          </div>
        </div>

        {/* Pass date-adjusted balances to AccountList */}
        <AccountList accounts={accountsAtDate} />

        <div className="flex items-center justify-between rounded-lg border border-border bg-muted/40 px-4 py-3">
          <span className="text-sm font-medium text-muted-foreground">Net worth</span>
          <span
            className={cn(
              'font-mono text-lg font-semibold tabular-nums',
              netWorth >= 0 ? 'text-income' : 'text-expense',
            )}
          >
            {formatCurrency(netWorth)}
          </span>
        </div>
      </div>

      {/* Bottom section: balance history chart */}
      <AccountsOverTime accounts={accounts} snapshots={snapshots} />
    </PageWrapper>
  )
}
