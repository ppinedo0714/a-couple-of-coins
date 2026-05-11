import { useMemo, useState } from 'react'
import { ChevronLeft } from 'lucide-react'
import type { Category, Transaction } from '@/types/models'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
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

function resolveGroupId(categoryId: string | null, catMap: Map<string, Category>): string {
  if (!categoryId) return '__uncategorized'
  const cat = catMap.get(categoryId)
  if (!cat) return '__uncategorized'
  return cat.parent_id ?? cat.id
}

export function CategoryBreakdown({
  transactions,
  categories,
  selectedCategoryId,
  onSelect,
}: Props) {
  const [drillGroupId, setDrillGroupId] = useState<string | null>(null)

  const catMap = useMemo(() => new Map(categories.map((c) => [c.id, c])), [categories])

  const overviewRows = useMemo(() => {
    const groupTotals = new Map<string, { name: string; color: string | null; total: number }>()
    let grandTotal = 0
    for (const tx of transactions) {
      if (tx.amount >= 0) continue
      const v = Math.abs(tx.amount)
      grandTotal += v
      const groupId = resolveGroupId(tx.category_id, catMap)
      const group = catMap.get(groupId)
      const entry = groupTotals.get(groupId) ?? {
        name: group?.name ?? 'Uncategorized',
        color: group?.color ?? null,
        total: 0,
      }
      entry.total += v
      groupTotals.set(groupId, entry)
    }
    return Array.from(groupTotals.entries())
      .map(([id, d]) => ({ id, ...d, percent: grandTotal === 0 ? 0 : (d.total / grandTotal) * 100 }))
      .sort((a, b) => b.total - a.total)
  }, [transactions, catMap])

  const drillRows = useMemo(() => {
    if (!drillGroupId) return []
    const group = catMap.get(drillGroupId)
    const txInGroup = transactions.filter(
      (tx) => tx.amount < 0 && resolveGroupId(tx.category_id, catMap) === drillGroupId,
    )
    const catTotals = new Map<string, { name: string; color: string | null; total: number }>()
    let groupTotal = 0
    for (const tx of txInGroup) {
      const v = Math.abs(tx.amount)
      groupTotal += v
      const isDirectToGroup = tx.category_id === drillGroupId
      const id = isDirectToGroup ? '__general' : (tx.category_id ?? '__general')
      const cat = isDirectToGroup ? null : (tx.category_id ? catMap.get(tx.category_id) : null)
      const name = isDirectToGroup
        ? `${group?.name ?? 'General'} (general)`
        : (cat?.name ?? 'General')
      const entry = catTotals.get(id) ?? { name, color: group?.color ?? null, total: 0 }
      entry.total += v
      catTotals.set(id, entry)
    }
    return Array.from(catTotals.entries())
      .map(([id, d]) => ({ id, ...d, percent: groupTotal === 0 ? 0 : (d.total / groupTotal) * 100 }))
      .sort((a, b) => b.total - a.total)
  }, [transactions, catMap, drillGroupId])

  const rows = drillGroupId ? drillRows : overviewRows
  const drillGroupName = drillGroupId
    ? (overviewRows.find((r) => r.id === drillGroupId)?.name ?? 'Group')
    : null

  const handleRowClick = (row: { id: string }) => {
    if (!drillGroupId) {
      if (row.id !== '__uncategorized') {
        setDrillGroupId(row.id)
        onSelect?.(null)
      }
    } else {
      const catId = row.id === '__general' ? drillGroupId : row.id
      const active = selectedCategoryId === catId
      onSelect?.(active ? null : catId)
    }
  }

  const handleBack = () => {
    setDrillGroupId(null)
    onSelect?.(null)
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-1">
          {drillGroupId && (
            <Button size="sm" variant="ghost" className="h-7 px-1.5" onClick={handleBack}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
          )}
          <CardTitle className="text-base">
            {drillGroupId ? `${drillGroupName} breakdown` : 'Spending by group'}
          </CardTitle>
        </div>
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
                <TableHead>{drillGroupId ? 'Category' : 'Group'}</TableHead>
                <TableHead className="text-right">Spent</TableHead>
                <TableHead className="w-32 text-right">Share</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((row) => {
                const catId = row.id === '__general' ? drillGroupId : row.id
                const active = drillGroupId ? selectedCategoryId === catId : false
                return (
                  <TableRow
                    key={String(row.id)}
                    onClick={() => handleRowClick(row)}
                    className={cn('cursor-pointer', active && 'bg-muted')}
                  >
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <span
                          className="h-2.5 w-2.5 rounded-full"
                          style={{ background: row.color ?? 'var(--muted-foreground)' }}
                        />
                        <span className="text-sm">{row.name}</span>
                        {!drillGroupId && row.id !== '__uncategorized' && (
                          <span className="text-xs text-muted-foreground">↗</span>
                        )}
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
        {!drillGroupId && rows.length > 0 && (
          <p className="px-6 pt-2 text-xs text-muted-foreground">
            Click a group to see category breakdown
          </p>
        )}
      </CardContent>
    </Card>
  )
}
