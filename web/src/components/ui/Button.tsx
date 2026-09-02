import type { ButtonHTMLAttributes } from 'react'

// Variants map to DESIGN_SYSTEM.md's fixed accent meanings: yellow = primary
// action, cyan = secondary/info, pink = live/urgent. Don't add a variant that
// reassigns an accent to a new meaning.
type Variant = 'primary' | 'secondary' | 'ghost' | 'urgent'

const variantClasses: Record<Variant, string> = {
  primary:
    'bg-accent-yellow text-page shadow-lg shadow-accent-yellow/20 hover:brightness-110 active:brightness-95',
  secondary:
    'border border-accent-cyan text-accent-cyan hover:bg-accent-cyan-muted active:brightness-110',
  ghost:
    'border border-border text-text-muted hover:border-text-muted hover:text-text-primary',
  urgent: 'bg-accent-pink text-page hover:brightness-110 active:brightness-95',
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
}

export function Button({ variant = 'primary', className = '', ...props }: ButtonProps) {
  return (
    <button
      {...props}
      className={`rounded-pill px-6 py-3 font-extrabold transition-[filter,background-color,border-color,color] duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent-cyan disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:brightness-100 ${variantClasses[variant]} ${className}`}
    />
  )
}
