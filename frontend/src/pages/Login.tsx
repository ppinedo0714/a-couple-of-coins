import { useEffect } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { ApiError } from '@/api/client'
import { useAuth } from '@/hooks/useAuth'
import { useAccounts } from '@/hooks/useAccounts'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { OAuthButton } from '@/components/shared/OAuthButton'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { PageWrapper } from '@/components/layout/PageWrapper'

const schema = z.object({
  email: z.string().email('Enter a valid email'),
  password: z.string().min(1, 'Password is required'),
})

type FormValues = z.infer<typeof schema>

export default function LoginPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { login, refetchMe, isAuthenticated } = useAuth()
  const accounts = useAccounts()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: '', password: '' },
  })

  const oauthSuccess = searchParams.get('oauth') === 'success'
  const next = searchParams.get('next')

  useEffect(() => {
    if (oauthSuccess) {
      void refetchMe()
    }
  }, [oauthSuccess, refetchMe])

  useEffect(() => {
    if (!isAuthenticated) return
    if (next) {
      navigate(next)
      return
    }
    if (!accounts.isLoading && (accounts.data?.length ?? 0) === 0) {
      navigate('/onboarding')
    } else {
      navigate('/dashboard')
    }
  }, [isAuthenticated, next, accounts.isLoading, accounts.data, navigate])

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      await login.mutateAsync(values)
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        form.setError('password', { message: 'Email or password is incorrect' })
      } else {
        form.setError('root', { message: err instanceof Error ? err.message : 'Login failed' })
      }
    }
  })

  return (
    <PageWrapper size="narrow">
      <Card>
        <CardHeader className="space-y-2 text-center">
          <CardTitle className="font-serif text-2xl">Welcome back</CardTitle>
          <CardDescription>Log in to your couple of coins.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-2">
            <OAuthButton provider="google" />
            <OAuthButton provider="github" />
          </div>
          <div className="flex items-center gap-3 text-xs uppercase tracking-wide text-muted-foreground">
            <span className="h-px flex-1 bg-border" />
            <span>or</span>
            <span className="h-px flex-1 bg-border" />
          </div>
          <form onSubmit={onSubmit} className="space-y-4" noValidate>
            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="you@example.com"
                {...form.register('email')}
              />
              {form.formState.errors.email ? (
                <p className="text-xs text-destructive">{form.formState.errors.email.message}</p>
              ) : null}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                {...form.register('password')}
              />
              {form.formState.errors.password ? (
                <p className="text-xs text-destructive">{form.formState.errors.password.message}</p>
              ) : null}
            </div>
            {form.formState.errors.root ? (
              <p className="text-sm text-destructive">{form.formState.errors.root.message}</p>
            ) : null}
            <Button type="submit" className="w-full" disabled={login.isPending}>
              {login.isPending ? <LoadingSpinner /> : null}
              Log in
            </Button>
          </form>
          <p className="text-center text-sm text-muted-foreground">
            Don&apos;t have an account?{' '}
            <Link to="/register" className="font-medium text-primary hover:underline">
              Sign up
            </Link>
          </p>
        </CardContent>
      </Card>
    </PageWrapper>
  )
}
