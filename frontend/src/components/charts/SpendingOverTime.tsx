import { memo, useMemo } from 'react'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import type { Transaction } from '@/types/models'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { EmptyState } from '@/components/shared/EmptyState'
import { formatCurrency, formatShortDate } from '@/lib/format'

type Props = {
  transactions: Transaction[]
  dateRange?: { from: string; to: string }
}

export const SpendingOverTime = memo(function SpendingOverTime({ transactions, dateRange }: Props) {
  const chartData = useMemo(() => {
    const byDay = new Map<string, { date: string; income: number; expense: number }>()
    for (const tx of transactions) {
      const entry = byDay.get(tx.date) ?? { date: tx.date, income: 0, expense: 0 }
      if (tx.amount >= 0) entry.income += tx.amount
      else entry.expense += Math.abs(tx.amount)
      byDay.set(tx.date, entry)
    }
    if (dateRange) {
      const cursor = new Date(dateRange.from + 'T12:00:00')
      const end = new Date(dateRange.to + 'T12:00:00')
      while (cursor <= end) {
        const y = cursor.getFullYear()
        const m = String(cursor.getMonth() + 1).padStart(2, '0')
        const d = String(cursor.getDate()).padStart(2, '0')
        const dateStr = `${y}-${m}-${d}`
        if (!byDay.has(dateStr)) byDay.set(dateStr, { date: dateStr, income: 0, expense: 0 })
        cursor.setDate(cursor.getDate() + 1)
      }
    }
    const sorted = Array.from(byDay.values()).sort((a, b) => (a.date < b.date ? -1 : 1))
    const peak = sorted.reduce((m, e) => Math.max(m, e.income, e.expense), 0)
    const yMax = Math.ceil(peak / 1000) * 1000 || 1000
    return { sorted, yMax }
  }, [transactions, dateRange])

  const { sorted: data, yMax } = chartData

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Cash flow</CardTitle>
        <CardDescription>Daily income vs. expense.</CardDescription>
      </CardHeader>
      <CardContent>
        {data.length === 0 ? (
          <EmptyState title="No activity in range" description="Try a wider date range." />
        ) : (
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={data} margin={{ top: 10, right: 10, left: 0, bottom: 0 }} barSize={8}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
                <XAxis
                  dataKey="date"
                  tickFormatter={(v) => formatShortDate(v)}
                  stroke="var(--muted-foreground)"
                  fontSize={11}
                  tickLine={false}
                  axisLine={false}
                  minTickGap={20}
                />
                <YAxis
                  stroke="var(--muted-foreground)"
                  fontSize={11}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={(v) => `$${Math.round(Number(v))}`}
                  domain={[0, yMax]}
                />
                <Tooltip
                  contentStyle={{
                    background: 'var(--popover)',
                    border: '1px solid var(--border)',
                    borderRadius: 'var(--radius)',
                    color: 'var(--popover-foreground)',
                    fontSize: '0.8rem',
                  }}
                  labelFormatter={(label) => formatShortDate(String(label))}
                  formatter={(value: number, name) => [formatCurrency(value), name]}
                />
                <Legend wrapperStyle={{ fontSize: '0.8rem' }} />
                <Bar dataKey="income" name="Income" fill="var(--income)" radius={[2, 2, 0, 0]} />
                <Bar dataKey="expense" name="Expense" fill="var(--expense)" radius={[2, 2, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </CardContent>
    </Card>
  )
})
