import { useState } from 'react'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import type { Account } from '@/types/models'
import { ApiError } from '@/api/client'
import { useAccounts, useDeleteAccount } from '@/hooks/useAccounts'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { AccountFormDialog } from '@/components/accounts/AccountFormDialog'
import { LoadingScreen } from '@/components/shared/LoadingSpinner'
import { EmptyState } from '@/components/shared/EmptyState'
import { formatCurrency, formatDate } from '@/lib/format'

export function AccountsTab() {
  const accountsQuery = useAccounts()
  const deleteMut = useDeleteAccount()
  const [editing, setEditing] = useState<Account | undefined>(undefined)
  const [creating, setCreating] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState<Account | null>(null)

  if (accountsQuery.isLoading) return <LoadingScreen />

  const accounts = accountsQuery.data ?? []

  const onConfirmDelete = async () => {
    if (!confirmDelete) return
    try {
      await deleteMut.mutateAsync(confirmDelete.id)
      toast.success('Account deleted')
      setConfirmDelete(null)
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        toast.error('This account has transactions. Move or delete them first.')
      } else {
        toast.error(err instanceof Error ? err.message : 'Could not delete')
      }
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">Manage the accounts you track here.</p>
        <Button size="sm" onClick={() => setCreating(true)}>
          <Plus className="h-4 w-4" />
          Add account
        </Button>
      </div>

      {accounts.length === 0 ? (
        <EmptyState
          title="No accounts yet"
          description="Create one to start tracking spend."
          action={<Button onClick={() => setCreating(true)}>Add account</Button>}
        />
      ) : (
        <div className="rounded-lg border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead className="text-right">Balance</TableHead>
                <TableHead>Created</TableHead>
                <TableHead className="w-24" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {accounts.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="font-medium">{a.name}</TableCell>
                  <TableCell className="capitalize text-sm text-muted-foreground">{a.type}</TableCell>
                  <TableCell className="text-right font-mono tabular-nums">
                    {formatCurrency(a.balance, a.currency)}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDate(a.created_at)}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="icon" onClick={() => setEditing(a)}>
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setConfirmDelete(a)}
                      className="text-destructive hover:text-destructive"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <AccountFormDialog open={creating} onOpenChange={setCreating} />
      <AccountFormDialog
        open={!!editing}
        onOpenChange={(open) => {
          if (!open) setEditing(undefined)
        }}
        account={editing}
      />

      <Dialog open={!!confirmDelete} onOpenChange={(open) => !open && setConfirmDelete(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete this account?</DialogTitle>
            <DialogDescription>
              {confirmDelete ? `“${confirmDelete.name}” will be removed.` : ''} This cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmDelete(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={onConfirmDelete}
              disabled={deleteMut.isPending}
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
