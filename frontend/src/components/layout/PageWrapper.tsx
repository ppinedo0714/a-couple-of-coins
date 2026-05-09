import { cn } from '@/lib/utils'

type PageWrapperProps = React.HTMLAttributes<HTMLDivElement> & {
  size?: 'narrow' | 'default'
}

export function PageWrapper({ className, size = 'default', children, ...props }: PageWrapperProps) {
  return (
    <div
      className={cn(
        'mx-auto px-6 py-8',
        size === 'narrow' ? 'max-w-md' : 'max-w-6xl',
        className,
      )}
      {...props}
    >
      {children}
    </div>
  )
}
