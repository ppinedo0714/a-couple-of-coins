import { Fragment, memo, useMemo, useState } from 'react'
import { ChevronDown, ChevronRight } from 'lucide-react'
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
import { resolveGroupId } from '@/lib/categories'
import { cn } from '@/lib/utils'

type Props = {
  transactions: Transaction[]
  categories: Category[]
  selectedCategoryId?: string | null
  onSelect?: (categoryId: string | null) => void
}

export const CategoryBreakdown = memo(function CategoryBreakdown({
  transactions,
  categories,
  selectedCategoryId,
  onSelect,
}: Props) {
  const [expandedGroupId, setExpandedGroupId] = useState<string | null>(null)

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
      .sort((a, b) => {
        if (a.id === '__uncategorized') return 1
        if (b.id === '__uncategorized') return -1
        return b.total - a.total
      })
  }, [transactions, catMap])

  const drillRows = useMemo(() => {
    if (!expandedGroupId) return []
    const group = catMap.get(expandedGroupId)
    const txInGroup = transactions.filter(
      (tx) => tx.amount < 0 && resolveGroupId(tx.category_id, catMap) === expandedGroupId,
    )
    const catTotals = new Map<string, { name: string; color: string | null; total: number }>()
    let groupTotal = 0
    for (const tx of txInGroup) {
      const v = Math.abs(tx.amount)
      groupTotal += v
      const isDirectToGroup = tx.category_id === expandedGroupId
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
  }, [transactions, catMap, expandedGroupId])

  const handleGroupClick = (groupId: string) => {
    if (groupId === '__uncategorized') return
    const isCollapsing = expandedGroupId === groupId
    setExpandedGroupId(isCollapsing ? null : groupId)
    onSelect?.(isCollapsing ? null : groupId)
  }

  const handleCategoryClick = (catRow: { id: string }) => {
    const catId = catRow.id === '__general' ? expandedGroupId! : catRow.id
    const active = selectedCategoryId === catId
    onSelect?.(active ? null : catId)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Group breakdown</CardTitle>
      </CardHeader>
      <CardContent className="px-0">
        {overviewRows.length === 0 ? (
          <div className="px-6">
            <EmptyState title="No spending in range" />
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Group</TableHead>
                <TableHead className="text-right">Spent</TableHead>
                <TableHead className="w-32 text-right">Share</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {overviewRows.map((row) => {
                const isExpanded = expandedGroupId === row.id
                const canExpand = row.id !== '__uncategorized'
                const catRows = isExpanded ? drillRows : []
                return (
                  <Fragment key={row.id}>
                    <TableRow
                      onClick={() => handleGroupClick(row.id)}
                      className={cn(canExpand && 'cursor-pointer')}
                    >
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {canExpand ? (
                            isExpanded ? (
                              <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                            ) : (
                              <ChevronRight className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                            )
                          ) : (
                            <span className="h-3.5 w-3.5 shrink-0" />
                          )}
                          <span
                            className="h-2.5 w-2.5 shrink-0 rounded-full"
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
                    {catRows.map((catRow) => {
                      const catId = catRow.id === '__general' ? expandedGroupId! : catRow.id
                      const active = selectedCategoryId === catId
                      return (
                        <TableRow
                          key={catRow.id}
                          onClick={() => handleCategoryClick(catRow)}
                          className={cn('cursor-pointer bg-muted/30', active && 'bg-muted')}
                        >
                          <TableCell>
                            <div className="flex items-center gap-2 pl-8">
                              <span
                                className="h-2 w-2 shrink-0 rounded-full opacity-60"
                                style={{ background: catRow.color ?? 'var(--muted-foreground)' }}
                              />
                              <span className="text-sm text-muted-foreground">{catRow.name}</span>
                            </div>
                          </TableCell>
                          <TableCell className="text-right font-mono text-sm tabular-nums text-muted-foreground">
                            {formatCurrency(catRow.total)}
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="flex items-center justify-end gap-2">
                              <div className="h-1.5 w-16 overflow-hidden rounded-full bg-muted">
                                <div
                                  className="h-full bg-primary opacity-50"
                                  style={{ width: `${Math.min(100, catRow.percent)}%` }}
                                />
                              </div>
                              <span className="w-12 text-right text-xs text-muted-foreground">
                                {formatPercent(catRow.percent, 0)}
                              </span>
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </Fragment>
                )
              })}
            </TableBody>
          </Table>
        )}
        {overviewRows.length > 0 && (
          <p className="px-6 pt-2 text-xs text-muted-foreground">
            Click a group to see category breakdown
          </p>
        )}
      </CardContent>
    </Card>
  )
})
