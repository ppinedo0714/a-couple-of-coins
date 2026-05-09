import { useState } from 'react'
import { Plus } from 'lucide-react'
import type { Account } from '@/types/models'
import { AccountCard } from './AccountCard'
import { AccountFormDialog } from './AccountFormDialog'

type Props = {
  accounts: Account[]
}

export function AccountList({ accounts }: Props) {
  const [open, setOpen] = useState(false)
  return (
    <>
      <div className="flex gap-3 overflow-x-auto pb-2">
        {accounts.map((account) => (
          <AccountCard key={account.id} account={account} />
        ))}
        <button
          type="button"
          onClick={() => setOpen(true)}
          className="flex w-56 shrink-0 flex-col items-center justify-center gap-2 rounded-lg border border-dashed border-border bg-card/50 p-4 text-sm text-muted-foreground transition-colors hover:bg-muted/40 hover:text-foreground"
        >
          <Plus className="h-5 w-5" />
          <span>Add account</span>
        </button>
      </div>
      <AccountFormDialog open={open} onOpenChange={setOpen} />
    </>
  )
}
