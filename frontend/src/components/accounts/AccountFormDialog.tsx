import { useEffect } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import type { Account, AccountType } from '@/types/models'
import { useCreateAccount, useUpdateAccount } from '@/hooks/useAccounts'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'

const schema = z.object({
  name: z.string().min(1, 'Required'),
  type: z.enum(['checking', 'savings', 'credit', 'investment']),
  balance: z.coerce.number(),
  currency: z.string().min(3).max(3).default('USD'),
})

type FormValues = z.infer<typeof schema>

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  account?: Account
}

const typeOptions: Array<{ value: AccountType; label: string }> = [
  { value: 'checking', label: 'Checking' },
  { value: 'savings', label: 'Savings' },
  { value: 'credit', label: 'Credit card' },
  { value: 'investment', label: 'Investment' },
]

export function AccountFormDialog({ open, onOpenChange, account }: Props) {
  const createMut = useCreateAccount()
  const updateMut = useUpdateAccount()
  const isEdit = !!account

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: account?.name ?? '',
      type: account?.type ?? 'checking',
      balance: account?.balance ?? 0,
      currency: account?.currency ?? 'USD',
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: account?.name ?? '',
        type: account?.type ?? 'checking',
        balance: account?.balance ?? 0,
        currency: account?.currency ?? 'USD',
      })
    }
  }, [open, account, form])

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      if (isEdit && account) {
        await updateMut.mutateAsync({
          id: account.id,
          body: { name: values.name, type: values.type },
        })
        toast.success('Account updated')
      } else {
        await createMut.mutateAsync(values)
        toast.success('Account created')
      }
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not save account')
    }
  })

  const pending = createMut.isPending || updateMut.isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit account' : 'Add account'}</DialogTitle>
          <DialogDescription>
            {isEdit ? 'Update the name or type for this account.' : 'Track a new account in your overview.'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="name">Name</Label>
            <Input id="name" placeholder="Chase Checking" {...form.register('name')} />
            {form.formState.errors.name ? (
              <p className="text-xs text-destructive">{form.formState.errors.name.message}</p>
            ) : null}
          </div>
          <div className="space-y-1.5">
            <Label>Type</Label>
            <Controller
              control={form.control}
              name="type"
              render={({ field }) => (
                <Select value={field.value} onValueChange={(v) => field.onChange(v as AccountType)}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a type" />
                  </SelectTrigger>
                  <SelectContent>
                    {typeOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
          </div>
          {!isEdit ? (
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="balance">Starting balance</Label>
                <Input id="balance" type="number" step="0.01" {...form.register('balance')} />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="currency">Currency</Label>
                <Input id="currency" maxLength={3} {...form.register('currency')} />
              </div>
            </div>
          ) : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              {pending ? <LoadingSpinner /> : null}
              {isEdit ? 'Save changes' : 'Create account'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
