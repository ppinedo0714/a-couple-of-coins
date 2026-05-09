import { Github } from 'lucide-react'
import { cn } from '@/lib/utils'

type Provider = 'google' | 'github'

const labels: Record<Provider, string> = {
  google: 'Continue with Google',
  github: 'Continue with GitHub',
}

function GoogleMark({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <path
        fill="#FFC107"
        d="M21.35 11.1H12v3.8h5.35c-.5 2.4-2.55 3.8-5.35 3.8a6.7 6.7 0 0 1 0-13.4c1.65 0 3.15.6 4.3 1.7l2.85-2.85A10.5 10.5 0 0 0 12 1a11 11 0 0 0 0 22c5.95 0 10.95-4.3 10.95-11 0-.7-.1-1.3-.2-1.9z"
      />
      <path fill="#FF3D00" d="M3.15 7.35l3.1 2.3A6.7 6.7 0 0 1 12 6.3c1.65 0 3.15.6 4.3 1.7L19.15 5A10.5 10.5 0 0 0 12 1.5 10.5 10.5 0 0 0 3.15 7.35z" />
      <path fill="#4CAF50" d="M12 22.5a10.5 10.5 0 0 0 7.05-2.7l-3.25-2.7c-1.05.7-2.4 1.1-3.8 1.1-2.8 0-5.05-1.85-5.85-4.4l-3.2 2.5A10.5 10.5 0 0 0 12 22.5z" />
      <path fill="#1976D2" d="M21.35 11.1H12v3.8h5.35c-.25 1.2-1 2.25-2.05 2.95l3.25 2.7c1.85-1.7 3-4.2 3-7.55 0-.7-.1-1.3-.2-1.9z" />
    </svg>
  )
}

export function OAuthButton({ provider, className }: { provider: Provider; className?: string }) {
  return (
    <a
      href={`/api/v1/auth/oauth/${provider}`}
      className={cn(
        'inline-flex h-10 w-full items-center justify-center gap-2 rounded-md border border-border bg-card px-4 text-sm font-medium text-foreground transition-colors hover:bg-muted',
        className,
      )}
    >
      {provider === 'google' ? (
        <GoogleMark className="h-4 w-4" />
      ) : (
        <Github className="h-4 w-4" />
      )}
      <span>{labels[provider]}</span>
    </a>
  )
}
