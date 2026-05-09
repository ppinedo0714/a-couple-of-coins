import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { ApiError } from '@/api/client'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ThemeToggle } from '@/components/shared/ThemeToggle'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'

const schema = z.object({
  email: z.string().email('Enter a valid email'),
})
type Values = z.infer<typeof schema>

export function ProfileTab() {
  const { user, updateMe, logout } = useAuth()
  const navigate = useNavigate()

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: { email: user?.email ?? '' },
    values: user ? { email: user.email } : undefined,
  })

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      await updateMe.mutateAsync({ email: values.email })
      toast.success('Email updated')
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        form.setError('email', { message: 'That email is already taken' })
      } else {
        toast.error(err instanceof Error ? err.message : 'Could not save')
      }
    }
  })

  const onLogout = () => {
    logout.mutate(undefined, {
      onSettled: () => navigate('/'),
    })
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Profile</CardTitle>
          <CardDescription>Update your sign-in email.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" {...form.register('email')} />
              {form.formState.errors.email ? (
                <p className="text-xs text-destructive">{form.formState.errors.email.message}</p>
              ) : null}
            </div>
            <Button type="submit" disabled={updateMe.isPending}>
              {updateMe.isPending ? <LoadingSpinner /> : null}
              Update email
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Appearance</CardTitle>
          <CardDescription>Choose how the app looks on this device.</CardDescription>
        </CardHeader>
        <CardContent>
          <ThemeToggle />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Session</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-wrap items-center gap-3">
          <Button variant="outline" onClick={onLogout}>
            Log out
          </Button>
          <Button variant="ghost" disabled className="text-destructive">
            Delete account · coming soon
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
