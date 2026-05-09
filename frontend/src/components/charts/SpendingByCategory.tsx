import { useMemo } from 'react'
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts'
import type { Category, Transaction } from '@/types/models'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { EmptyState } from '@/components/shared/EmptyState'
import { formatCurrency } from '@/lib/format'

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

export function SpendingByCategory({ transactions, categories }: Props) {
  const data = useMemo(() => {
    const byId = new Map<string, { name: string; color: string | null; total: number }>()
    for (const tx of transactions) {
      if (tx.amount >= 0) continue
      const cat = categories.find((c) => c.id === tx.category_id)
      const id = cat?.id ?? '__uncategorized'
      const name = cat?.name ?? 'Uncategorized'
      const color = cat?.color ?? null
      const entry = byId.get(id) ?? { name, color, total: 0 }
      entry.total += Math.abs(tx.amount)
      byId.set(id, entry)
    }
    return Array.from(byId.values())
      .sort((a, b) => b.total - a.total)
      .map((d, i) => ({
        ...d,
        fill: d.color ?? FALLBACK_COLORS[i % FALLBACK_COLORS.length]!,
      }))
  }, [transactions, categories])

  const total = data.reduce((sum, d) => sum + d.total, 0)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Spending by category</CardTitle>
        <CardDescription>Total: {formatCurrency(total)}</CardDescription>
      </CardHeader>
      <CardContent>
        {data.length === 0 ? (
          <EmptyState title="No spending in range" description="Try a wider date range." />
        ) : (
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={data}
                  dataKey="total"
                  nameKey="name"
                  innerRadius={60}
                  outerRadius={90}
                  paddingAngle={2}
                  stroke="var(--background)"
                >
                  {data.map((entry, idx) => (
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
      </CardContent>
    </Card>
  )
}
