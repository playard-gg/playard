import type { ReactNode } from 'react'
import * as RadixDialog from '@radix-ui/react-dialog'

interface DialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: string
  children: ReactNode
}

/** Radix dialog primitive dressed in our tokens — no component kit involved. */
export function Dialog({ open, onOpenChange, title, description, children }: DialogProps) {
  return (
    <RadixDialog.Root open={open} onOpenChange={onOpenChange}>
      <RadixDialog.Portal>
        <RadixDialog.Overlay className="fixed inset-0 z-40 bg-page/80 backdrop-blur-sm" />
        <RadixDialog.Content className="fixed top-1/2 left-1/2 z-50 w-[calc(100vw-3rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-lg border border-border bg-card p-7 shadow-2xl focus:outline-none">
          <RadixDialog.Title className="text-xl font-bold text-text-primary">{title}</RadixDialog.Title>
          {description && (
            <RadixDialog.Description className="mt-1 text-sm text-text-muted">
              {description}
            </RadixDialog.Description>
          )}
          <div className="mt-6">{children}</div>
        </RadixDialog.Content>
      </RadixDialog.Portal>
    </RadixDialog.Root>
  )
}
