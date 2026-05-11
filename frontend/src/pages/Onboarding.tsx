import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Controller, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Check, Upload } from 'lucide-react'
import type { AccountType } from '@/types/models'
import { useCreateAccount } from '@/hooks/useAccounts'
import { useCreateCategory } from '@/hooks/useCategories'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { PageWrapper } from '@/components/layout/PageWrapper'
import { cn } from '@/lib/utils'

const STEPS = ['Account', 'Categories', 'Import'] as const

const accountSchema = z.object({
  name: z.string().min(1, 'Required'),
  type: z.enum(['checking', 'savings', 'credit', 'investment']),
  balance: z.coerce.number(),
  currency: z.string().min(3).max(3),
})
type AccountValues = z.infer<typeof accountSchema>

const SUGGESTED = ['Groceries', 'Rent', 'Transport', 'Salary', 'Entertainment', 'Dining', 'Utilities']

export default function OnboardingPage() {
  const [step, setStep] = useState(0)
  const navigate = useNavigate()

  const createAccount = useCreateAccount()
  const createCategory = useCreateCategory()

  const accountForm = useForm<AccountValues>({
    resolver: zodResolver(accountSchema),
    defaultValues: { name: '', type: 'checking', balance: 0, currency: 'USD' },
  })

  const [selected, setSelected] = useState<string[]>(SUGGESTED.slice(0, 5))
  const [customCategory, setCustomCategory] = useState('')

  const submitAccount = accountForm.handleSubmit(async (values) => {
    try {
      await createAccount.mutateAsync(values)
      setStep(1)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not create account')
    }
  })

  const toggleCategory = (name: string) => {
    setSelected((prev) =>
      prev.includes(name) ? prev.filter((n) => n !== name) : [...prev, name],
    )
  }

  const submitCategories = async () => {
    const all = [...new Set([...selected, ...(customCategory ? [customCategory.trim()] : [])])].filter(
      Boolean,
    )
    try {
      await Promise.all(all.map((name) => createCategory.mutateAsync({ name })))
      setStep(2)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not save categories')
    }
  }

  const finish = () => navigate('/accounts')

  return (
    <PageWrapper size="narrow" className="max-w-xl">
      <div className="mb-6 flex items-center justify-between">
        {STEPS.map((label, i) => {
          const active = i === step
          const done = i < step
          return (
            <div key={label} className="flex flex-1 items-center gap-2">
              <div
                className={cn(
                  'flex h-8 w-8 items-center justify-center rounded-full text-xs font-medium',
                  active && 'bg-primary text-primary-foreground',
                  done && 'bg-primary/20 text-primary',
                  !active && !done && 'bg-muted text-muted-foreground',
                )}
              >
                {done ? <Check className="h-4 w-4" /> : i + 1}
              </div>
              <span className={cn('text-sm', active ? 'text-foreground' : 'text-muted-foreground')}>
                {label}
              </span>
              {i < STEPS.length - 1 ? <div className="h-px flex-1 bg-border" /> : null}
            </div>
          )
        })}
      </div>

      {step === 0 ? (
        <Card>
          <CardHeader>
            <CardTitle>Add your first account</CardTitle>
            <CardDescription>
              You can add more later. This is just to get the dashboard going.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={submitAccount} className="space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="name">Account name</Label>
                <Input id="name" placeholder="Chase Checking" {...accountForm.register('name')} />
                {accountForm.formState.errors.name ? (
                  <p className="text-xs text-destructive">{accountForm.formState.errors.name.message}</p>
                ) : null}
              </div>
              <div className="space-y-1.5">
                <Label>Type</Label>
                <Controller
                  control={accountForm.control}
                  name="type"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={(v) => field.onChange(v as AccountType)}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="checking">Checking</SelectItem>
                        <SelectItem value="savings">Savings</SelectItem>
                        <SelectItem value="credit">Credit card</SelectItem>
                        <SelectItem value="investment">Investment</SelectItem>
                      </SelectContent>
                    </Select>
                  )}
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label htmlFor="balance">Starting balance</Label>
                  <Input id="balance" type="number" step="0.01" {...accountForm.register('balance')} />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="currency">Currency</Label>
                  <Input id="currency" maxLength={3} {...accountForm.register('currency')} />
                </div>
              </div>
              <div className="flex justify-between pt-2">
                <Button type="button" variant="ghost" onClick={() => setStep(1)}>
                  Skip
                </Button>
                <Button type="submit" disabled={createAccount.isPending}>
                  {createAccount.isPending ? <LoadingSpinner /> : null}
                  Continue
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      ) : null}

      {step === 1 ? (
        <Card>
          <CardHeader>
            <CardTitle>Pick a few categories</CardTitle>
            <CardDescription>You can edit, add, or remove these later.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap gap-2">
              {SUGGESTED.map((name) => {
                const active = selected.includes(name)
                return (
                  <button
                    key={name}
                    type="button"
                    onClick={() => toggleCategory(name)}
                    className={cn(
                      'rounded-full border px-3 py-1.5 text-sm transition-colors',
                      active
                        ? 'border-primary bg-primary/10 text-primary'
                        : 'border-border bg-card text-muted-foreground hover:text-foreground',
                    )}
                  >
                    {name}
                  </button>
                )
              })}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="custom">Add your own</Label>
              <Input
                id="custom"
                placeholder="e.g. Pet supplies"
                value={customCategory}
                onChange={(e) => setCustomCategory(e.target.value)}
              />
            </div>
            <div className="flex justify-between pt-2">
              <Button type="button" variant="ghost" onClick={() => setStep(2)}>
                Skip
              </Button>
              <Button type="button" onClick={submitCategories} disabled={createCategory.isPending}>
                {createCategory.isPending ? <LoadingSpinner /> : null}
                Continue
              </Button>
            </div>
          </CardContent>
        </Card>
      ) : null}

      {step === 2 ? (
        <Card>
          <CardHeader>
            <CardTitle>Import some history</CardTitle>
            <CardDescription>
              Drop a CSV from your bank to seed your dashboard. You can also do this later.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <Button type="button" className="w-full" onClick={() => navigate('/import')}>
              <Upload className="h-4 w-4" />
              Upload a CSV
            </Button>
            <Button type="button" variant="outline" className="w-full" onClick={finish}>
              I&apos;ll do this later
            </Button>
          </CardContent>
        </Card>
      ) : null}
    </PageWrapper>
  )
}
