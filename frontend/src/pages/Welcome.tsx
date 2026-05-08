import { Link } from 'react-router-dom'

export default function WelcomePage() {
  return (
    <div className="mx-auto max-w-3xl px-6 py-16">
      <h1 className="text-4xl font-bold tracking-tight">a couple of coins</h1>
      <p className="mt-4 text-lg text-[var(--muted-foreground)]">
        Track your money. Understand your spending. Predict what's coming.
      </p>
      <div className="mt-8 flex gap-3">
        <Link
          to="/register"
          className="rounded-md bg-[var(--primary)] px-4 py-2 text-sm font-medium text-[var(--primary-foreground)]"
        >
          Get started
        </Link>
        <Link
          to="/login"
          className="rounded-md border border-[var(--border)] px-4 py-2 text-sm font-medium"
        >
          Log in
        </Link>
      </div>
      <p className="mt-12 text-sm text-[var(--muted-foreground)]">
        TODO: interactive demo preview, feature highlights, footer.
      </p>
    </div>
  )
}
