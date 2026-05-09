import { useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Sparkles, Wallet } from 'lucide-react'
import { toast } from 'sonner'
import { useAccounts } from '@/hooks/useAccounts'
import { useCategories } from '@/hooks/useCategories'
import { useTransactions, useClassifyTransactions } from '@/hooks/useTransactions'
import { PageWrapper } from '@/components/layout/PageWrapper'
import { AccountList } from '@/components/accounts/AccountList'
import { TransactionFilters, type TransactionFilterValue } from '@/components/transactions/TransactionFilters'
import { TransactionTable } from '@/components/transactions/TransactionTable'
import { SpendingByCategory } from '@/components/charts/SpendingByCategory'
import { SpendingOverTime } from '@/components/charts/SpendingOverTime'
import { CategoryBreakdown } from '@/components/charts/CategoryBreakdown'
import { LoadingScreen, LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { EmptyState } from '@/components/shared/EmptyState'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { formatShortDate } from '@/lib/format'
import { periodLabel, periodToRange, type PeriodKey } from '@/lib/period'

const PAGE_SIZE = 50

export default function DashboardPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const periodKey = (searchParams.get('period') as PeriodKey | null) ?? 'this-month'
  const accountId = searchParams.get('account_id')
  const categoryId = searchParams.get('category_id')
  const search = searchParams.get('search') ?? ''
  const [page, setPage] = useState(0)
  const [customOpen, setCustomOpen] = useState(false)

  const customFrom = searchParams.get('from') ?? undefined
  const customTo = searchParams.get('to') ?? undefined
  const range = useMemo(
    () => periodToRange(periodKey, customFrom, customTo),
    [periodKey, customFrom, customTo],
  )

  const accountsQuery = useAccounts()
  const categoriesQuery = useCategories()
  const classifyMutation = useClassifyTransactions()
  const allTxQuery = useTransactions({ from: range.from, to: range.to, limit: 200 })
  const filteredTxQuery = useTransactions({
    from: range.from,
    to: range.to,
    account_id: accountId ?? undefined,
    category_id: categoryId ?? undefined,
    search: search || undefined,
    limit: PAGE_SIZE,
    offset: page * PAGE_SIZE,
  })

  const updateParam = (key: string, value: string | null) => {
    const next = new URLSearchParams(searchParams)
    if (value === null || value === '') next.delete(key)
    else next.set(key, value)
    setSearchParams(next, { replace: true })
    setPage(0)
  }

  const filterValue: TransactionFilterValue = {
    accountId: accountId,
    categoryId: categoryId,
    search,
  }

  const onFilterChange = (next: TransactionFilterValue) => {
    const params = new URLSearchParams(searchParams)
    if (next.accountId) params.set('account_id', next.accountId)
    else params.delete('account_id')
    if (next.categoryId) params.set('category_id', next.categoryId)
    else params.delete('category_id')
    if (next.search) params.set('search', next.search)
    else params.delete('search')
    setSearchParams(params, { replace: true })
    setPage(0)
  }

  const onPeriodChange = (next: PeriodKey) => {
    const params = new URLSearchParams(searchParams)
    if (next === 'this-month') params.delete('period')
    else params.set('period', next)
    if (next !== 'custom') {
      params.delete('from')
      params.delete('to')
    }
    setSearchParams(params, { replace: true })
    setPage(0)
  }

  const onCustomDateChange = (key: 'from' | 'to', value: string) => {
    const params = new URLSearchParams(searchParams)
    if (value) params.set(key, value)
    else params.delete(key)
    setSearchParams(params, { replace: true })
    setPage(0)
  }

  if (accountsQuery.isLoading || categoriesQuery.isLoading) {
    return (
      <PageWrapper>
        <LoadingScreen label="Loading your dashboard" />
      </PageWrapper>
    )
  }

  const accounts = accountsQuery.data ?? []
  const categories = categoriesQuery.data ?? []
  const allTransactions = allTxQuery.data?.transactions ?? []
  const filteredTransactions = filteredTxQuery.data?.transactions ?? []
  const hasUncategorized = allTransactions.some((t) => t.category_id === null)

  const onClassify = async () => {
    try {
      const { classified, failed } = await classifyMutation.mutateAsync()
      if (classified === 0) {
        toast('Nothing to classify')
      } else {
        const base = `Classified ${classified} transaction${classified === 1 ? '' : 's'}`
        toast.success(failed > 0 ? `${base} · ${failed} could not be classified` : base)
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not classify')
    }
  }

  if (accounts.length === 0) {
    return (
      <PageWrapper>
        <EmptyState
          icon={Wallet}
          title="Welcome — let's add your first account"
          description="You'll see your charts and recent activity here once you add an account."
          action={<Button onClick={() => navigate('/onboarding')}>Get started</Button>}
        />
      </PageWrapper>
    )
  }

  return (
    <PageWrapper className="space-y-8">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="font-serif text-3xl">Dashboard</h1>
          <p className="text-sm text-muted-foreground">
            {periodLabel(periodKey)} · {range.from} – {range.to}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={periodKey} onValueChange={(v) => onPeriodChange(v as PeriodKey)}>
            <SelectTrigger className="w-44">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="this-month">This month</SelectItem>
              <SelectItem value="last-month">Last month</SelectItem>
              <SelectItem value="last-3-months">Last 3 months</SelectItem>
              <SelectItem value="custom">Custom range</SelectItem>
            </SelectContent>
          </Select>
          {periodKey === 'custom' && (
            <Popover open={customOpen} onOpenChange={setCustomOpen}>
              <PopoverTrigger asChild>
                <Button variant="outline" className="font-normal">
                  {formatShortDate(range.from)} – {formatShortDate(range.to)}
                </Button>
              </PopoverTrigger>
              <PopoverContent align="end" className="w-64 space-y-3">
                <div className="space-y-1.5">
                  <Label htmlFor="custom-from">From</Label>
                  <Input
                    id="custom-from"
                    type="date"
                    value={range.from}
                    onChange={(e) => onCustomDateChange('from', e.target.value)}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="custom-to">To</Label>
                  <Input
                    id="custom-to"
                    type="date"
                    value={range.to}
                    onChange={(e) => onCustomDateChange('to', e.target.value)}
                  />
                </div>
                <div className="flex justify-end">
                  <Button size="sm" onClick={() => setCustomOpen(false)}>
                    Done
                  </Button>
                </div>
              </PopoverContent>
            </Popover>
          )}
        </div>
      </div>

      <section className="space-y-3">
        <h2 className="text-sm font-medium text-muted-foreground">Accounts</h2>
        <AccountList accounts={accounts} />
      </section>

      <section className="grid gap-6 lg:grid-cols-2">
        <SpendingByCategory transactions={allTransactions} categories={categories} />
        <SpendingOverTime transactions={allTransactions} />
      </section>

      <section>
        <CategoryBreakdown
          transactions={allTransactions}
          categories={categories}
          selectedCategoryId={categoryId}
          onSelect={(id) => updateParam('category_id', id)}
        />
      </section>

      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="font-serif text-xl">Transactions</h2>
          {hasUncategorized && (
            <Button
              size="sm"
              variant="outline"
              onClick={onClassify}
              disabled={classifyMutation.isPending}
            >
              {classifyMutation.isPending ? (
                <LoadingSpinner className="mr-2" />
              ) : (
                <Sparkles className="mr-2 h-4 w-4" />
              )}
              Classify uncategorized
            </Button>
          )}
        </div>
        <TransactionFilters
          accounts={accounts}
          categories={categories}
          value={filterValue}
          onChange={onFilterChange}
        />
        <TransactionTable
          transactions={filteredTransactions}
          total={filteredTxQuery.data?.total ?? 0}
          page={page}
          pageSize={PAGE_SIZE}
          onPageChange={setPage}
          accounts={accounts}
          categories={categories}
          loading={filteredTxQuery.isLoading}
        />
      </section>
    </PageWrapper>
  )
}
