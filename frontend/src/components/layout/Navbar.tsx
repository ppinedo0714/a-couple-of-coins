import { useState } from 'react'
import { Link, NavLink, useLocation } from 'react-router-dom'
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

function buildTransactionsLink(search: string): string {
  const params = new URLSearchParams(search)
  const forward = new URLSearchParams()
  for (const key of ['period', 'from', 'to']) {
    if (params.has(key)) forward.set(key, params.get(key)!)
  }
  const qs = forward.toString()
  return qs ? `/transactions?${qs}` : '/transactions'
}

export function Navbar() {
  const { user, isLoading } = useAuth()
  const { search } = useLocation()
  const [mobileOpen, setMobileOpen] = useState(false)
  const transactionsLink = buildTransactionsLink(search)

  return (
    <nav className="border-b border-border bg-background/80 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3">
        <Link to={user ? '/accounts' : '/'} className="font-serif text-lg tracking-tight text-foreground">
          a-couple-of-coins
        </Link>

        <div className="hidden items-center gap-6 md:flex">
          {user ? (
            <>
              <NavLink to="/accounts" className={navLinkClass}>
                Accounts
              </NavLink>
              <NavLink to="/transactions" className={navLinkClass}>
                Transactions
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
                <NavLink to="/accounts" className={navLinkClass} onClick={() => setMobileOpen(false)}>
                  Accounts
                </NavLink>
                <NavLink to={transactionsLink} className={navLinkClass} onClick={() => setMobileOpen(false)}>
                  Transactions
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
