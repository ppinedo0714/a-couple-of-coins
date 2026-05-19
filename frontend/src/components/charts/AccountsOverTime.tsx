import { useMemo, useState } from 'react'
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { TrendingUp } from 'lucide-react'
import type { Account, AccountBalanceSnapshot } from '@/types/models'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { EmptyState } from '@/components/shared/EmptyState'
import { formatCurrency, formatShortDate } from '@/lib/format'
import { historyStartIso } from '@/lib/period'

type TimeRange = '1M' | '3M' | '6M' | '1Y' | 'ALL'

const TIME_RANGES: { label: string; value: TimeRange }[] = [
  { label: '1M', value: '1M' },
  { label: '3M', value: '3M' },
  { label: '6M', value: '6M' },
  { label: '1Y', value: '1Y' },
  { label: 'All', value: 'ALL' },
]

const LINE_COLORS = ['var(--primary)', 'var(--accent)', 'var(--muted-foreground)']
const NET_WORTH_KEY = '__net_worth__'

function getFromDate(range: TimeRange): string {
  if (range === 'ALL') return historyStartIso()
  const d = new Date()
  if (range === '1M') d.setMonth(d.getMonth() - 1)
  else if (range === '3M') d.setMonth(d.getMonth() - 3)
  else if (range === '6M') d.setMonth(d.getMonth() - 6)
  else if (range === '1Y') d.setFullYear(d.getFullYear() - 1)
  return d.toISOString().slice(0, 10)
}

type Props = {
  accounts: Account[]
  snapshots: AccountBalanceSnapshot[]
}

export function AccountsOverTime({ accounts, snapshots }: Props) {
  const [timeRange, setTimeRange] = useState<TimeRange>('3M')
  const [selectedAccount, setSelectedAccount] = useState<string>('all')

  const chartData = useMemo(() => {
    const from = getFromDate(timeRange)
    const filtered = snapshots.filter((s) => s.date >= from)

    const byDate = new Map<string, Record<string, number>>()
    for (const snap of filtered) {
      const entry = byDate.get(snap.date) ?? {}
      entry[snap.account_id] = snap.balance
      byDate.set(snap.date, entry)
    }

    return Array.from(byDate.entries())
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([date, balances]) => {
        const netWorth = accounts.reduce((sum, acc) => sum + (balances[acc.id] ?? 0), 0)
        return { date, ...balances, [NET_WORTH_KEY]: netWorth }
      })
  }, [snapshots, accounts, timeRange])

  const lineOpacity = (accountId: string) => {
    if (selectedAccount === 'all') return 1
    if (selectedAccount === accountId) return 1
    return 0.2
  }

  const lineWidth = (accountId: string) => {
    if (selectedAccount === 'all') return 1.5
    return selectedAccount === accountId ? 2.5 : 1.5
  }

  if (chartData.length === 0) {
    return (
      <Card>
        <CardContent className="pt-6">
          <EmptyState
            icon={TrendingUp}
            title="No history available"
            description="Balance history will appear here as your data grows."
          />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <CardTitle className="text-base">Balance History</CardTitle>
          <div className="flex flex-wrap items-center gap-2">
            <Select value={selectedAccount} onValueChange={setSelectedAccount}>
              <SelectTrigger className="h-8 w-40 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All accounts</SelectItem>
                {accounts.map((acc) => (
                  <SelectItem key={acc.id} value={acc.id}>
                    {acc.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <div className="flex gap-1">
              {TIME_RANGES.map(({ label, value }) => (
                <Button
                  key={value}
                  variant={timeRange === value ? 'outline' : 'ghost'}
                  size="sm"
                  className="h-8 px-2 text-xs"
                  onClick={() => setTimeRange(value)}
                >
                  {label}
                </Button>
              ))}
            </div>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
              <XAxis
                dataKey="date"
                tickFormatter={(v) => formatShortDate(v)}
                stroke="var(--muted-foreground)"
                fontSize={11}
                tickLine={false}
                axisLine={false}
                minTickGap={30}
              />
              <YAxis
                stroke="var(--muted-foreground)"
                fontSize={11}
                tickLine={false}
                axisLine={false}
                tickFormatter={(v) => {
                  const n = Number(v)
                  if (Math.abs(n) >= 1000) return `$${Math.round(n / 1000)}k`
                  return `$${Math.round(n)}`
                }}
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
                formatter={(value: number, name: string) => {
                  if (name === NET_WORTH_KEY) return [formatCurrency(value), 'Net Worth']
                  const acc = accounts.find((a) => a.id === name)
                  return [formatCurrency(value), acc?.name ?? name]
                }}
              />
              {accounts.map((acc, i) => (
                <Line
                  key={acc.id}
                  type="monotone"
                  dataKey={acc.id}
                  stroke={LINE_COLORS[i % LINE_COLORS.length]}
                  strokeWidth={lineWidth(acc.id)}
                  dot={false}
                  opacity={lineOpacity(acc.id)}
                  style={{ cursor: 'pointer' }}
                  onClick={() => setSelectedAccount(acc.id)}
                  activeDot={{ r: 4, style: { cursor: 'pointer' } }}
                />
              ))}
              <Line
                type="monotone"
                dataKey={NET_WORTH_KEY}
                stroke="var(--foreground)"
                strokeWidth={lineWidth('all')}
                strokeDasharray="6 2"
                dot={false}
                opacity={lineOpacity('all')}
                style={{ cursor: 'pointer' }}
                onClick={() => setSelectedAccount('all')}
                activeDot={{ r: 4, style: { cursor: 'pointer' } }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}
