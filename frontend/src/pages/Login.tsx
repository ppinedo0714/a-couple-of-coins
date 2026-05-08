import { Link } from 'react-router-dom'

export default function LoginPage() {
  return (
    <div className="mx-auto max-w-md px-6 py-16">
      <h1 className="text-2xl font-bold">Log in</h1>
      <p className="mt-2 text-sm text-[var(--muted-foreground)]">
        TODO: email + password form, OAuth buttons.
      </p>
      <p className="mt-8 text-sm">
        Don't have an account?{' '}
        <Link to="/register" className="underline">
          Sign up
        </Link>
      </p>
    </div>
  )
}
