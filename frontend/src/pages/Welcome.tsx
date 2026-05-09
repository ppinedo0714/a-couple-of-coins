import { Link, Navigate } from 'react-router-dom'
import { ArrowRight, BarChart3, Lock, Sparkles, Wallet } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { SpendingByCategory } from '@/components/charts/SpendingByCategory'
import { SpendingOverTime } from '@/components/charts/SpendingOverTime'
import { CategoryBreakdown } from '@/components/charts/CategoryBreakdown'
import { AccountCard } from '@/components/accounts/AccountCard'
import { demoAccounts, demoCategories, demoTransactions } from '@/fixtures/demo-data'
import { useAuth } from '@/hooks/useAuth'

const FEATURES = [
  {
    icon: Wallet,
    title: 'All your accounts in one place',
    body: 'Checking, savings, credit cards, investments — see balances and activity at a glance.',
  },
  {
    icon: BarChart3,
    title: 'Charts that actually help',
    body: 'Spending by category and cash flow over time, with filters that sync with your URL.',
  },
  {
    icon: Sparkles,
    title: 'Smart categorization',
    body: 'Drag-drop a CSV and let predictions suggest categories you can accept or override.',
  },
  {
    icon: Lock,
    title: 'Built for privacy',
    body: 'Your data never leaves the API. Auth lives in httpOnly cookies — never in the browser.',
  },
]

export default function WelcomePage() {
  const { isAuthenticated, isLoading } = useAuth()
  if (isLoading) {
    return null
  }
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }
  return (
    <div>
      <section className="mx-auto max-w-5xl px-6 pt-16 pb-10 text-center">
        <p className="mb-4 inline-flex items-center gap-2 rounded-full border border-border bg-card/60 px-3 py-1 text-xs font-medium text-muted-foreground">
          <Sparkles className="h-3 w-3 text-primary" /> Predictive money tracking
        </p>
        <h1 className="font-serif text-5xl tracking-tight text-foreground sm:text-6xl">
          A couple of coins,
          <br />
          looked after.
        </h1>
        <p className="mx-auto mt-6 max-w-2xl text-base text-muted-foreground">
          Track every account, categorize transactions automatically, and see what your spending
          will look like next month — not just last.
        </p>
        <div className="mt-8 flex flex-wrap justify-center gap-3">
          <Button asChild size="lg">
            <Link to="/register">
              Get started
              <ArrowRight className="h-4 w-4" />
            </Link>
          </Button>
          <Button asChild size="lg" variant="outline">
            <Link to="/login">Log in</Link>
          </Button>
        </div>
      </section>

      <section className="mx-auto max-w-6xl px-6 pb-16">
        <div className="rounded-xl border border-border bg-card/40 p-4 shadow-sm sm:p-6">
          <div className="mb-2 text-xs uppercase tracking-wide text-muted-foreground">
            Live preview · sample data
          </div>
          <div className="pointer-events-none space-y-6">
            <div className="flex gap-3 overflow-x-auto pb-1">
              {demoAccounts.map((a) => (
                <AccountCard key={a.id} account={a} />
              ))}
            </div>
            <div className="grid gap-6 lg:grid-cols-2">
              <SpendingByCategory transactions={demoTransactions} categories={demoCategories} />
              <SpendingOverTime transactions={demoTransactions} />
            </div>
            <CategoryBreakdown transactions={demoTransactions} categories={demoCategories} />
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-6xl px-6 pb-20">
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {FEATURES.map(({ icon: Icon, title, body }) => (
            <Card key={title} className="bg-card/60">
              <CardHeader>
                <div className="flex h-10 w-10 items-center justify-center rounded-md bg-primary/10 text-primary">
                  <Icon className="h-4 w-4" />
                </div>
                <CardTitle className="mt-2 text-base">{title}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">{body}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      <footer className="border-t border-border bg-card/30 py-8">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-3 px-6 text-sm text-muted-foreground">
          <span className="font-serif text-foreground">a-couple-of-coins</span>
          <div className="flex gap-4">
            <Link to="/login" className="hover:text-foreground">
              Log in
            </Link>
            <Link to="/register" className="hover:text-foreground">
              Sign up
            </Link>
          </div>
        </div>
      </footer>
    </div>
  )
}
