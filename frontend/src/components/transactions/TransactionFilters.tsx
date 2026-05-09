import { useEffect, useState } from 'react'
import { Search, X } from 'lucide-react'
import type { Account, Category } from '@/types/models'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'

export type TransactionFilterValue = {
  accountId: string | null
  categoryId: string | null
  search: string
}

type Props = {
  accounts: Account[]
  categories: Category[]
  value: TransactionFilterValue
  onChange: (next: TransactionFilterValue) => void
  className?: string
}

export function TransactionFilters({ accounts, categories, value, onChange, className }: Props) {
  const [search, setSearch] = useState(value.search)

  useEffect(() => {
    if (search === value.search) return
    const handle = setTimeout(() => onChange({ ...value, search }), 250)
    return () => clearTimeout(handle)
  }, [search, value, onChange])

  const hasFilters =
    value.accountId !== null || value.categoryId !== null || value.search.length > 0

  return (
    <div className={cn('flex flex-wrap items-center gap-2', className)}>
      <div className="relative min-w-48 flex-1">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search transactions"
          className="pl-9"
        />
      </div>
      <Select
        value={value.accountId ?? 'all'}
        onValueChange={(v) => onChange({ ...value, accountId: v === 'all' ? null : v })}
      >
        <SelectTrigger className="w-44">
          <SelectValue placeholder="Account" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All accounts</SelectItem>
          {accounts.map((a) => (
            <SelectItem key={a.id} value={a.id}>
              {a.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select
        value={value.categoryId ?? 'all'}
        onValueChange={(v) => onChange({ ...value, categoryId: v === 'all' ? null : v })}
      >
        <SelectTrigger className="w-44">
          <SelectValue placeholder="Category" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All categories</SelectItem>
          {categories.map((c) => (
            <SelectItem key={c.id} value={c.id}>
              {c.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {hasFilters ? (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => {
            setSearch('')
            onChange({ accountId: null, categoryId: null, search: '' })
          }}
        >
          <X className="h-3.5 w-3.5" />
          Clear
        </Button>
      ) : null}
    </div>
  )
}
