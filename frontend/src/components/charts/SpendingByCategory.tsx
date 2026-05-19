import { memo, useMemo, useState } from 'react'
import { ChevronLeft } from 'lucide-react'
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts'
import type { Category, Transaction } from '@/types/models'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { EmptyState } from '@/components/shared/EmptyState'
import { formatCurrency } from '@/lib/format'
import { resolveGroupId } from '@/lib/categories'

const FALLBACK_COLORS = [
  'var(--primary)',
  'var(--accent)',
  'var(--expense)',
  'oklch(0.65 0.12 140)',
  'oklch(0.60 0.13 280)',
  'oklch(0.70 0.12 50)',
  'oklch(0.55 0.10 320)',
]

type Props = {
  transactions: Transaction[]
  categories: Category[]
}

export const SpendingByCategory = memo(function SpendingByCategory({ transactions, categories }: Props) {
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null)

  const catMap = useMemo(() => new Map(categories.map((c) => [c.id, c])), [categories])

  const overviewData = useMemo(() => {
    const groupMap = new Map<string, { name: string; color: string | null; total: number }>()
    for (const tx of transactions) {
      if (tx.amount >= 0) continue
      const groupId = resolveGroupId(tx.category_id, catMap)
      const group = catMap.get(groupId)
      const entry = groupMap.get(groupId) ?? {
        name: group?.name ?? 'Uncategorized',
        color: group?.color ?? null,
        total: 0,
      }
      entry.total += Math.abs(tx.amount)
      groupMap.set(groupId, entry)
    }
    return Array.from(groupMap.entries())
      .sort(([, a], [, b]) => b.total - a.total)
      .map(([id, d], i) => ({
        id,
        ...d,
        fill: d.color ?? FALLBACK_COLORS[i % FALLBACK_COLORS.length]!,
      }))
  }, [transactions, catMap])

  const drillData = useMemo(() => {
    if (!selectedGroupId) return []
    const group = catMap.get(selectedGroupId)
    const txInGroup = transactions.filter(
      (tx) => tx.amount < 0 && resolveGroupId(tx.category_id, catMap) === selectedGroupId,
    )
    const byId = new Map<string, { name: string; color: string | null; total: number }>()
    for (const tx of txInGroup) {
      const isDirectToGroup = tx.category_id === selectedGroupId
      const id = isDirectToGroup ? '__general' : (tx.category_id ?? '__general')
      const cat = isDirectToGroup ? null : (tx.category_id ? catMap.get(tx.category_id) : null)
      const name = isDirectToGroup
        ? `${group?.name ?? 'General'} (general)`
        : (cat?.name ?? 'General')
      const entry = byId.get(id) ?? { name, color: group?.color ?? null, total: 0 }
      entry.total += Math.abs(tx.amount)
      byId.set(id, entry)
    }
    return Array.from(byId.values())
      .sort((a, b) => b.total - a.total)
      .map((d, i) => ({
        ...d,
        fill: d.color ?? FALLBACK_COLORS[i % FALLBACK_COLORS.length]!,
      }))
  }, [transactions, catMap, selectedGroupId])

  const currentData = selectedGroupId ? drillData : overviewData
  const total = currentData.reduce((sum, d) => sum + d.total, 0)
  const selectedGroupName = selectedGroupId
    ? (overviewData.find((d) => d.id === selectedGroupId)?.name ?? 'Group')
    : null

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-1">
          {selectedGroupId && (
            <Button
              size="sm"
              variant="ghost"
              className="h-7 px-1.5"
              onClick={() => setSelectedGroupId(null)}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
          )}
          <CardTitle className="text-base">
            {selectedGroupId ? `${selectedGroupName} breakdown` : 'Spending distribution'}
          </CardTitle>
        </div>
        <CardDescription>Total: {formatCurrency(total)}</CardDescription>
      </CardHeader>
      <CardContent>
        {currentData.length === 0 ? (
          <EmptyState title="No spending in range" description="Try a wider date range." />
        ) : (
          <div className="h-80 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={currentData}
                  dataKey="total"
                  nameKey="name"
                  innerRadius={75}
                  outerRadius={120}
                  paddingAngle={2}
                  stroke="var(--background)"
                  onClick={
                    selectedGroupId
                      ? undefined
                      : (data: { id?: string } | null) => {
                          if (data?.id && data.id !== '__uncategorized') {
                            setSelectedGroupId(data.id)
                          }
                        }
                  }
                  className={selectedGroupId ? undefined : 'cursor-pointer'}
                >
                  {currentData.map((entry, idx) => (
                    <Cell key={idx} fill={entry.fill} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    background: 'var(--popover)',
                    border: '1px solid var(--border)',
                    borderRadius: 'var(--radius)',
                    color: 'var(--popover-foreground)',
                    fontSize: '0.8rem',
                  }}
                  formatter={(value: number) => formatCurrency(value)}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>
        )}
        {!selectedGroupId && currentData.length > 0 && (
          <p className="mt-1 text-center text-xs text-muted-foreground">
            Click a segment to see category breakdown
          </p>
        )}
      </CardContent>
    </Card>
  )
})
