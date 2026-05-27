import { type ButtonHTMLAttributes } from 'react'

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'ghost'
  loading?: boolean
}

export function Button({ variant = 'primary', loading, children, className = '', disabled, ...props }: Props) {
  const base = 'inline-flex items-center justify-center gap-2 px-6 py-2.5 text-sm font-medium tracking-wide transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed'
  const variants = {
    primary: 'bg-gold-400 text-void-950 hover:bg-gold-300 active:bg-gold-500',
    ghost:   'border border-void-600 text-ink-200 hover:border-gold-400 hover:text-gold-300',
  }

  return (
    <button
      {...props}
      disabled={disabled || loading}
      className={`${base} ${variants[variant]} ${className}`}
    >
      {loading && (
        <span className="size-4 rounded-full border-2 border-current border-t-transparent animate-spin" />
      )}
      {children}
    </button>
  )
}
