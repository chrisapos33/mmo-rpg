import { type InputHTMLAttributes } from 'react'

interface Props extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export function Input({ label, error, className = '', ...props }: Props) {
  return (
    <div className="flex flex-col gap-1.5">
      {label && (
        <label className="text-xs font-medium tracking-widest uppercase text-ink-400">
          {label}
        </label>
      )}
      <input
        {...props}
        className={`
          w-full bg-void-800 border px-4 py-2.5 text-sm text-ink-50
          placeholder:text-ink-600 outline-none transition-colors
          ${error ? 'border-danger-400' : 'border-void-600 focus:border-gold-500'}
          ${className}
        `}
      />
      {error && <p className="text-xs text-danger-400">{error}</p>}
    </div>
  )
}
