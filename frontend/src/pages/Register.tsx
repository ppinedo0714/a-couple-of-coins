import { Link } from 'react-router-dom'

export default function RegisterPage() {
  return (
    <div className="mx-auto max-w-md px-6 py-16">
      <h1 className="text-2xl font-bold">Create account</h1>
      <p className="mt-2 text-sm text-[var(--muted-foreground)]">
        TODO: email + password + confirm form, OAuth buttons.
      </p>
      <p className="mt-8 text-sm">
        Already have an account?{' '}
        <Link to="/login" className="underline">
          Log in
        </Link>
      </p>
    </div>
  )
}
