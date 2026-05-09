import { useMemo } from 'react'
import type { Category, Transaction } from '@/types/models'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { EmptyState } from '@/components/shared/EmptyState'
import { formatCurrency, formatPercent } from '@/lib/format'
import { cn } from '@/lib/utils'

type Props = {
  transactions: Transaction[]
  categories: Category[]
  selectedCategoryId?: string | null
  onSelect?: (categoryId: string | null) => void
}

export function CategoryBreakdown({
  transactions,
  categories,
  selectedCategoryId,
  onSelect,
}: Props) {
  const rows = useMemo(() => {
    const totals = new Map<string | null, number>()
    let grandTotal = 0
    for (const tx of transactions) {
      if (tx.amount >= 0) continue
      const v = Math.abs(tx.amount)
      grandTotal += v
      totals.set(tx.category_id, (totals.get(tx.category_id) ?? 0) + v)
    }
    const out: Array<{
      id: string | null
      name: string
      color: string | null
      total: number
      percent: number
    }> = []
    for (const [id, total] of totals) {
      const cat = categories.find((c) => c.id === id)
      out.push({
        id,
        name: cat?.name ?? 'Uncategorized',
        color: cat?.color ?? null,
        total,
        percent: grandTotal === 0 ? 0 : (total / grandTotal) * 100,
      })
    }
    return out.sort((a, b) => b.total - a.total)
  }, [transactions, categories])

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Category breakdown</CardTitle>
      </CardHeader>
      <CardContent className="px-0">
        {rows.length === 0 ? (
          <div className="px-6">
            <EmptyState title="No spending in range" />
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Category</TableHead>
                <TableHead className="text-right">Spent</TableHead>
                <TableHead className="w-32 text-right">Share</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((row) => {
                const active = selectedCategoryId === row.id
                return (
                  <TableRow
                    key={String(row.id)}
                    onClick={() => onSelect?.(active ? null : row.id)}
                    className={cn('cursor-pointer', active && 'bg-muted')}
                  >
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <span
                          className="h-2.5 w-2.5 rounded-full"
                          style={{ background: row.color ?? 'var(--muted-foreground)' }}
                        />
                        <span className="text-sm">{row.name}</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-right font-mono text-sm tabular-nums">
                      {formatCurrency(row.total)}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-2">
                        <div className="h-1.5 w-16 overflow-hidden rounded-full bg-muted">
                          <div
                            className="h-full bg-primary"
                            style={{ width: `${Math.min(100, row.percent)}%` }}
                          />
                        </div>
                        <span className="w-12 text-right text-xs text-muted-foreground">
                          {formatPercent(row.percent, 0)}
                        </span>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}
