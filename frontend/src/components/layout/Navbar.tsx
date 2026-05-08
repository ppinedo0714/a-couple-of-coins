import { Link, NavLink } from 'react-router-dom'
import { cn } from '@/lib/utils'

export function Navbar() {
  return (
    <nav className="border-b border-border">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3">
        <Link
          to="/"
          className="font-serif text-lg tracking-tight text-foreground"
        >
          a-couple-of-coins
        </Link>
        <div className="flex items-center gap-4 text-sm">
          <NavLink
            to="/dashboard"
            className={({ isActive }) =>
              cn(
                'text-muted-foreground transition-colors hover:text-foreground',
                isActive && 'font-medium text-primary',
              )
            }
          >
            Dashboard
          </NavLink>
          <NavLink
            to="/import"
            className={({ isActive }) =>
              cn(
                'text-muted-foreground transition-colors hover:text-foreground',
                isActive && 'font-medium text-primary',
              )
            }
          >
            Import
          </NavLink>
          <NavLink
            to="/settings"
            className={({ isActive }) =>
              cn(
                'text-muted-foreground transition-colors hover:text-foreground',
                isActive && 'font-medium text-primary',
              )
            }
          >
            Settings
          </NavLink>
          <Link
            to="/login"
            className="rounded-md border border-border bg-card px-3 py-1 text-sm transition-colors hover:bg-muted"
          >
            Log in
          </Link>
        </div>
      </div>
    </nav>
  )
}
