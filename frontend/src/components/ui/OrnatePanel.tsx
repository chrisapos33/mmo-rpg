import type { ReactNode } from 'react'

interface OrnatePanelProps {
  children: ReactNode
  className?: string
  title?: string
}

const CornerDiamond = () => (
  <svg width="10" height="10" viewBox="0 0 10 10" fill="none" aria-hidden="true">
    <path d="M5 0L10 5L5 10L0 5Z" fill="rgba(200,150,50,0.8)" />
  </svg>
)

export function OrnatePanel({ children, className = '', title }: OrnatePanelProps) {
  return (
    <div className={`relative ${className}`} style={{ overflow: 'visible' }}>
      {/* Corner diamonds */}
      <span className="absolute -top-1.5 -left-1.5 z-10 pointer-events-none">
        <CornerDiamond />
      </span>
      <span className="absolute -top-1.5 -right-1.5 z-10 pointer-events-none">
        <CornerDiamond />
      </span>
      <span className="absolute -bottom-1.5 -left-1.5 z-10 pointer-events-none">
        <CornerDiamond />
      </span>
      <span className="absolute -bottom-1.5 -right-1.5 z-10 pointer-events-none">
        <CornerDiamond />
      </span>

      {/* Panel body */}
      <div
        className="relative bg-void-900 px-6 py-5"
        style={{ border: '1px solid rgba(200,150,50,0.3)' }}
      >
        {title && (
          <p
            className="text-xs tracking-[0.3em] uppercase mb-4"
            style={{ fontFamily: '"Cinzel", serif', color: '#c99224', fontVariant: 'small-caps' }}
          >
            {title}
          </p>
        )}
        {children}
      </div>
    </div>
  )
}
