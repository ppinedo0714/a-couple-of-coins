import { Link } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { PageWrapper } from '@/components/layout/PageWrapper'

export default function NotFoundPage() {
  const { isAuthenticated } = useAuth()
  const home = isAuthenticated ? '/dashboard' : '/'
  return (
    <PageWrapper size="narrow" className="py-24 text-center">
      <h1 className="font-serif text-6xl text-primary">404</h1>
      <p className="mt-4 text-lg text-foreground">This page doesn&apos;t exist.</p>
      <p className="mt-2 text-sm text-muted-foreground">
        The link may be broken, or the page may have moved.
      </p>
      <Button asChild className="mt-8">
        <Link to={home}>Go home</Link>
      </Button>
    </PageWrapper>
  )
}
