import { Link, NavLink } from 'react-router-dom'
import { cn } from '@/lib/utils'

export function Navbar() {
  return (
    <nav className="border-b border-[var(--border)]">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3">
        <Link to="/" className="font-semibold">
          a couple of coins
        </Link>
        <div className="flex items-center gap-4 text-sm">
          <NavLink
            to="/dashboard"
            className={({ isActive }) =>
              cn('hover:underline', isActive && 'font-medium')
            }
          >
            Dashboard
          </NavLink>
          <NavLink
            to="/import"
            className={({ isActive }) =>
              cn('hover:underline', isActive && 'font-medium')
            }
          >
            Import
          </NavLink>
          <NavLink
            to="/settings"
            className={({ isActive }) =>
              cn('hover:underline', isActive && 'font-medium')
            }
          >
            Settings
          </NavLink>
          <Link
            to="/login"
            className="rounded-md border border-[var(--border)] px-3 py-1 text-sm"
          >
            Log in
          </Link>
        </div>
      </div>
    </nav>
  )
}
