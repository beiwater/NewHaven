import { Toaster as SonnerToaster } from 'sonner'
import { cn } from '@/lib/utils'

type ToasterProps = React.ComponentProps<typeof SonnerToaster>

function Toaster({ className, ...props }: ToasterProps) {
  return (
    <SonnerToaster
      className={cn('toaster group', className)}
      toastOptions={{
        classNames: {
          toast: 'group toast group-[.toaster]:bg-white group-[.toaster]:text-amber-900 group-[.toaster]:border-amber-200 group-[.toaster]:shadow-lg',
          description: 'group-[.toast]:text-amber-600',
          actionButton: 'group-[.toast]:bg-amber-600 group-[.toast]:text-white',
          cancelButton: 'group-[.toast]:bg-amber-100 group-[.toast]:text-amber-900',
        },
      }}
      {...props}
    />
  )
}

export { Toaster }
