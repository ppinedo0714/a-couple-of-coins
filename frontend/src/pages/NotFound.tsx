import { Link } from 'react-router-dom'

export default function NotFoundPage() {
  return (
    <div className="mx-auto max-w-md px-6 py-24 text-center">
      <h1 className="text-4xl font-bold">404</h1>
      <p className="mt-2 text-[var(--muted-foreground)]">Page not found.</p>
      <Link to="/" className="mt-6 inline-block underline">
        Go home
      </Link>
    </div>
  )
}
