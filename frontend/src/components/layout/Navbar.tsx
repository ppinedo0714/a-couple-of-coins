import { useState } from 'react'
import { Link, NavLink } from 'react-router-dom'
import { Menu, X } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { ProfileMenu } from './ProfileMenu'
import { cn } from '@/lib/utils'

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    'text-sm text-muted-foreground transition-colors hover:text-foreground',
    isActive && 'font-medium text-primary',
  )

export function Navbar() {
  const { user, isLoading } = useAuth()
  const [mobileOpen, setMobileOpen] = useState(false)

  return (
    <nav className="border-b border-border bg-background/80 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3">
        <Link to={user ? '/dashboard' : '/'} className="font-serif text-lg tracking-tight text-foreground">
          a-couple-of-coins
        </Link>

        <div className="hidden items-center gap-6 md:flex">
          {user ? (
            <>
              <NavLink to="/dashboard" className={navLinkClass}>
                Dashboard
              </NavLink>
              <NavLink to="/import" className={navLinkClass}>
                Import
              </NavLink>
              <ProfileMenu />
            </>
          ) : isLoading ? null : (
            <>
              <NavLink to="/login" className={navLinkClass}>
                Log in
              </NavLink>
              <Button asChild size="sm">
                <Link to="/register">Get started</Link>
              </Button>
            </>
          )}
        </div>

        <button
          type="button"
          className="rounded-md p-2 text-foreground md:hidden"
          onClick={() => setMobileOpen((v) => !v)}
          aria-label="Toggle navigation"
        >
          {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {mobileOpen ? (
        <div className="border-t border-border md:hidden">
          <div className="mx-auto flex max-w-6xl flex-col gap-3 px-6 py-4">
            {user ? (
              <>
                <NavLink to="/dashboard" className={navLinkClass} onClick={() => setMobileOpen(false)}>
                  Dashboard
                </NavLink>
                <NavLink to="/import" className={navLinkClass} onClick={() => setMobileOpen(false)}>
                  Import
                </NavLink>
                <NavLink to="/settings" className={navLinkClass} onClick={() => setMobileOpen(false)}>
                  Settings
                </NavLink>
                <div className="pt-2">
                  <ProfileMenu />
                </div>
              </>
            ) : (
              <>
                <NavLink to="/login" className={navLinkClass} onClick={() => setMobileOpen(false)}>
                  Log in
                </NavLink>
                <Button asChild size="sm" className="w-fit">
                  <Link to="/register" onClick={() => setMobileOpen(false)}>
                    Get started
                  </Link>
                </Button>
              </>
            )}
          </div>
        </div>
      ) : null}
    </nav>
  )
}
