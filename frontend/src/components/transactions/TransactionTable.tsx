import { memo, useMemo, useState } from 'react'
import { ArrowDown, ArrowUp, ArrowUpDown, Inbox } from 'lucide-react'
import { toast } from 'sonner'
import type { Account, Category, Transaction } from '@/types/models'
import { useUpdateTransaction } from '@/hooks/useTransactions'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { EmptyState } from '@/components/shared/EmptyState'
import { formatDate, formatSignedCurrency } from '@/lib/format'
import { cn } from '@/lib/utils'

type SortKey = 'date' | 'description' | 'amount'
type SortDir = 'asc' | 'desc'

type SortHeaderProps = {
  label: string
  sortKey: SortKey
  currentKey: SortKey
  currentDir: SortDir
  onSort: (next: SortKey) => void
  align?: 'right'
}

function SortHeader({ label, sortKey, currentKey, currentDir, onSort, align }: SortHeaderProps) {
  const active = currentKey === sortKey
  const Icon = !active ? ArrowUpDown : currentDir === 'asc' ? ArrowUp : ArrowDown
  return (
    <button
      type="button"
      onClick={() => onSort(sortKey)}
      className={cn(
        'inline-flex items-center gap-1 text-xs font-medium uppercase tracking-wide text-muted-foreground hover:text-foreground',
        align === 'right' && 'ml-auto',
      )}
    >
      {label}
      <Icon className="h-3 w-3" />
    </button>
  )
}

type Props = {
  transactions: Transaction[]
  total: number
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  accounts: Account[]
  categories: Category[]
  loading?: boolean
}

export const TransactionTable = memo(function TransactionTable({
  transactions,
  total,
  page,
  pageSize,
  onPageChange,
  accounts,
  categories,
  loading,
}: Props) {
  const [sortKey, setSortKey] = useState<SortKey>('date')
  const [sortDir, setSortDir] = useState<SortDir>('desc')
  const updateMut = useUpdateTransaction()

  const sorted = useMemo(() => {
    const list = [...transactions]
    list.sort((a, b) => {
      let cmp: number
      if (sortKey === 'date') cmp = a.date < b.date ? -1 : a.date > b.date ? 1 : 0
      else if (sortKey === 'description') cmp = a.description.localeCompare(b.description)
      else cmp = a.amount - b.amount
      return sortDir === 'asc' ? cmp : -cmp
    })
    return list
  }, [transactions, sortKey, sortDir])

  const handleSort = (next: SortKey) => {
    if (next === sortKey) {
      setSortDir(sortDir === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(next)
      setSortDir('desc')
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const accountName = (id: string) => accounts.find((a) => a.id === id)?.name ?? '—'

  const groups = useMemo(() => categories.filter((c) => c.parent_id === null), [categories])
  const categoriesByGroupId = useMemo(() => {
    const map: Record<string, Category[]> = {}
    for (const c of categories) {
      if (c.parent_id !== null) {
        if (!map[c.parent_id]) map[c.parent_id] = []
        map[c.parent_id]!.push(c)
      }
    }
    return map
  }, [categories])

  const onCategoryChange = async (tx: Transaction, categoryId: string) => {
    const next = categoryId === 'none' ? null : categoryId
    try {
      await updateMut.mutateAsync({ id: tx.id, body: { category_id: next } })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not update category')
    }
  }

  if (!loading && transactions.length === 0) {
    return (
      <EmptyState
        icon={Inbox}
        title="No transactions"
        description="Adjust your filters or import a CSV to get started."
      />
    )
  }

  return (
    <div className="rounded-lg border border-border bg-card">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>
              <SortHeader
                label="Date"
                sortKey="date"
                currentKey={sortKey}
                currentDir={sortDir}
                onSort={handleSort}
              />
            </TableHead>
            <TableHead>
              <SortHeader
                label="Description"
                sortKey="description"
                currentKey={sortKey}
                currentDir={sortDir}
                onSort={handleSort}
              />
            </TableHead>
            <TableHead className="hidden sm:table-cell">Merchant</TableHead>
            <TableHead className="hidden md:table-cell">Account</TableHead>
            <TableHead>Category</TableHead>
            <TableHead className="text-right">
              <SortHeader
                label="Amount"
                sortKey="amount"
                currentKey={sortKey}
                currentDir={sortDir}
                onSort={handleSort}
                align="right"
              />
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sorted.map((tx) => (
            <TableRow key={tx.id}>
              <TableCell className="whitespace-nowrap text-sm text-muted-foreground">
                {formatDate(tx.date)}
              </TableCell>
              <TableCell className="text-sm font-medium">{tx.description}</TableCell>
              <TableCell className="hidden text-sm text-muted-foreground sm:table-cell">
                {tx.merchant_name ?? '—'}
              </TableCell>
              <TableCell className="hidden text-sm text-muted-foreground md:table-cell">
                {accountName(tx.account_id)}
              </TableCell>
              <TableCell>
                <Select
                  value={tx.category_id ?? 'none'}
                  onValueChange={(v) => void onCategoryChange(tx, v)}
                >
                  <SelectTrigger className="h-8 w-44 text-xs">
                    <SelectValue placeholder="Uncategorized" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">Uncategorized</SelectItem>
                    {groups.map((group) => {
                      const children = categoriesByGroupId[group.id] ?? []
                      return (
                        <SelectGroup key={group.id}>
                          <SelectLabel className="flex items-center gap-1.5 text-xs">
                            <span
                              className="h-2 w-2 rounded-full"
                              style={{ background: group.color ?? 'var(--muted-foreground)' }}
                            />
                            {group.name}
                          </SelectLabel>
                          <SelectItem value={group.id}>
                            {group.name} (general)
                          </SelectItem>
                          {children.map((c) => (
                            <SelectItem key={c.id} value={c.id}>
                              {c.name}
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      )
                    })}
                  </SelectContent>
                </Select>
                {!tx.classified ? (
                  <Badge variant="outline" className="ml-2 align-middle">
                    new
                  </Badge>
                ) : null}
              </TableCell>
              <TableCell
                className={cn(
                  'text-right font-mono tabular-nums',
                  tx.amount >= 0 ? 'text-income' : 'text-expense',
                )}
              >
                {formatSignedCurrency(tx.amount)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <div className="flex items-center justify-between border-t border-border px-4 py-3 text-xs text-muted-foreground">
        <span>
          {total === 0
            ? '0 transactions'
            : `Showing ${page * pageSize + 1}–${Math.min((page + 1) * pageSize, total)} of ${total}`}
        </span>
        <div className="flex gap-1">
          <Button
            variant="outline"
            size="sm"
            disabled={page === 0}
            onClick={() => onPageChange(Math.max(0, page - 1))}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages - 1}
            onClick={() => onPageChange(Math.min(totalPages - 1, page + 1))}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  )
})
