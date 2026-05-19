import { useMemo, useState } from 'react'
import { keepPreviousData } from '@tanstack/react-query'
import { Sparkles } from 'lucide-react'
import { toast } from 'sonner'
import { useAccounts } from '@/hooks/useAccounts'
import { useCategories } from '@/hooks/useCategories'
import { useTransactions, useClassifyTransactions } from '@/hooks/useTransactions'
import { PageWrapper } from '@/components/layout/PageWrapper'
import { TransactionFilters, type TransactionFilterValue } from '@/components/transactions/TransactionFilters'
import { TransactionTable } from '@/components/transactions/TransactionTable'
import { SpendingByCategory } from '@/components/charts/SpendingByCategory'
import { SpendingOverTime } from '@/components/charts/SpendingOverTime'
import { CategoryBreakdown } from '@/components/charts/CategoryBreakdown'
import { LoadingScreen, LoadingSpinner } from '@/components/shared/LoadingSpinner'
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

export default function TransactionsPage() {
  const [periodKey, setPeriodKey] = useState<PeriodKey>('this-month')
  const [accountId, setAccountId] = useState<string | null>(null)
  const [categoryId, setCategoryId] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [customFrom, setCustomFrom] = useState<string | undefined>(undefined)
  const [customTo, setCustomTo] = useState<string | undefined>(undefined)
  const [page, setPage] = useState(0)
  const [customOpen, setCustomOpen] = useState(false)

  const range = useMemo(
    () => periodToRange(periodKey, customFrom, customTo),
    [periodKey, customFrom, customTo],
  )

  const accountsQuery = useAccounts()
  const categoriesQuery = useCategories()
  const classifyMutation = useClassifyTransactions()

  const chartQuery = useTransactions(
    { from: range.from, to: range.to, limit: 2000 },
    { placeholderData: keepPreviousData },
  )

  const tableQuery = useTransactions(
    {
      from: range.from,
      to: range.to,
      account_id: accountId ?? undefined,
      category_id: categoryId ?? undefined,
      search: search || undefined,
      limit: PAGE_SIZE,
      offset: page * PAGE_SIZE,
    },
    { placeholderData: keepPreviousData },
  )

  const categories = categoriesQuery.data ?? []
  const filterValue: TransactionFilterValue = { accountId, categoryId, search }

  const onFilterChange = (next: TransactionFilterValue) => {
    setAccountId(next.accountId ?? null)
    setCategoryId(next.categoryId ?? null)
    setSearch(next.search ?? '')
    setPage(0)
  }

  const onPeriodChange = (next: PeriodKey) => {
    setPeriodKey(next)
    if (next !== 'custom') {
      setCustomFrom(undefined)
      setCustomTo(undefined)
    }
    setPage(0)
  }

  const onCustomDateChange = (key: 'from' | 'to', value: string) => {
    if (key === 'from') setCustomFrom(value || undefined)
    else setCustomTo(value || undefined)
    setPage(0)
  }

  if (accountsQuery.isLoading || categoriesQuery.isLoading) {
    return (
      <PageWrapper>
        <LoadingScreen label="Loading transactions" />
      </PageWrapper>
    )
  }

  const accounts = accountsQuery.data ?? []
  const chartTransactions = chartQuery.data?.transactions ?? []
  const hasUncategorized = chartTransactions.some((t) => t.category_id === null)

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

  return (
    <PageWrapper className="space-y-8">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="font-serif text-3xl">Transactions</h1>
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
              <SelectItem value="this-year">This year</SelectItem>
              <SelectItem value="last-year">Last year</SelectItem>
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

      <section>
        <SpendingOverTime transactions={chartTransactions} dateRange={range} />
      </section>

      <section className="grid gap-6 lg:grid-cols-2">
        <SpendingByCategory transactions={chartTransactions} categories={categories} />
        <CategoryBreakdown
          transactions={chartTransactions}
          categories={categories}
          selectedCategoryId={categoryId}
          onSelect={(id) => { setCategoryId(id); setPage(0) }}
        />
      </section>

      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="font-serif text-xl">All transactions</h2>
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
          transactions={tableQuery.data?.transactions ?? []}
          total={tableQuery.data?.total ?? 0}
          page={page}
          pageSize={PAGE_SIZE}
          onPageChange={setPage}
          accounts={accounts}
          categories={categories}
          loading={tableQuery.isLoading}
        />
      </section>
    </PageWrapper>
  )
}
